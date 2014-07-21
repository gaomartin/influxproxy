// Package orchestrator provides an infrastucture that allows the program to communicate
// with plugin software that make use of the 'plugin' package (github.com/influxproxy/influxproxy/plugin).
// The package is developed for the purpose and requirements of 'influxproxy'
// (github.com/influxproxy/influxproxy), nonetheless the conceplt is kept abstract so with small changes,
// the code should be reusable for other prjects.
//
// The concept is relatively simple and is based on the concept of the plugin infrastructure of 'packer.io':
// The orchestrator launches external executables (eg. the plugins) and provides basic configuration
// infromation via their environment variables. Information on the plugins are kept by plugin brokers. Plugin
// brokers are registerd in the registry 'owned' by the orchestrator.
//
// As soon as the plugins are lauchned, they call the RPC interface of the orchestrator (provided via the
// connector of the orchestrator) to share their connection details. The orchestrator then connects via RPC
// to the plugins. When the connection is established via 'handshake', the program can invoke the
// functionality of the plugins via Orchestrator > Registry > Brokers
package orchestrator

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
)

const (
	localhost = "127.0.0.1"
)

// ---------------------------------------------------------------------------------
// Orchestrator
// ---------------------------------------------------------------------------------

// Orchestrator is used to orchestrate the plugins, manage their life cycle and handle
// the communication between the orchestrating programm and its plugins
type Orchestrator struct {
	Config    *OrchestratorConfiguration // holds all necessary configuration
	Registry  *BrokerRegistry            // holds all plugin broker information
	Connector *Connector                 // holds the methodes that are exposed via RPC
	Port      int                        // holds its own port that is exposed via RPC
}

// NewOrchestrator returnd a fully initialized orchestrator and registres the given
// plugins to its registry.
func NewOrchestrator(conf *OrchestratorConfiguration) (*Orchestrator, error) {
	if conf.PluginMaxPort == 0 || conf.PluginMinPort == 0 {
		return nil, errors.New("Insufficent orchestrator configuration")
	}
	var out string
	var err error

	o := &Orchestrator{
		Config: conf,
	}

	o.Registry = NewBrokerRegistry()

	for _, plugin := range o.Config.Plugins {
		err = o.Registry.RegisterBroker(plugin)
		if err != nil {
			out += err.Error()
		}
	}

	if err != nil {
		err = errors.New(out)
	}
	return o, err
}

// Starts the orchestrator instance and all its Plugins.
func (orch *Orchestrator) Start() ([]string, error) {
	var err error
	var messages []string

	// Get the orchestrator itself started. Since spinup() lives forever in order to
	// serve the RPC connection, it is starded in a goroutine and continues unblocks
	// as soon as orchChan recieves any value.
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

	// Get plugins started via their brokers. Since Spinup() lasts as long as the plugin
	// does not crash, a goroutine and continues unblocks as soon as bChan recieves any value.
	bChan := make(chan bool)
	go func() {
		for _, b := range *orch.Registry {
			err = b.Spinup(orch)
			if err != nil {
				messages = append(messages, fmt.Sprintf("Plugin %s could not be loaded: %s. ", b.Name, err))
			} else {
				messages = append(messages, fmt.Sprintf("Plugin %s successfully loaded. ", b.Name))
			}
		}
		bChan <- true
	}()
	<-bChan
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
		fmt.Sprintf("ORCHESTRATOR_CONN_STRING=%s:%v", localhost, orch.Port),
		fmt.Sprintf("PLUGIN_MIN_PORT=%d", orch.Config.PluginMinPort),
		fmt.Sprintf("PLUGIN_MAX_PORT=%d", orch.Config.PluginMaxPort),
	}
	env = append(os.Environ(), env...)
	return env
}

func (orch *Orchestrator) getListener() (net.Listener, int, error) {
	for port := orch.Config.PluginMinPort; port <= orch.Config.PluginMaxPort; port++ {
		connection := fmt.Sprintf("%s:%v", localhost, port)
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
	PluginMinPort int
	PluginMaxPort int
	Plugins       []string
}
