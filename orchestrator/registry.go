package orchestrator

import (
	"errors"
	"path/filepath"
)

type PluginRegistry []*PluginBroker

func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{}
}

func (r *PluginRegistry) RegisterPlugin(path string, address string) error {
	name := filepath.Base(path)
	for _, p := range *r {
		if p.Name == name {
			return errors.New("Plugin '" + name + "' is already registered, '" + path + "' not registered. ")
		}
	}
	p, _ := NewPluginBroker(name, path, address)
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
