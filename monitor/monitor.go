package monitor

import (
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"log"
	"mlm/commands"
	"net/url"
)

// Monitors tasks
type Monitor struct {
	parameters []*Parameter
	masterUrl  *url.URL
}

var (
	Colors = []string{
		"\u001b[31m", // red
		"\u001b[32m", // green
		"\u001b[33m", // yellow
		"\u001b[34m", // blue
		"\u001b[35m", // magenta
		"\u001b[36m", // cyan
		"\u001b[37m", // white
		"\u001b[0m",  // reset
	}
)

type Parameter struct {
	// Task which to monitor
	Task mesos.Task
	// Agent on which the task is running
	Agent mesos.AgentInfo
	// Names of files to monitor
	Files []string
	// Ansi color string for output
	color string
}

func NewMonitor(parameters []*Parameter, masterUrl *url.URL) *Monitor {
	return &Monitor{
		parameters: parameters,
		masterUrl:  masterUrl,
	}
}

func SetLogColor(params []*Parameter) {
	for i, p := range params {
		p.color = Colors[i%len(Colors)]
	}
}

// Start sets up listeners for all specified tasks
func (m *Monitor) Start(output chan string, commandStream <-chan commands.Command, done chan struct{}) {
	commandChannels := make([]chan<- commands.Command, 0)
	SetLogColor(m.parameters)

	for _, p := range m.parameters {
		for _, fileName := range p.Files {
			listener, err := NewListener(fileName, p.Task, p.Agent, p.color, m.masterUrl.Scheme)
			if err != nil {
				log.Println("error creating listener: ", err.Error())
				continue
			}
			cmdChannel := make(chan commands.Command)
			commandChannels = append(commandChannels, cmdChannel)
			go listener.Listen(output, cmdChannel, done)
		}
	}
	// range through the commandStream until it closes and fan them out to listeners' command channels
	for command := range commandStream {
		for _, c := range commandChannels {
			c <- command
		}
	}
	for _, c := range commandChannels {
		close(c)
	}
}
