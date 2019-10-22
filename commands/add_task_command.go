package commands

const (
	AddTaskCommandName = "add_task"
)


type AddTaskCommand struct {

}

func NewAddTaskCommand(taskName string) {

}

func (a *AddTaskCommand) Name() string {
	return AddTaskCommandName
}

func (a AddTaskCommand) Parameters() []string {
	return nil
}

