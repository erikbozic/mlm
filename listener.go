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
	"mlm/commands"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	maxLogSize   = 2000
	pollInterval = 500
)

// Listener streams the content of a file
type Listener struct {
	agentSender   calls.Sender
	task          mesos.Task
	agent         mesos.AgentInfo
	fileName      string
	logIdentifier string
	filterString  string
	timer         <-chan time.Time
	color         string
}

func NewListener(fileName string, task mesos.Task, agentInfo mesos.AgentInfo, color string) (*Listener, error) {
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
		color:       color,
	}, nil
}

// Listen starts listening to the specified file and streams out the content
func (l *Listener) Listen(output chan<- string, commandStream <-chan commands.Command, done <-chan struct{}) {
	// Get container info
	containers, err := l.getContainers()
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
	l.timer = time.NewTimer(0).C // right away
	// TODO configurable log identifiers
	l.logIdentifier = fmt.Sprintf("%s:%d", l.agent.Hostname, l.task.GetDiscovery().GetPorts().Ports[0].Number)

	poll := func() {
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
			return
		} else {
			offset = r.GetSize()
		}

		data := r.GetData()
		if len(data) != 0 {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if len(strings.TrimSpace(line)) > 0 {
					// TODO use templates
					logMessage := fmt.Sprintf("%s[%s]%s: %s", l.color, l.logIdentifier, "\u001b[0m", line)
					if l.filterString != "" && !strings.Contains(logMessage, l.filterString) {
						continue
					}
					output <- logMessage
				}
			}
		}
	}

	// listen loop
	for {
		select {
		case <-l.timer:
			l.timer = time.After(time.Duration(pollInterval) * time.Millisecond)
			poll()
		case _, ok := <-done:
			if !ok {
				log.Println("stop listening to ", l.logIdentifier)
				return
			}
		case cmd := <-commandStream:
			l.handleCommand(cmd)
		}
	}
}

func (l *Listener) handleCommand(cmd commands.Command) {
	// TODO type switch better? and then we can get typed parameters?
	if cmd.Name() == commands.FilterCommandName {
		l.filterString = cmd.Parameters()[0]
	} else if cmd.Name() == commands.PauseCommandName {
		l.timer = nil
	} else if cmd.Name() == commands.UnpauseCommandName {
		l.timer = time.NewTimer(0).C // right away
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
