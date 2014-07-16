package orchestrator

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
)

// ---------------------------------------------------------------------------------
// Orchestrator
// ---------------------------------------------------------------------------------

//
type Orchestrator struct {
	Config    *OrchestratorConfiguration
	Registry  *PluginRegistry
	Connector *Connector
	Port      int
}

// NewOrchestrator returnd a fully initialized orchestrator
func NewOrchestrator(conf *OrchestratorConfiguration) (*Orchestrator, error) {
	if conf.Address == "" || conf.PluginMaxPort == 0 || conf.PluginMinPort == 0 {
		return nil, errors.New("Insufficent orchestrator configuration")
	}
	var out string
	var err error

	o := &Orchestrator{
		Config: conf,
	}

	o.Registry = NewPluginRegistry()

	for _, plugin := range o.Config.Plugins {
		err = o.Registry.RegisterPlugin(plugin, o.Config.Address)
		if err != nil {
			out += err.Error()
		}
	}

	if err != nil {
		err = errors.New(out)
	}
	return o, err
}

// Starts the orchestrator instance and all its Plugins concurrently
func (orch *Orchestrator) Start() ([]string, error) {
	var err error
	var messages []string

	// get the orchestrator itself started
	orchChan := make(chan bool)
	go func() {
		err = orch.spinup(orchChan)
	}()
	<-orchChan
	if err != nil {
		messages = append(messages, "Could not launch orchestrator")
		return messages, err
	}

	messages = append(messages, fmt.Sprintf("Orchestrator started on port %v.", orch.Port))

	// get plugins started concurrently
	pluginChan := make(chan bool)
	go func() {
		for _, plugin := range *orch.Registry {
			err = plugin.Spinup(orch)
			if err != nil {
				messages = append(messages, fmt.Sprintf("Plugin %s could not be loaded: %s. ", plugin.Name, err))
			} else {
				messages = append(messages, fmt.Sprintf("Plugin %s successfully loaded. ", plugin.Name))
			}
		}
		pluginChan <- true
	}()
	<-pluginChan
	messages = append(messages, "All plugins loaded")
	return messages, nil
}

// Inits the Orchestrator
func (orch *Orchestrator) spinup(done chan bool) error {
	connector := NewConnector(orch.Registry)
	orch.Connector = connector

	rpc.Register(orch.Connector)

	ln, port, err := orch.getListener()
	if err != nil {
		return err
	}

	orch.Port = port
	done <- true

	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(c)
	}

	return nil
}

func (orch *Orchestrator) getEnv() []string {
	env := []string{
		fmt.Sprintf("ORCHESTRATOR_CONN_STRING=%s:%v", orch.Config.Address, orch.Port),
		fmt.Sprintf("PLUGIN_MIN_PORT=%d", orch.Config.PluginMinPort),
		fmt.Sprintf("PLUGIN_MAX_PORT=%d", orch.Config.PluginMaxPort),
		fmt.Sprintf("PLUGIN_ADDRESS=%s", orch.Config.Address),
	}
	env = append(os.Environ(), env...)
	return env
}

func (orch *Orchestrator) getListener() (net.Listener, int, error) {
	for port := orch.Config.PluginMinPort; port <= orch.Config.PluginMaxPort; port++ {
		connection := fmt.Sprintf("%s:%d", orch.Config.Address, port)
		listener, err := net.Listen("tcp", connection)
		if err == nil {
			return listener, port, nil
		}
	}

	return nil, 0, errors.New("Could not get TCP listener, maybe all ports are already used")
}

// ---------------------------------------------------------------------------------
// OrchestratorConfiguration
// ---------------------------------------------------------------------------------

// OrchestratorConfiguration hold all required configuration data
type OrchestratorConfiguration struct {
	Address       string
	PluginMinPort int
	PluginMaxPort int
	Plugins       []string
}
