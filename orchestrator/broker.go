package orchestrator

import (
	"errors"
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
)

// ---------------------------------------------------------------------------------
// PluginBroker
// ---------------------------------------------------------------------------------

// The broker holds information about the plugin itself, its state
// The broker also manages the life cycle of the plugin
type PluginBroker struct {
	Name    string
	Path    string
	Address string
	Port    int
	Client  *rpc.Client
	Status  *PluginStatus
}

func NewPluginBroker(name string, path string, address string) (*PluginBroker, error) {
	s := &PluginStatus{
		Started:    false,
		Handshaked: false,
		Connected:  false,
		Failed:     false,
	}

	p := &PluginBroker{
		Name:    name,
		Path:    path,
		Address: address,
		Port:    0,
		Client:  nil,
		Status:  s,
	}

	return p, nil
}

// Maintains the start process of a plugin
func (p *PluginBroker) Spinup(orch *Orchestrator) error {
	c := make(chan bool)
	go p.launch(c, orch)
	<-c
	// TODO: waiting for Plugins to connect should time out and set status to "failed=true"

	// orch.waitForPlugin(p)
	// orch.salutePlugin(p)
	fmt.Println("Done:   " + p.Name)
	return nil
}

func (p *PluginBroker) Describe() (string, error) {
	if !p.Status.Connected {
		return "", errors.New("Plugin not yet connected")
	}
	var reply string
	call := new([]interface{})
	err := p.Client.Call("Connector.Describe", *call, &reply)
	if err != nil {
		return "", err
	}
	return reply, nil
}

func (p *PluginBroker) Ping() (bool, error) {
	if !p.Status.Connected {
		return false, errors.New("Plugin not yet connected")
	}
	var reply bool
	call := new([]interface{})
	err := p.Client.Call("Connector.Ping", *call, &reply)
	if err != nil {
		return false, err
	}
	return reply, nil
}

// Launch the plugin binary
func (p *PluginBroker) launch(c chan bool, orch *Orchestrator) error {

	fmt.Println("Launch: " + p.Name)

	if _, err := os.Stat(p.Path); os.IsNotExist(err) {
		p.Status.Failed = true
		return err
	}

	cmd := exec.Command(p.Path)
	cmd.Env = append(cmd.Env, orch.getEnv()...)
	err := cmd.Start()
	if err != nil {
		return err
	}

	p.Status.Started = true

	defer p.cleanup(c, err, cmd)

	exitCh := make(chan struct{})
	go p.watch(cmd, exitCh)

	return nil
}

func (p *PluginBroker) watch(cmd *exec.Cmd, exitCh chan struct{}) {
	cmd.Wait()
	fmt.Println("Failed: " + cmd.Path)
	p.Status.Failed = true
	close(exitCh)
}

func (p *PluginBroker) cleanup(c chan bool, err error, cmd *exec.Cmd) {
	c <- true
	r := recover()
	if err != nil || r != nil {
		cmd.Process.Kill()
	}
	if r != nil {
		panic(r)
	}
}

// ---------------------------------------------------------------------------------
// PluginStatus
// ---------------------------------------------------------------------------------

type PluginStatus struct {
	Started    bool
	Handshaked bool
	Connected  bool
	Failed     bool
}

func (s *PluginStatus) Print() string {
	return fmt.Sprintf("\n       Started:    %v\n       Handshaked: %v\n       Connected:  %v\n       Failed:     %v\n", s.Started, s.Handshaked, s.Connected, s.Failed)
}
