package main

import (
	"fmt"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
)

// Monitors tasks
type Monitor struct {
	parameters []MonitorParameter
}

type MonitorParameter struct {
	// Task which to monitor
	Task  mesos.Task
	// Agent on which the task is running
	Agent mesos.AgentInfo
}

func NewMonitor(parameters []MonitorParameter) *Monitor {
	return &Monitor{
		parameters: parameters,
	}
}

// Start sets up listeners to all specified tasks
func (m *Monitor) Start(output chan string) {
	for _, p := range m.parameters {
		agentId := p.Task.GetAgentID().Value
		frameworkId := p.Task.GetFrameworkID().Value
		taskId := p.Task.GetTaskID().Value
		executorId := p.Task.GetStatuses()[0].GetContainerStatus().GetContainerID().Value

		// TODO base path (agent work dir) should be configurable
		// TODO filename could be configurable
		// var/lib/mesos/agent/slaves/df471129-e2a3-49b3-83f5-ed5937abb659-S0/frameworks/df471129-e2a3-49b3-83f5-ed5937abb659-0000/executors/marco-polo2.6d019081-d979-11e9-afcb-02423fbab6ee/runs/2918b0c4-3c38-41d0-bd21-287e0a23486a/stdout
		stdOutPath := fmt.Sprintf("/var/lib/mesos/agent/slaves/%s/frameworks/%s/executors/%s/runs/%s/stdout", agentId, frameworkId, taskId, executorId)
		stdErrPath := fmt.Sprintf("/var/lib/mesos/agent/slaves/%s/frameworks/%s/executors/%s/runs/%s/stderr", agentId, frameworkId, taskId, executorId)

		stdOutListener := NewListener(stdOutPath, p.Task, p.Agent)
		stdErrListener := NewListener(stdErrPath, p.Task, p.Agent)
		go stdOutListener.Listen(output)
		go stdErrListener.Listen(output)
		// TODO handle errors and cancellation
	}
}
