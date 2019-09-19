package main

import (
	"context"
	"errors"
	"fmt"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/agent"
	"github.com/mesos/mesos-go/api/v1/lib/agent/calls"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpagent"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	maxLogSize = 2000
)

// Listener streams the content of a file
type Listener struct {
	agentSender   calls.Sender
	task          mesos.Task
	agent         mesos.AgentInfo
	fileName      string
	logIdentifier string
	filterString  string
}

func NewListener(fileName string, task mesos.Task, agentInfo mesos.AgentInfo) (*Listener, error) {
	if task.AgentID.Value != agentInfo.ID.Value {
		return nil, errors.New("tasks agent id doesn't match provided agent info")
	}

	// TODO https?
	agentUrl := fmt.Sprintf("http://%s/api/v1", net.JoinHostPort(agentInfo.GetHostname(), strconv.Itoa(int(agentInfo.GetPort()))))
	agentSender := httpagent.NewSender(httpcli.New(httpcli.Endpoint(agentUrl)).Send)
	return &Listener{
		agentSender: agentSender,
		task:        task,
		fileName:    fileName,
		agent:       agentInfo,
	}, nil
}

// Listen starts listening to the specified file and streams out the content
func (l *Listener) Listen(output chan string, commandStream chan Command, done chan struct{}) {
	// Get container info
	containers, err := l.getContainers() // r.GetGetContainers().getContainers()
	if err != nil {
		log.Println("error while getting containers: ", err.Error())
		return
	}

	// Match the correct container
	containerId := ""
	for _, c := range containers {
		if c.GetExecutorID().Value == l.task.GetTaskID().Value { // TODO assuming this is ok. Doublecheck
			containerId = c.GetContainerID().Value
			break
		}
	}

	if containerId == "" {
		log.Println("container not found")
		return
	}

	// Get flags
	flags, err := l.getFlags()
	agentWorkDir := ""
	for _, f := range flags {
		if f.GetName() == "work_dir" {
			agentWorkDir = f.GetValue()
		}
	}

	agentId := l.task.GetAgentID().Value
	frameworkId := l.task.GetFrameworkID().Value
	taskId := l.task.GetTaskID().Value

	// {workdir}/slaves/{agentId}/frameworks/{frameworkId}/executors/{taskId}/runs/{containerId}/stdout
	fullPath := fmt.Sprintf("%s/slaves/%s/frameworks/%s/executors/%s/runs/%s/%s", agentWorkDir, agentId, frameworkId, taskId, containerId, l.fileName)
	offset := uint64(0)
	initial := true
	var resp mesos.Response
	timer := time.After(time.Duration(1000) * time.Millisecond)
	stopReqested := false
	// TODO configurable log identifiers
	l.logIdentifier = fmt.Sprintf("%s:%d", l.agent.Hostname, l.task.GetDiscovery().GetPorts().Ports[0].Number) //  l.task.GetTaskID().Value

	// listen loop
	for {
		// TODO what about: http://mesos.apache.org/documentation/latest/operator-http-api/#attach_container_output
		if initial {
			resp, err = l.agentSender.Send(context.TODO(), calls.NonStreaming(calls.ReadFileWithLength(fullPath, offset, 0))) // only to get the current size
		} else {
			resp, err = l.agentSender.Send(context.TODO(), calls.NonStreaming(calls.ReadFile(fullPath, offset))) // read to the end of the file
		}

		if err != nil {
			log.Println("error while reading file", err.Error())
			return
		}

		var agentResponse agent.Response
		err = resp.Decode(&agentResponse)
		if err != nil {
			log.Println("error while decoding read file response", err.Error())
			return
		}

		r := agentResponse.GetReadFile()

		// initial call to get size
		if initial {
			if r.GetSize() > maxLogSize {
				offset = r.GetSize() - maxLogSize
			}
			initial = false
			continue
		} else {
			offset = r.GetSize()
		}

		data := r.GetData()
		if len(data) != 0 {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if len(strings.TrimSpace(line)) > 0 {
					// TODO use templates
					if l.filterString != "" && !strings.Contains(line, l.filterString) {
						continue
					}
					// TODO implement grep like filter. Use a channel to push the filter string to all listeners
					output <- fmt.Sprintf("[%s]: %s", l.logIdentifier, line)
				}
			}
		}

		select {
			case <-timer:
				timer = time.After(time.Duration(1000) * time.Millisecond)
				continue
			case _, ok := <-done:
				if !ok {
					stopReqested = true
					break
				}
			case cmd := <- commandStream:
				l.handleCommand(cmd)
		}
		if stopReqested {
			log.Println("stop listening to ", l.logIdentifier)
			return
		}
	}
}

func (l *Listener) handleCommand(cmd Command) {
	// TODO type switch better? and then we can get typed parameters?
	if cmd.Name() == "test" {
		log.Printf("%s command in listener %s!\n", cmd.Name(), l.logIdentifier)
	} else if cmd.Name() == FilterCommandName {
		l.filterString =  cmd.Parameters()[0]
	}
}

func (l *Listener) getContainers() (containers []agent.Response_GetContainers_Container, err error) {
	resp, err := l.agentSender.Send(context.TODO(), calls.NonStreaming(calls.GetContainers()))
	if err != nil {
		return containers, err
	}

	var r agent.Response
	err = resp.Decode(&r)
	if err != nil {
		return containers, err
	}

	return r.GetGetContainers().GetContainers(), nil
}

func (l *Listener) getFlags() (flags []mesos.Flag, err error) {
	resp, err := l.agentSender.Send(context.TODO(), calls.NonStreaming(calls.GetFlags()))
	if err != nil {
		return flags, err
	}
	var r agent.Response
	err = resp.Decode(&r)
	if err != nil {
		return flags, err
	}

	return r.GetGetFlags().GetFlags(), nil
}
