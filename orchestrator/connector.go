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
	Registry *BrokerRegistry
}

// returns an orchestrator interface with a given channel
func NewConnector(reg *BrokerRegistry) *Connector {
	o := &Connector{
		Registry: reg,
	}
	return o
}

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
	b.readyChan <- true
	return nil
}

func (c *Connector) Ping(in []*interface{}, pong *bool) error {
	*pong = true
	return nil
}

func (c *Connector) connect(b *PluginBroker) (*rpc.Client, error) {
	connStr := fmt.Sprintf("%s:%v", localhost, b.Port)
	client, err := rpc.Dial("tcp", connStr)
	if err != nil {
		return nil, err
	}
	return client, nil
}
