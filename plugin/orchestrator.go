package plugin

import (
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
)

// ---------------------------------------------------------------------------------
// Orchestrator
// ---------------------------------------------------------------------------------

//
type Orchestrator struct {
	Config    *OrchestratorConfiguration
	Registry  *PluginRegistry
	Conductor *Conductor
	Port      int
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

// Starts the orchestrator instance and all its Plugins concurrently
func (orch *Orchestrator) Start() error {
	// get the orchestrator itself started
	orchChan := make(chan bool)
	go orch.spinupOrchestrator(orchChan)
	<-orchChan

	// get plugins started concurrently
	pluginChan := make(chan bool)
	go func() {
		for _, plugin := range orch.Registry.GetPlugins() {
			orch.spinupPlugin(&plugin)
		}
		pluginChan <- true
	}()
	<-pluginChan

	return nil
}

// Inits the Orchestrator
func (orch *Orchestrator) spinupOrchestrator(c chan bool) error {
	conductorChan := make(chan *ConductorData)
	conductor, err := NewConductor(conductorChan)
	if err != nil {
		return err
	}
	orch.Conductor = conductor

	rpc.Register(orch.Conductor)

	ln, port, err := getListener(orch.Config.MinPort, orch.Config.MaxPort, orch.Config.Address)
	if err != nil {
		return err
	}

	orch.Port = port

	c <- true
	out := fmt.Sprintf("Orchestrator is listening on port %v", orch.Port)
	fmt.Println(out)
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(c)
	}

	return nil
}

// Maintains the start process of a plugin
func (orch *Orchestrator) spinupPlugin(p *PluginRegister) error {
	c := make(chan bool)
	go orch.launchPlugin(p, c)
	<-c
	// orch.plugins.registerPlugin(p)
	// orch.waitForPlugin(p)
	// orch.salutePlugin(p)
	return nil
}

// Launch the plugin binary
func (orch *Orchestrator) launchPlugin(p *PluginRegister, c chan bool) error {
	fmt.Println("launchPlugin " + p.Name)
	cmd := exec.Command(p.Name)
	cmd.Env = append(cmd.Env, orch.getEnv()...)
	err := cmd.Start()
	if err != nil {
		return nil
	}

	p.Status.Started = true

	c <- true

	defer func() {
		r := recover()
		if err != nil || r != nil {
			cmd.Process.Kill()
		}
		if r != nil {
			panic(r)
		}
	}()

	exitCh := make(chan struct{})
	go func() {
		cmd.Wait()
		fmt.Println("Failed: " + p.Name)
		p.Status.Failed = true
		close(exitCh)
	}()

	return nil
}

func (orch *Orchestrator) getEnv() []string {
	env := []string{
		fmt.Sprintf("ORCHESTRATOR_CONN_STRING=%s:%v", orch.Config.Address, orch.Port),
		fmt.Sprintf("PLUGIN_MIN_PORT=%d", orch.Config.MinPort),
		fmt.Sprintf("PLUGIN_MAX_PORT=%d", orch.Config.MaxPort),
	}
	env = append(os.Environ(), env...)
	return env
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

// Print() dumps the orchestrator configuration to StdOut
func (conf *OrchestratorConfiguration) Print() string {
	return fmt.Sprintf("%+v\n", *conf)
}
