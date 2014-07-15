package orchestrator

import (
	"errors"
	"fmt"
	"path/filepath"
)

type PluginRegistry []*PluginBroker

func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{}
}

func (r *PluginRegistry) RegisterPlugin(path string) error {
	name := filepath.Base(path)
	for _, p := range *r {
		if p.Name == name {
			return errors.New("Plugin '" + name + "' is already registered, '" + path + "' not registered. ")
		}
	}
	p, _ := NewPluginBroker(name, path)
	*r = append(*r, p)
	return nil
}

func (r *PluginRegistry) GetPluginByName(name string) *PluginBroker {
	for _, p := range *r {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func (r *PluginRegistry) Print() string {
	var out string
	for i, p := range *r {
		out += fmt.Sprintf("\nPLUGIN %v\n    Name:   %v\n    Path:   %v\n    Port:   %v\n    Status: %v", i, p.Name, p.Path, p.Port, p.Status.Print())
	}
	return out
}
