package main

import (
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"log"
	"mlm/commands"
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
	// Names of files to monitor
	Files []string
}

func NewMonitor(parameters []MonitorParameter) *Monitor {
	return &Monitor{
		parameters: parameters,
	}
}

// Start sets up listeners for all specified tasks
func (m *Monitor) Start(output chan string, commandStream <-chan commands.Command, done chan struct{}) {
	commandChannels := make([]chan<- commands.Command, 0)
	for _, p := range m.parameters {
		for _, fileName := range p.Files {
			listener, err := NewListener(fileName, p.Task, p.Agent)
			if err != nil {
				log.Println("error creating listener: ", err.Error())
				continue
			}
			cmdChannel := make(chan commands.Command)
			commandChannels = append(commandChannels, cmdChannel)
			go listener.Listen(output, cmdChannel, done)
		}
	}

	for {
		select {
		case command, ok := <-commandStream:
			if !ok {
				log.Printf("closed command channel... ?")
				for _, c := range commandChannels {
					close(c)
				}
				return
			}
			for _, c := range commandChannels {
				c <- command
			}
		}
	}
}
