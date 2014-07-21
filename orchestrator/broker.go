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

// PluginBroker holds information about the plugin itself, its state
// The broker also manages the life cycle of the plugin.
type PluginBroker struct {
	Name      string        // name of the plugin
	Plugin    string        // file system path of the plugin
	Port      int           // port of the RPC server of the plugin
	readyChan chan bool     // channel that is used in order to get the connected state from the connector
	client    *rpc.Client   // RPC client used to access the functionality of the plugin
	Status    *PluginStatus // status of the plugin
}

// NewPluginBroker return an initialized plugin broker of a not yet started plugin.
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

// Maintains the start process of a plugin.
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

// Ping calles the plugin. Its only purpose is to ensure that the plugin is alive
// and responding.
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

// Describe requests information of the plugin and returns them to the caller.
// The returned plugin.Description provides detailed information on the
// funtionality and the arguments of the plugin.
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

// Run invoces the main functionality of the plugin.
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

// launch starts the plugin binary with its configuration as environment and runs for
// the whole life of the plugin. It also initiates the cleanup if the plugin panics or
// fails for any reason.
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

// fail makes shure that the state of the plugin is reset and channels are unblocked.
func (b *PluginBroker) fail(c chan error, err error) {
	b.reset()
	b.Status.FailCount += 1
	c <- err
	b.readyChan <- false
}

// watch keeps track of the plugin and cleans up if the plugin dies for any reason.
func (b *PluginBroker) watch(c chan error, cmd *exec.Cmd, exitCh chan struct{}) {
	cmd.Wait()
	err := errors.New("Plugin ended")
	b.fail(c, err)
	close(exitCh)
}

// cleanup makes shure that any plugin process is cleaned up.
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

// reset resets the state of a plugin.
func (b *PluginBroker) reset() {
	b.Port = 0
	b.client = nil
	b.Status.State = None
}

// ---------------------------------------------------------------------------------
// PluginStatus
// ---------------------------------------------------------------------------------

// PluginStatus holds relevant information on the state of the plugin resp. the broker.
type PluginStatus struct {
	State     State  // current state of the plugin
	FailCount uint32 // number of crashes of the plugin
	RunCount  uint32 // number of Run() calls of the plugin
}

// State is the representation of the plugin state.
type State int

const (
	None State = iota + 1
	Started
	Handshaked
	Connected
)

// String implements the Stringer interface and returns an textual representation of the state.
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
