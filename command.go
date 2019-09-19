package main

const (
	FilterCommandName = "filter"
)

type Command interface {
	Name() string
	Parameters() []string
}

// test_command.go
type TestCommand struct {
	name string
	parameters []string
}

func NewTestCommand(name string, parameters []string) *TestCommand {
	return &TestCommand{
		name:       name,
		parameters: parameters,
	}
}

func (t *TestCommand) Name() string {
	return t.name
}

func (t *TestCommand) Parameters() []string {
	return t.parameters
}

// filter_command.go
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
