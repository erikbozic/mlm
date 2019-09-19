package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpmaster"
	"github.com/mesos/mesos-go/api/v1/lib/master/calls"
	"log"
	"mesos-monitor/commands"
	"mesos-monitor/config"
	"os"
	"strings"
)

var (
	masterSender calls.Sender
	input        *UserInput
	logStream    chan string
	done         chan struct{}
	commandStream chan commands.Command
)

func main() {
	var mesosMasterUrl string
	flag.StringVar(&mesosMasterUrl, "m", "", "http url of mesos master (e.g.: http://localhost:5050)")
	flag.Parse()

	cfg := config.ReadConfig()
	if mesosMasterUrl == "" {
		mesosMasterUrl = cfg.MesosMasterUrl
	} else {
		cfg.MesosMasterUrl = mesosMasterUrl
		cfg.SaveConfig()
	}

	input = &UserInput{}
	if mesosMasterUrl == "" {
		err := askForMesosMaster(input)
		if err != nil {
			log.Println(err.Error())
			return
		}
		mesosMasterUrl = input.MesosMasterUrl
		cfg.MesosMasterUrl = mesosMasterUrl
		cfg.SaveConfig()
	} else {
		fmt.Println("using mesos master:", mesosMasterUrl)
	}

	uri := fmt.Sprintf("%s/api/v1", mesosMasterUrl)
	masterSender = httpmaster.NewSender(httpcli.New(httpcli.Endpoint(uri)).Send)

	start(input) // asks for task selection and starts monitor/listeners in goroutines
	handleInput()
}

func start(input *UserInput) {
	input.SelectedTaskNames = nil
	fmt.Println("discovery...")
	tasks, err := getTasks()
	if err != nil {
		log.Printf("error: %s ", err.Error())
	}

	agents, err := getAgents()
	if err != nil {
		log.Printf("error: %s ", err.Error())
	}

	if len(tasks) > 0 {
		err = askForTasks(input, tasks)
	}

	logStream = make(chan string)
	done = make(chan struct{})
	commandStream = make(chan commands.Command)

	params := make([]MonitorParameter, 0)
	// build monitor params
	for name, task := range tasks {
		isSelected := false
		for _, selectedName := range input.SelectedTaskNames {
			if name == selectedName {
				isSelected = true
				break
			}
		}

		if !isSelected {
			continue
		}

		for _, taskInstance := range task {
			if agentInfo, ok := agents[taskInstance.GetAgentID().Value]; ok {
				param := MonitorParameter{
					Task:  taskInstance,
					Agent: agentInfo,
				}
				params = append(params, param)
			} // TODO didn't find the agent for this task
		}
	}
	monitor := NewMonitor(params)
	go monitor.Start(logStream, commandStream, done)
	go printLogs()
}

func handleInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		if text == ":b\n" { // back (to task selection)
			close(done)      // will stop all listeners
			close(logStream) // will stop printLogs func
			start(input)     // wil show the task selection survey again
		} else if text == ":q\n" { // quit
			close(done)      // will stop all listeners
			close(logStream) // will stop printLogs func
			log.Println("bye!")
			os.Exit(0)
		} else if text == ":a\n" { // test
			commandStream <- commands.NewTestCommand("test", nil)
		} else if strings.HasPrefix(text, ":f") { // filter
			filterText := strings.TrimSpace(strings.TrimPrefix(text, ":f"))
			commandStream <- commands.NewFilterCommand(filterText)
			log.Printf("filter set to: \"%s\" on all listeners", filterText )
		}
	}
}

func printLogs() {
	for text := range logStream {
		fmt.Println(text)
	}
}
