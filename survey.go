package main

import (
	"github.com/AlecAivazis/survey/v2"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
)

type UserInput struct {
	MesosMasterUrl    string
	SelectedTaskNames []string
}

func askForMesosMaster(input *UserInput) error {
	masterQ := []*survey.Question{
		{
			Name: "MesosMasterUrl",
			Prompt: &survey.Input{
				Message: "Enter mesos master http url:",
				Default: "http://localhost:5050",
				Help:    "e.g.: http://localhost:5050",
			},
			Validate: survey.Required,
		},
	}

	err := survey.Ask(masterQ, input)
	if err != nil {
		return err
	}
	return nil
}

func askForTasks(input *UserInput, tasks map[string][]mesos.Task) error {
	taskNames := make([]string, len(tasks))
	i := 0
	for name, _ := range tasks {
		taskNames[i] = name
		i++
	}

	tasksQ := []*survey.Question{
		{
			Name: "SelectedTaskNames",
			Prompt: &survey.MultiSelect{
				Message: "Selected tasks which you wish to monitor",
				Options: taskNames,
				Help:    "Will monitor all instances of selected tasks",
				PageSize: 15,
			},
			Validate: survey.Required,
		},
	}

	err := survey.Ask(tasksQ, input)
	if err != nil {
		return err
	}
	return nil
}
