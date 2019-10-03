package main

import (
	"context"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/master"
	"github.com/mesos/mesos-go/api/v1/lib/master/calls"
)

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
		if v, ok := tasks[task.GetName()]; ok { // TODO make sure this only includes active tasks
			tasks[task.GetName()] = append(v, task)
		} else {
			tasks[task.GetName()] = []mesos.Task{task}
		}
	}
	return tasks, err
}

func getAgents() (agents map[string]mesos.AgentInfo, err error) {
	resp, err := masterSender.Send(context.TODO(), calls.NonStreaming(calls.GetAgents()))
	agents = make(map[string]mesos.AgentInfo)
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
