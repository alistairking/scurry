package measurement

type Noop struct{}

func (n Noop) AsCommand() string {
	return ""
}
