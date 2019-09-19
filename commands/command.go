package commands

type Command interface {
	Name() string
	Parameters() []string
}
