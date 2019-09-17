package main

import (
	"context"
	"fmt"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpmaster"
	"github.com/mesos/mesos-go/api/v1/lib/master"
	"github.com/mesos/mesos-go/api/v1/lib/master/calls"
	"log"
	"net"
	"strconv"
	"time"
)

var (
	masterSender calls.Sender
)

func main() {
	masterHost := "localhost"
	masterPort := 5050

	uri := fmt.Sprintf("http://%s/api/v1", net.JoinHostPort(masterHost, strconv.Itoa(masterPort)))

	masterSender = httpmaster.NewSender(httpcli.New(httpcli.Endpoint(uri)).Send)

	tasks, err := getTasks()
	if err != nil {
		log.Printf("error: %s ", err.Error())
	}

	agents, err := getAgents()
	if err != nil {
		log.Printf("error: %s ", err.Error())
	}

	selectedTaskName := "marco-polo"

	for name, t := range tasks {
		if name != selectedTaskName {
			continue
		}
		out := make(chan string)
		for _, v := range t {
			monitor := NewTaskMonitor(v, agents[v.GetAgentID().Value])
			go monitor.StartReadingLogs(out)
		}
		go printLogs(out)
	}

	for {
		//printTasks(tasks)
		//printAgents(agents)
		time.Sleep(time.Duration(3) * time.Second)
	}
}

func printLogs(logs chan string) {
	for text := range logs {
		fmt.Println(text)
	}
}

func getTasks() (tasks map[string][]mesos.Task, err error) {
	resp, err := masterSender.Send(context.TODO(), calls.NonStreaming(calls.GetTasks()))
	tasks = make(map[string][]mesos.Task)
	if err != nil {
		return tasks, err
	}
	defer func() {
		if resp != nil {
			err = resp.Close()
		}
	}()

	var e master.Response
	if err := resp.Decode(&e); err != nil {
		return tasks, err
	}

	for _, task := range e.GetTasks.Tasks {
		if v, ok := tasks[task.GetName()]; ok {
			tasks[task.GetName()] = append(v, task)
		} else {
			tasks[task.GetName()] = []mesos.Task{task}
		}
	}
	return tasks, err
}

func getAgents() (agents map[string]mesos.AgentInfo, err error) {
	resp, err := masterSender.Send(context.TODO(), calls.NonStreaming(calls.GetAgents()))
	agents = make(map[string]mesos.AgentInfo, 0)
	if err != nil {
		return agents, err
	}
	defer func() {
		if resp != nil {
			err = resp.Close()
		}
	}()

	var e master.Response
	if err := resp.Decode(&e); err != nil {
		return agents, err
	}

	for _, agent := range e.GetGetAgents().GetAgents() {
		agentInfo := agent.GetAgentInfo()
		agents[agentInfo.GetID().Value] = agentInfo
	}
	return agents, err
}
