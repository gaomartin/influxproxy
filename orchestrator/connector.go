package orchestrator

import (
	"errors"
	"fmt"
	"net/rpc"

	"github.com/influxproxy/influxproxy/plugin"
)

// ---------------------------------------------------------------------------------
// Connector
// ---------------------------------------------------------------------------------

// Connector provides all functionality that is exposed via RPC to recieve messages from
// the plugins.
type Connector struct {
	Registry *BrokerRegistry
}

// NewConnector returns an initialized connector.
func NewConnector(reg *BrokerRegistry) *Connector {
	o := &Connector{
		Registry: reg,
	}
	return o
}

// Handshake is exposed via RPC. Every plugin needs to call this method.
// The plugin fingerprint identifies the plugin and allows the connector
// to find its relevant broker. Only if the handshake succeeded, the
// plugin is considered 'connected' and accessable for the orchestrator.
// It also adds the RPC client to the plugin broker.
func (c *Connector) Handshake(p plugin.Fingerprint, ok *bool) error {
	b := c.Registry.GetBrokerByName(p.Name)
	if b == nil {
		*ok = false
		return errors.New("Plugin broker not found for " + p.Name)
	}
	b.Port = p.Port
	b.Status.State = Handshaked
	client, err := c.connect(b)
	if err != nil {
		*ok = false
		return err
	}
	b.client = client
	b.Status.State = Connected
	ping, err := b.Ping()
	*ok = ping
	if err != nil {
		return errors.New("Plugin could not be pinged")
	}

	b.readyChan <- true // this unblocks the Spinup of the broker
	return nil
}

// Ping is exposed via RPC and used by the plugin to detect if the orchestrator
// program has exited regularly (if the orchestrator is killed, plugins will
// killed as well).
func (c *Connector) Ping(in []*interface{}, pong *bool) error {
	*pong = true
	return nil
}

// connect gets the RPC client required to talk to the plugins.
func (c *Connector) connect(b *PluginBroker) (*rpc.Client, error) {
	connStr := fmt.Sprintf("%s:%v", localhost, b.Port)
	client, err := rpc.Dial("tcp", connStr)
	if err != nil {
		return nil, err
	}
	return client, nil
}
