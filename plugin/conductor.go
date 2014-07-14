package plugin

import (
	"errors"
)

// The interface that is exposed via rpc by the orchestrator. the channel
// is used to send data to the orchestrator
type Conductor struct {
	c chan *ConductorData
}

// ConductorData is used to communicate with the orchestrator,
// responses can be recieved via channel r
type ConductorData struct {
	r chan bool
}

// returns an orchestrator interface with a given channel
func NewConductor(c chan *ConductorData) (*Conductor, error) {
	if c == nil {
		return nil, errors.New("no channel specified")
	} else {
		o := &Conductor{
			c: c,
		}
		return o, nil
	}
}
