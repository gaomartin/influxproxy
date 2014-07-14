package plugin

import (
	"fmt"
)

type PluginRegistry []PluginRegister

func NewPluginRegistry() (*PluginRegistry, error) {
	return &PluginRegistry{}, nil
}

func (r *PluginRegistry) RegisterPlugin(id string) {
	p, _ := NewPluginRegister(id)
	*r = append(*r, *p)
}

func (r *PluginRegistry) Print() string {
	return fmt.Sprintf("%+v\n", *r)
}

type PluginRegister struct {
	Id     string
	Port   int
	Status *PluginStatus
}

type PluginStatus struct {
	Started   bool
	Connected bool
	Failed    bool
}

func NewPluginRegister(id string) (*PluginRegister, error) {
	s := &PluginStatus{
		Started:   false,
		Connected: false,
		Failed:    false,
	}

	p := &PluginRegister{
		Id:     id,
		Port:   0,
		Status: s,
	}

	return p, nil
}
