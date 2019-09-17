package main

import (
	"fmt"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
)

func printAgents(agents map[string]mesos.AgentInfo) {
	for _, agent := range agents {
		fmt.Println()
		fmt.Println("agent id: ", agent.GetID().Value)
		fmt.Println("agent hostname: ", agent.GetHostname())
		fmt.Println("agent port: ", agent.GetPort())
	}
	fmt.Println()
}

func printTasks(tasks map[string][]mesos.Task) {
	fmt.Println()
	for name, task := range tasks {
		fmt.Println("task name: ", name)
		fmt.Println("instances: ")
		for _, t := range task {
			fmt.Println("task id: ", t.GetTaskID().Value)
			fmt.Println("agent id:", t.GetAgentID().Value)
			fmt.Println()
		}
		fmt.Println()
	}
	fmt.Println()
}
