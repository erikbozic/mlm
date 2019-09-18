package main

import (
	"context"
	"fmt"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/agent/calls"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpagent"
	"github.com/mesos/mesos-go/api/v1/lib/master"
	"net"
	"strconv"
	"strings"
	"time"
)

// Listener streams the content of a file
type Listener struct {
	agentSender calls.Sender
	task        mesos.Task
	filePath    string
}

func NewListener(filePath string, task mesos.Task, agentInfo mesos.AgentInfo) *Listener {
	if task.AgentID.Value != agentInfo.ID.Value {
		panic("tasks agent id doesn't match provided agent info") // err? constructor should be safe though... MustNewListener ?
	}

	agentUrl := fmt.Sprintf("http://%s/api/v1", net.JoinHostPort(agentInfo.GetHostname(), strconv.Itoa(int(agentInfo.GetPort()))))
	agentSender := httpagent.NewSender(httpcli.New(httpcli.Endpoint(agentUrl)).Send)
	return &Listener{
		agentSender: agentSender,
		task:        task,
		filePath:    filePath,
	}
}

// Listen starts listening to the specified file and streams out the content
func (l *Listener) Listen(output chan string) error {
	offset := uint64(0)
	for {
		resp, err := l.agentSender.Send(context.TODO(), calls.NonStreaming(calls.ReadFile(l.filePath, offset)))

		if err != nil {
			return err
		}

		var e master.Response

		err = resp.Decode(&e)
		if err != nil {
			return err
		}

		r := e.GetReadFile()
		data := r.GetData()
		offset = r.GetSize()
		if len(data) != 0 {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if len(strings.TrimSpace(line)) > 0 {
					// TODO use templates
					// TODO implement grep like filter. Use a channel to push the filter string to all listeners
					output <- fmt.Sprintf("[%s]: %s", l.task.GetTaskID().Value, line)
				}
			}
		}
		// TODO sleep should be configurable
		time.Sleep(time.Duration(1000) * time.Millisecond)
	}
}
