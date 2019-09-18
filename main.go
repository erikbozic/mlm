package main

import (
	"flag"
	"fmt"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpmaster"
	"github.com/mesos/mesos-go/api/v1/lib/master/calls"
	"log"
	"time"
)

var (
	masterSender calls.Sender
)

func main() {

	var mesosMasterUrl string

	flag.StringVar(&mesosMasterUrl, "master", "", "http url of mesos master (e.g.: http://localhost:5050) " )
	flag.Parse()

	input := UserInput{}
	if mesosMasterUrl == "" {
		err := askForMesosMaster(&input)
		if err != nil {
			log.Println(err.Error())
			return
		}
		mesosMasterUrl = input.MesosMasterUrl
	}

	uri := fmt.Sprintf("%s/api/v1", mesosMasterUrl)
	masterSender = httpmaster.NewSender(httpcli.New(httpcli.Endpoint(uri)).Send)
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
		err = askForTasks(&input, tasks)
	}


	for name, t := range tasks {
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

	// TODO surely there's a better way of doing this?
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

