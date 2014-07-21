package plugin

// ---------------------------------------------------------------------------------
// Description
// ---------------------------------------------------------------------------------

// Description holds generic information about the plugin. The Description is privided
// to the orchestrator via Descripe RPC call.
type Description struct {
	Description string     `json:"description"`
	Author      string     `json:"author"`
	Version     string     `json:"version"`
	Arguments   []Argument `json:"arguments"`
}

// ---------------------------------------------------------------------------------
// Argument
// ---------------------------------------------------------------------------------

// Argument hold information about the possible arguments that can be provided to the
// Plugin on a Run RPC call.
type Argument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Default     string `json:"default"`
	Optional    bool   `json:"optional"`
}
