package tui

import (
	"fmt"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/rivo/tview"
	"log"
	"mlm/commands"
	"mlm/monitor"
	"mlm/tui/views"
	"time"
)

type Tui struct {
	LogView   *views.LogsView
	TasksView *views.TasksView
	App       *tview.Application

	mon       *monitor.Monitor

	taskEventChannel chan interface{}
	output           chan string
	commandStream    <-chan commands.Command
	done             chan struct{}

	// holds current tasks received from master (grouped by task name)
	tasks  map[string][]mesos.Task
	// holds all agents master has access to
	agents map[string]mesos.AgentInfo
}

func NewTui() *Tui {
	app := tview.NewApplication()
	taskEventChannel := make(chan interface{})
	tasks := views.NewTasksViews(taskEventChannel)
	logs := views.NewLogsView()

	t := &Tui{
		LogView:          logs,
		TasksView:        tasks,
		App:              app,
		taskEventChannel: taskEventChannel,
	}
	return t
}

func (t *Tui) updateTasks() {
	for {
		tasks, err := t.mon.GetTasks()
		if err != nil {
			panic(err)
		}
		t.App.QueueUpdateDraw(func() {
			t.TasksView.SetTasks(tasks)
		})
		t.tasks = tasks
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func (t *Tui) Run() error {
	grid := tview.NewGrid().
		SetRows(0).
		SetColumns(-1, -3).
		AddItem(t.TasksView, 0, 0, 1, 1, 0, 0, true).
		AddItem(t.LogView, 0, 1, 1, 1, 0, 0, false)

	masterUrl := "http://localhost:5050"
	t.mon = monitor.NewMonitor(masterUrl)

	t.commandStream = make(chan commands.Command)
	t.output = make(chan string)
	t.done = make(chan struct{})

	go t.handleTaskEvents(t.taskEventChannel)
	go t.updateTasks()

	go t.updateAgents()

	go t.printLogs(t.output)

	// TODO just for testing ...
	go func() {
		for {
			time.Sleep(time.Duration(500) * time.Millisecond)
			t.output <- fmt.Sprintf("log at %v\n", time.Now())
		}
	}()


	// let monitor know of the event (selection, cancel ...)
	go t.mon.Start(t.output, t.commandStream, t.done)

	if err := t.App.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
	return nil
}

func (t *Tui) handleTaskEvents(taskEventChannel chan interface{}) {
	for val := range taskEventChannel {
		log.Println("handled:", val)
	}
}

func (t *Tui) printLogs(strings chan string) {
	for s := range strings {
		log.Println("printed", s)
		t.App.QueueUpdateDraw(func(){
			_, _ = t.LogView.Write([]byte(s))
		})
	}
}

func (t *Tui) updateAgents() {

	for {
		agents, err := t.mon.GetAgents()
		if err != nil {
			log.Printf("error: %s ", err.Error())
		}
		t.agents = agents
		time.Sleep(time.Duration(1) * time.Second)
	}

}
