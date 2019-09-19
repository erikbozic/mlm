package main

import (
	"fmt"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"log"
	"mesos-monitor/commands"
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
func (m *Monitor) Start(output chan string, commandStream chan commands.Command, done chan struct{}) {
	commandChannels := make([]chan commands.Command, 0)
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

		c1 := make(chan commands.Command)
		c2 := make(chan commands.Command)
		commandChannels = append(commandChannels, c1)
		commandChannels = append(commandChannels, c2)

		go stdOutListener.Listen(output, c1, done)
		go stdErrListener.Listen(output, c2, done)
	}

	for {
		select {
		case command, ok := <-commandStream:
			if !ok {
				fmt.Printf("closed command channel... ?")
				return
			}
			for _, c := range commandChannels {
				c <- command
			}
		}
	}
}
