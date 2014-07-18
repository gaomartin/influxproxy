package orchestrator

import (
	"errors"
	"path/filepath"
)

type BrokerRegistry []*PluginBroker

func NewBrokerRegistry() *BrokerRegistry {
	return &BrokerRegistry{}
}

func (r *BrokerRegistry) RegisterBroker(plugin string) error {
	name := filepath.Base(plugin)
	for _, b := range *r {
		if b.Name == name {
			return errors.New("Broker of plugin '" + name + "' is already registered, plugin '" + plugin + "' not registered. ")
		}
	}
	b, _ := NewPluginBroker(name, plugin)
	*r = append(*r, b)
	return nil
}

func (r *BrokerRegistry) GetBrokerByName(name string) *PluginBroker {
	for _, b := range *r {
		if b.Name == name {
			return b
		}
	}
	return nil
}
