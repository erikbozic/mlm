package main

import (
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"log"
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
func (m *Monitor) Start(output chan string, done chan struct{}) {
	for _, p := range m.parameters {
		// TODO filename could be configurable

		stdOutListener, err := NewListener("stdout", p.Task, p.Agent)
		if err != nil {
			log.Println("error creating listener: ", err.Error())
			continue
		}

		stdErrListener, err := NewListener("stderr", p.Task, p.Agent)
		if err != nil {
			log.Println("error creating listener: ", err.Error())
			continue
		}

		// TODO what about: http://mesos.apache.org/documentation/latest/operator-http-api/#attach_container_output
		go stdOutListener.Listen(output, done)
		go stdErrListener.Listen(output, done)

	}
}
