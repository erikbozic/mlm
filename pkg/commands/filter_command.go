package commands

const (
	FilterCommandName = "filter"
)

type FilterCommand struct {
	parameters []string
}

func NewFilterCommand(filterString string) *FilterCommand {
	return &FilterCommand{
		parameters: []string{filterString},
	}
}

func (t *FilterCommand) Name() string {
	return FilterCommandName
}

func (t *FilterCommand) Parameters() []string {
	return t.parameters
}
