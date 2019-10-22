package main

import (
	"github.com/mesos/mesos-go/api/v1/lib/master/calls"
	"log"
	"mlm/commands"
	"mlm/tui"
	"os"
)

var (
	masterSender  calls.Sender
	input         *UserInput
	logStream     chan string
	done          chan struct{}
	commandStream chan commands.Command
	mesosMasterUrl string
)

func main() {

	f,_ := os.Create("mlm.log")
	log.SetOutput(f)
	log.Println("main")
	app := tui.NewTui()
	if err := app.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}


}


//func start(input *UserInput) {
//	fmt.Println("discovery...")
//	tasks, err := getTasks()
//	if err != nil {
//		log.Printf("error: %s ", err.Error())
//	}
//
//	agents, err := getAgents()
//	if err != nil {
//		log.Printf("error: %s ", err.Error())
//	}
//
//	if len(tasks) > 0 {
//		taskNames := make([]string, len(tasks))
//		i := 0
//		for name := range tasks {
//			taskNames[i] = name
//			i++
//		}
//		err = askForTasks(input, taskNames)
//		if err != nil {
//			log.Fatal("error selecting tasks")
//		}
//	} else {
//		log.Println("didn't get any active tasks from master!\nbye!")
//		os.Exit(0)
//	}
//
//	commandStream = make(chan commands.Command)
//	logStream = make(chan string)
//	done = make(chan struct{})
//
//	params := make([]*monitor.Parameter, 0)
//	// build monitor params
//	for name, task := range tasks {
//		isSelected := false
//		for _, selectedName := range input.SelectedTaskNames {
//			if name == selectedName {
//				isSelected = true
//				break
//			}
//		}
//
//		if !isSelected {
//			continue
//		}
//
//		for _, taskInstance := range task {
//			if agentInfo, ok := agents[taskInstance.GetAgentID().Value]; ok {
//				param := &monitor.Parameter{
//					Task:  taskInstance,
//					Agent: agentInfo,
//					Files: []string{"stdout", "stderr"},
//				}
//				params = append(params, param)
//			} else {
//				log.Printf("didn't find agent on which task %s is running", taskInstance.GetTaskID().Value)
//			}
//		}
//	}
//	// Url must be ok, since we already used it
//
//	mon := monitor.NewMonitor(params, mesosMasterUrl)
//	go mon.Start(logStream, commandStream, done)
//
//}
//
//func handleInput() {
//	reader := bufio.NewReader(os.Stdin)
//	for {
//		text, _ := reader.ReadString('\n')
//		if text == ":b\n" { 			  // back (to task selection)
//			close(done)                   // will stop all listeners
//			close(logStream)              // will stop printLogs func
//			input.SelectedTaskNames = nil // reset selected task names
//			start(input)                  // wil show the task selection survey again
//		} else if text == ":q\n" { // quit
//			close(done)
//			close(logStream)
//			log.Println("bye!")
//			os.Exit(0)
//		} else if strings.HasPrefix(text, ":f") { // filter
//			filterText := strings.TrimSpace(strings.TrimPrefix(text, ":f"))
//			commandStream <- commands.NewFilterCommand(filterText)
//			log.Printf("filter set to: \"%s\" on all listeners", filterText)
//		} else if text == ":p\n" { // pause
//			commandStream <- commands.NewPauseCommand()
//			log.Printf("paused all listeners")
//		} else if text == ":u\n" { // unpause
//			commandStream <- commands.NewUnpauseCommand()
//			log.Printf("unpaused all listeners")
//		}
//	}
//}
//
//func printLogs() {
//	for text := range logStream {
//		fmt.Println(text)
//	}
//}
