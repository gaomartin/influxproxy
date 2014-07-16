package plugin

type Description struct {
	Description string     `json:"description"`
	Author      string     `json:"author"`
	Version     string     `json:"version"`
	Arguments   []Argument `json:"arguments"`
}

type Argument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Optional    bool   `json:"optional"`
}
