package orchestrator

import (
	"errors"
	"github.com/influxproxy/influxproxy/plugin"
)

// The interface that is exposed via rpc by the orchestrator. the channel
// c is used to send data to the orchestrator
type Connector struct {
	c chan *ConnectorData
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
			c: c,
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
	p.Status.Connected = true
	*ok = true
	return nil
}
