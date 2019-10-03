package commands

const (
	UnpauseCommandName = "unpause"
)

type UnpauseCommand struct {
}

func NewUnpauseCommand() *UnpauseCommand {
	return &UnpauseCommand{}
}

func (t *UnpauseCommand) Name() string {
	return UnpauseCommandName
}

func (t *UnpauseCommand) Parameters() []string {
	return nil
}
