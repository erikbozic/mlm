package main

import (
	"fmt"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpmaster"
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

	for _, t := range tasks {
		out := make(chan string)
		params := make([]MonitorParameter, 0)
		for _, v := range t {
			if agentInfo, ok := agents[v.GetAgentID().Value]; ok {
				param := MonitorParameter{
					Task:  v,
					Agent: agentInfo,
				}
				params = append(params, param)
			}
		}

		monitor := NewMonitor(params)
		go monitor.Start(out)
		go printLogs(out)
	}

	for {
		//printTasks(parameters)
		//printAgents(agents)
		time.Sleep(time.Duration(3) * time.Second)
	}
}

func printLogs(logs chan string) {
	for text := range logs {
		fmt.Println(text)
	}
}

