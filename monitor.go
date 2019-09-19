package main

import (
	mesos "github.com/mesos/mesos-go/api/v1/lib"
)

// Monitors tasks
type Monitor struct {
	parameters []MonitorParameter
}

type MonitorParameter struct {
	// Task which to monitor
	Task mesos.Task
	// Agent on which the task is running
	Agent mesos.AgentInfo
}

func NewMonitor(parameters []MonitorParameter) *Monitor {
	return &Monitor{
		parameters: parameters,
	}
}

// Start sets up listeners for all specified tasks
func (m *Monitor) Start(output chan string, commandStream chan string) {
	for _, p := range m.parameters {
		// TODO filename could be configurable
		stdOutListener := NewListener("stdout", p.Task, p.Agent)
		stdErrListener := NewListener("stderr", p.Task, p.Agent)
		// TODO what about: http://mesos.apache.org/documentation/latest/operator-http-api/#attach_container_output
		go stdOutListener.Listen(output, commandStream)
		go stdErrListener.Listen(output, commandStream)
		// TODO handle errors and cancellation
	}
}
