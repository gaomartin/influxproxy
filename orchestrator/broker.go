package orchestrator

import (
	"errors"
	"net/rpc"
	"os"
	"os/exec"

	"github.com/influxdb/influxdb-go"
	"github.com/influxproxy/influxproxy/plugin"
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
	readyChan chan bool
	client    *rpc.Client
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

	b := &PluginBroker{
		Name:      name,
		Path:      path,
		Address:   address,
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
	if !b.Status.Connected {
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
	if !b.Status.Connected {
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

func (b *PluginBroker) Run(data string) (*[]influxdb.Series, error) {
	var reply *[]influxdb.Series
	if !b.Status.Connected {
		return reply, errors.New("Plugin not connected")
	}

	err := b.client.Call("Connector.Run", data, &reply)
	if err != nil {
		return reply, err
	}
	return reply, nil
}

// Launch the plugin binary
func (b *PluginBroker) launch(c chan error, orch *Orchestrator) {

	if _, err := os.Stat(b.Path); os.IsNotExist(err) {
		b.fail(c, err)
		return
	}

	cmd := exec.Command(b.Path)
	cmd.Env = append(cmd.Env, orch.getEnv()...)
	err := cmd.Start()
	if err != nil {
		b.fail(c, err)
		return
	}

	b.Status.Started = true

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
	b.Status.Started = false
	b.Status.Handshaked = false
	b.Status.Connected = false
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
