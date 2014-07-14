package plugin

import (
	"fmt"
)

type Orchestrator struct {
	Config   *OrchestratorConfiguration
	Registry *PluginRegistry
}

// NewOrchestrator returnd a fully initialized orchestrator
func NewOrchestrator(conf *OrchestratorConfiguration) (*Orchestrator, error) {
	o := &Orchestrator{
		Config: conf,
	}

	o.Registry, _ = NewPluginRegistry()

	for _, plugin := range o.Config.Plugins {
		o.Registry.RegisterPlugin(plugin)
	}

	return o, nil
}

func (orch *Orchestrator) PrintConfiguration() string {
	return fmt.Sprintf("%+v\n", *orch.Config)
}

func (orch *Orchestrator) Start() error {
	// ...
	return nil
}

type OrchestratorConfiguration struct {
	Address string
	MinPort int
	MaxPort int
	Plugins []string
}

func (conf *OrchestratorConfiguration) Print() string {
	return fmt.Sprintf("%+v\n", *conf)
}
