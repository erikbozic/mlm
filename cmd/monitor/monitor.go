package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpmaster"
	"github.com/mesos/mesos-go/api/v1/lib/master/calls"
	"log"
	"mlm/internal/config"
	"mlm/pkg/commands"
	"mlm/pkg/monitor"
	"os"
	"strings"
)

var (
	masterSender  calls.Sender
	input         *UserInput
	logStream     chan string
	done          chan struct{}
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
		taskNames := make([]string, len(tasks))
		i := 0
		for name := range tasks {
			taskNames[i] = name
			i++
		}
		err = askForTasks(input, taskNames)
	} else {
		log.Println("didn't get any active tasks from master!\nbye!")
		os.Exit(0)
	}

	commandStream = make(chan commands.Command)
	logStream = make(chan string)
	done = make(chan struct{})

	params := make([]*monitor.Parameter, 0)
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
				param := &monitor.Parameter{
					Task:  taskInstance,
					Agent: agentInfo,
					Files: []string{"stdout", "stderr"},
				}
				params = append(params, param)
			} else {
				log.Printf("didn't find agent on which task %s is running", taskInstance.GetTaskID().Value)
			}
		}
	}
	mon := monitor.NewMonitor(params)
	go mon.Start(logStream, commandStream, done)
	go printLogs()
}

func handleInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		if text == ":b\n" { // back (to task selection)
			close(done)                   // will stop all listeners
			close(logStream)              // will stop printLogs func
			input.SelectedTaskNames = nil // reset selected task names
			start(input)                  // wil show the task selection survey again
		} else if text == ":q\n" { // quit
			close(done)
			close(logStream)
			log.Println("bye!")
			os.Exit(0)
		} else if strings.HasPrefix(text, ":f") { // filter
			filterText := strings.TrimSpace(strings.TrimPrefix(text, ":f"))
			commandStream <- commands.NewFilterCommand(filterText)
			log.Printf("filter set to: \"%s\" on all listeners", filterText)
		} else if text == ":p\n" { // pause
			commandStream <- commands.NewPauseCommand()
			log.Printf("paused all listeners")
		} else if text == ":u\n" { // unpause
			commandStream <- commands.NewUnpauseCommand()
			log.Printf("unpaused all listeners")
		}
	}
}

func printLogs() {
	for text := range logStream {
		fmt.Println(text)
	}
}
