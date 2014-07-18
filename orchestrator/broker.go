package orchestrator

import (
	"errors"
	"net/rpc"
	"os"
	"os/exec"

	"github.com/influxproxy/influxproxy/plugin"
)

// ---------------------------------------------------------------------------------
// PluginBroker
// ---------------------------------------------------------------------------------

// The broker holds information about the plugin itself, its state
// The broker also manages the life cycle of the plugin
type PluginBroker struct {
	Name      string
	Plugin    string
	Port      int
	readyChan chan bool
	client    *rpc.Client
	Status    *PluginStatus
}

func NewPluginBroker(name string, plugin string) (*PluginBroker, error) {
	s := &PluginStatus{
		State:     None,
		FailCount: 0,
		RunCount:  0,
	}

	c := make(chan bool)

	b := &PluginBroker{
		Name:      name,
		Plugin:    plugin,
		Port:      0,
		readyChan: c,
		client:    nil,
		Status:    s,
	}

	return b, nil
}

// Maintains the start process of a plugin
func (b *PluginBroker) Spinup(orch *Orchestrator) error {
	c := make(chan error)
	go b.launch(c, orch)
	err := <-c
	<-b.readyChan

	if err != nil {
		return err
	}
	return nil
}

func (b *PluginBroker) Ping() (bool, error) {
	if b.Status.State != Connected {
		return false, errors.New("Plugin not connected")
	}
	var reply bool
	call := new([]interface{})
	err := b.client.Call("Connector.Ping", *call, &reply)
	if err != nil {
		return false, err
	}
	return reply, nil
}

func (b *PluginBroker) Describe() (*plugin.Description, error) {
	if b.Status.State != Connected {
		return nil, errors.New("Plugin not connected")
	}
	var reply *plugin.Description
	call := new([]interface{})
	err := b.client.Call("Connector.Describe", *call, &reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (b *PluginBroker) Run(data plugin.Request) (*plugin.Response, error) {
	var reply *plugin.Response
	if b.Status.State != Connected {
		return reply, errors.New("Plugin not connected")
	}

	err := b.client.Call("Connector.Run", data, &reply)
	if err != nil {
		return reply, err
	}
	b.Status.RunCount += 1
	return reply, nil
}

// Launch the plugin binary
func (b *PluginBroker) launch(c chan error, orch *Orchestrator) {

	if _, err := os.Stat(b.Plugin); os.IsNotExist(err) {
		b.fail(c, err)
		return
	}

	cmd := exec.Command(b.Plugin)
	cmd.Env = append(cmd.Env, orch.getEnv()...)
	err := cmd.Start()
	if err != nil {
		b.fail(c, err)
		return
	}

	b.Status.State = Started

	defer b.cleanup(c, err, cmd)

	exitCh := make(chan struct{})
	go b.watch(c, cmd, exitCh)
}

func (b *PluginBroker) fail(c chan error, err error) {
	b.reset()
	b.Status.FailCount += 1
	c <- err
	b.readyChan <- false
}

func (b *PluginBroker) watch(c chan error, cmd *exec.Cmd, exitCh chan struct{}) {
	cmd.Wait()
	err := errors.New("Plugin ended")
	b.fail(c, err)
	close(exitCh)
}

func (b *PluginBroker) cleanup(c chan error, err error, cmd *exec.Cmd) {
	c <- nil
	r := recover()
	if err != nil || r != nil {
		cmd.Process.Kill()
	}
	if r != nil {
		panic(r)
	}
}

func (b *PluginBroker) reset() {
	b.Port = 0
	b.client = nil
	b.Status.State = None
}

// ---------------------------------------------------------------------------------
// PluginStatus
// ---------------------------------------------------------------------------------

type PluginStatus struct {
	State     State
	FailCount uint32
	RunCount  uint32
}

type State int

const (
	None State = iota + 1
	Started
	Handshaked
	Connected
)

func (s State) String() string {
	switch s {
	case None:
		return "None"
	case Started:
		return "Started"
	case Handshaked:
		return "Handshaked"
	case Connected:
		return "Connected"
	default:
		return "Unknown"
	}
}
