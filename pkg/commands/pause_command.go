package commands

const (
	PauseCommandName = "pause"
)

type PauseCommand struct {
}

func NewPauseCommand() *PauseCommand {
	return &PauseCommand{}
}

func (t *PauseCommand) Name() string {
	return PauseCommandName
}

func (t *PauseCommand) Parameters() []string {
	return nil
}
