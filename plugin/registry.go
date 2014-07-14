package plugin

import (
	"fmt"
)

type PluginRegistry []PluginRegister

func NewPluginRegistry() (*PluginRegistry, error) {
	return &PluginRegistry{}, nil
}

func (r *PluginRegistry) RegisterPlugin(name string) {
	p, _ := NewPluginRegister(name)
	*r = append(*r, *p)
}

func (r *PluginRegistry) GetPlugins() []PluginRegister {
	return *r
}

func (r *PluginRegistry) Print() string {
	var out string
	for index, plugin := range r.GetPlugins() {
		out += fmt.Sprintf("---\nPLUGIN %v\n  NAME:   %v,\n  PORT:   %v,\n  STATUS: %+v\n", index, plugin.Name, plugin.Port, plugin.Status)
	}
	return out
}

type PluginRegister struct {
	Name   string
	Port   int
	Status *PluginStatus
}

type PluginStatus struct {
	Started   bool
	Connected bool
	Failed    bool
}

func NewPluginRegister(name string) (*PluginRegister, error) {
	s := &PluginStatus{
		Started:   false,
		Connected: false,
		Failed:    false,
	}

	p := &PluginRegister{
		Name:   name,
		Port:   0,
		Status: s,
	}

	return p, nil
}
