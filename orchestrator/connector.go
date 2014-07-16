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
	c        chan *ConnectorData
	Registry *PluginRegistry
}

// ConductorData is used to communicate with the orchestrator,
// responses can be recieved via channel r
type ConnectorData struct {
	r chan bool
}

// returns an orchestrator interface with a given channel
func NewConnector(c chan *ConnectorData, reg *PluginRegistry) (*Connector, error) {
	if c == nil {
		return nil, errors.New("No channel specified")
	} else {
		o := &Connector{
			c:        c,
			Registry: reg,
		}
		return o, nil
	}
}

func (c *Connector) Handshake(plugin plugin.Fingerprint, ok *bool) error {
	p := c.Registry.GetPluginByName(plugin.Name)
	if p == nil {
		*ok = false
		return errors.New("Plugin not found")
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
