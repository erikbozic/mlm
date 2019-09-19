package commands

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
