package main

import (
	"context"
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

// Listener streams the content of a file
type Listener struct {
	agentSender calls.Sender
	task        mesos.Task
	agent       mesos.AgentInfo
	fileName    string
}

func NewListener(fileName string, task mesos.Task, agentInfo mesos.AgentInfo) *Listener {
	if task.AgentID.Value != agentInfo.ID.Value {
		panic("tasks agent id doesn't match provided agent info") // err? constructor should be safe though... MustNewListener ?
	}

	agentUrl := fmt.Sprintf("http://%s/api/v1", net.JoinHostPort(agentInfo.GetHostname(), strconv.Itoa(int(agentInfo.GetPort()))))
	agentSender := httpagent.NewSender(httpcli.New(httpcli.Endpoint(agentUrl)).Send)
	return &Listener{
		agentSender: agentSender,
		task:        task,
		fileName:    fileName,
		agent:       agentInfo,
	}
}

// Listen starts listening to the specified file and streams out the content
func (l *Listener) Listen(output chan string) {
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
	// listen loop
	for {
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
			if r.GetSize() > 2000 {
				offset = r.GetSize() - 2000
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
					// TODO implement grep like filter. Use a channel to push the filter string to all listeners
					output <- fmt.Sprintf("[%s:%d]: %s", l.agent.Hostname, l.task.GetDiscovery().GetPorts().Ports[0].Number, line) //  l.task.GetTaskID().Value
				}
			}
		}
		// TODO sleep should be configurable
		time.Sleep(time.Duration(1000) * time.Millisecond)
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
