package orchestrator

import (
	"errors"
	"fmt"
	"net/rpc"

	"github.com/influxproxy/influxproxy/plugin"
)

// The interface that is exposed via rpc by the orchestrator. the channel
// c is used to send data to the orchestrator
type Connector struct {
	Registry *PluginRegistry
}

// returns an orchestrator interface with a given channel
func NewConnector(reg *PluginRegistry) *Connector {
	o := &Connector{
		Registry: reg,
	}
	return o
}

func (c *Connector) Handshake(plugin plugin.Fingerprint, ok *bool) error {
	p := c.Registry.GetPluginByName(plugin.Name)
	if p == nil {
		*ok = false
		return errors.New("Plugin broker not found for " + plugin.Name)
	}
	p.Port = plugin.Port
	p.Status.Handshaked = true
	client, err := c.connect(p)
	if err != nil {
		*ok = false
		return err
	}
	p.Client = client
	p.Status.Connected = true
	ping, err := p.Ping()
	*ok = ping
	if err != nil {
		return errors.New("Plugin could not be pinged")
	}
	p.ReadyChan <- true
	return nil
}

func (c *Connector) connect(p *PluginBroker) (*rpc.Client, error) {
	connStr := fmt.Sprintf("%s:%v", p.Address, p.Port)
	client, err := rpc.Dial("tcp", connStr)
	if err != nil {
		return nil, err
	}
	return client, nil
}
