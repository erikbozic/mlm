package main

import (
	"context"
	"fmt"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpmaster"
	"github.com/mesos/mesos-go/api/v1/lib/master"
	"github.com/mesos/mesos-go/api/v1/lib/master/calls"
	"strings"
	"time"
)

type TaskMonitor struct {
	task        mesos.Task
	agentInfo   mesos.AgentInfo
	agentSender calls.Sender
}

func NewTaskMonitor(task mesos.Task, agent mesos.AgentInfo) *TaskMonitor {
	agentUrl := fmt.Sprintf("http://%s:%d/api/v1", agent.GetHostname(), agent.GetPort())
	agentSender := httpmaster.NewSender(httpcli.New(httpcli.Endpoint(agentUrl)).Send)

	return &TaskMonitor{
		task:        task,
		agentSender: agentSender,
	}
}

func (m *TaskMonitor) StartReadingLogs(output chan string) error {
	agentId := m.task.GetAgentID().Value
	frameworkId := m.task.GetFrameworkID().Value
	taskId := m.task.GetTaskID().Value
	executorId := m.task.GetStatuses()[0].GetContainerStatus().GetContainerID().Value

	// var/lib/mesos/agent/slaves/df471129-e2a3-49b3-83f5-ed5937abb659-S0/frameworks/df471129-e2a3-49b3-83f5-ed5937abb659-0000/executors/marco-polo2.6d019081-d979-11e9-afcb-02423fbab6ee/runs/2918b0c4-3c38-41d0-bd21-287e0a23486a/stdout
	filePath := fmt.Sprintf("/var/lib/mesos/agent/slaves/%s/frameworks/%s/executors/%s/runs/%s/stdout", agentId, frameworkId, taskId, executorId)

	offset := uint64(0)
	for {
		resp, err := m.agentSender.Send(context.TODO(), calls.NonStreaming(calls.ReadFile(filePath, offset)))

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
					output <- fmt.Sprintf("[%s]: %s", m.task.GetTaskID().Value, line)
				}
			}
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}
