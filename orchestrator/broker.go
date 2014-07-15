package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ---------------------------------------------------------------------------------
// PluginBroker
// ---------------------------------------------------------------------------------

// The broker holds information about the plugin itself, its state
// The broker also manages the life cycle of the plugin
type PluginBroker struct {
	Name   string
	Path   string
	Port   int
	Status *PluginStatus
}

func NewPluginBroker(name string, path string) (*PluginBroker, error) {
	s := &PluginStatus{
		Started:   false,
		Connected: false,
		Failed:    false,
	}

	p := &PluginBroker{
		Name:   name,
		Path:	path,
		Port:   0,
		Status: s,
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
	go p.watch(cmd, exitCh, orch)

	return nil
}

func (p *PluginBroker) watch(cmd *exec.Cmd, exitCh chan struct{}, orch *Orchestrator) {
	cmd.Wait()
	fmt.Println("Failed: " + cmd.Path)
	// TODO: Somehow, p *PluginBroker points to the wrong broker. func (orch *Orchestrator) Start()
	//       needs to pass the right thing into plugin of plugin.Spinup(orch).
	//       The line below is therfore workaround and needs to be cleaned up.
	p = orch.Registry.GetPluginByName(filepath.Base(cmd.Path))
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
	Started   bool
	Connected bool
	Failed    bool
}

func (s *PluginStatus) Print() string {
	return fmt.Sprintf("\n       Started:   %v\n       Connected: %v\n       Failed:    %v\n", s.Started, s.Connected, s.Failed)
}