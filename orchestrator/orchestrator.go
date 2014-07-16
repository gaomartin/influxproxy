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
func (orch *Orchestrator) Start() error {
	// get the orchestrator itself started
	orchChan := make(chan bool)
	go func() {
		orch.spinup(orchChan)
	}()
	<-orchChan

	// get plugins started concurrently
	pluginChan := make(chan bool)
	go func() {
		for _, plugin := range *orch.Registry {
			plugin.Spinup(orch)
		}
		pluginChan <- true
	}()
	<-pluginChan

	return nil
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
	out := fmt.Sprintf("Orchestrator is listening on port %v", orch.Port)
	fmt.Println(out)
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
		fmt.Sprintf("PLUGIN_MIN_PORT=%d", orch.Config.MinPort),
		fmt.Sprintf("PLUGIN_MAX_PORT=%d", orch.Config.MaxPort),
		fmt.Sprintf("PLUGIN_ADDRESS=%s", orch.Config.Address),
	}
	env = append(os.Environ(), env...)
	return env
}

func (orch *Orchestrator) getListener() (net.Listener, int, error) {
	for port := orch.Config.MinPort; port <= orch.Config.MaxPort; port++ {
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
	Address string
	MinPort int
	MaxPort int
	Plugins []string
}

// Print() dumps the orchestrator configuration to string
func (c *OrchestratorConfiguration) Print() string {
	return fmt.Sprintf("\nCONFIGURATION\n    Address:   %v\n    MinPort:   %v\n    MaxPort:   %v\n    Plugins:   %v", c.Address, c.MinPort, c.MaxPort, c.Plugins)
}
