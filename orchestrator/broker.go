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
	Name      string
	Path      string
	Address   string
	Port      int
	ReadyChan chan bool
	Client    *rpc.Client
	Status    *PluginStatus
}

func NewPluginBroker(name string, path string, address string) (*PluginBroker, error) {
	s := &PluginStatus{
		Started:    false,
		Handshaked: false,
		Connected:  false,
		FailCount:  0,
	}

	c := make(chan bool)

	p := &PluginBroker{
		Name:      name,
		Path:      path,
		Address:   address,
		Port:      0,
		ReadyChan: c,
		Client:    nil,
		Status:    s,
	}

	return p, nil
}

// Maintains the start process of a plugin
func (p *PluginBroker) Spinup(orch *Orchestrator) error {
	c := make(chan error)
	go p.launch(c, orch)
	err := <-c
	<-p.ReadyChan

	if err != nil {
		return err
	}
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
func (p *PluginBroker) launch(c chan error, orch *Orchestrator) {

	if _, err := os.Stat(p.Path); os.IsNotExist(err) {
		p.fail(c, err)
		return
	}

	cmd := exec.Command(p.Path)
	cmd.Env = append(cmd.Env, orch.getEnv()...)
	err := cmd.Start()
	if err != nil {
		p.fail(c, err)
		return
	}

	p.Status.Started = true

	defer p.cleanup(c, err, cmd)

	exitCh := make(chan struct{})
	go p.watch(c, cmd, exitCh)
}

func (p *PluginBroker) fail(c chan error, err error) {
	p.reset()
	c <- err
	p.ReadyChan <- false
}

func (p *PluginBroker) watch(c chan error, cmd *exec.Cmd, exitCh chan struct{}) {
	cmd.Wait()
	err := errors.New("Plugin ended")
	p.fail(c, err)
	close(exitCh)
}

func (p *PluginBroker) cleanup(c chan error, err error, cmd *exec.Cmd) {
	c <- nil
	r := recover()
	if err != nil || r != nil {
		cmd.Process.Kill()
	}
	if r != nil {
		panic(r)
	}
}

func (p *PluginBroker) reset() {
	p.Port = 0
	p.Client = nil
	p.Status.Started = false
	p.Status.Handshaked = false
	p.Status.Connected = false
	p.Status.FailCount += 1
}

// ---------------------------------------------------------------------------------
// PluginStatus
// ---------------------------------------------------------------------------------

type PluginStatus struct {
	Started    bool
	Handshaked bool
	Connected  bool
	FailCount  int
}

func (s *PluginStatus) Print() string {
	return fmt.Sprintf("\n       Started:    %v\n       Handshaked: %v\n       Connected:  %v\n       Failed:     %v\n", s.Started, s.Handshaked, s.Connected, s.FailCount)
}
