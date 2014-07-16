package plugin

type Description struct {
	Description string
	Author      string
	Version     string
	Arguments   []*Argument
}

type Argument struct {
	Name        string
	Description string
}
