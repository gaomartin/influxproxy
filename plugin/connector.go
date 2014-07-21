package plugin

import (
	"errors"
	"net/url"

	"github.com/influxdb/influxdb-go"
)

// ---------------------------------------------------------------------------------
// Connector
// ---------------------------------------------------------------------------------

// Connector provides all functionality that is exposed via RPC to recieve messages from
// the orchestrator.
type Connector struct {
	e Exposer
}

// NewConnector returns a fully initialiyed Connector. It requires anything that implements
// the Exposer interface.
func NewConnector(e Exposer) (*Connector, error) {
	if e == nil {
		return nil, errors.New("No exposer provided")
	} else {
		c := &Connector{
			e: e,
		}
		return c, nil
	}
}

// Ping allows the orchestrator to check if the plugin is responing on RPC calls.
func (c *Connector) Ping(in []*interface{}, pong *bool) error {
	*pong = true
	return nil
}

// Describe returns a detailed desciption of the plugin to the orchestrator.
func (c *Connector) Describe(in []*interface{}, description *Description) error {
	*description = c.e.Describe()
	return nil
}

// Run invokes the main functionality provided by the plugin.
func (c *Connector) Run(in Request, out *Response) error {
	*out = c.e.Run(in)
	return nil
}

// ---------------------------------------------------------------------------------
// Exposer
// ---------------------------------------------------------------------------------

// Exposer needs to be implemented by the plugin program itself in order to provide
// the required functionalities required by the orchestrator. As long as no more
// functionality is required, this can be reused in other projecs than InfluxProxy.
// If more/other functionality is required, the Exposer interface, the plugin.Connector
// methodes as well as the orchestrator.PluginBroker methodes would be required to be
// remodeled
type Exposer interface {
	Describe() Description
	Run(in Request) Response
}

// ---------------------------------------------------------------------------------
// Request
// ---------------------------------------------------------------------------------

// Request contains all information that needs to be shipped from the orchestrator
// to the plugin in order to execute its main functionality via RPC Run function.
// Since this is specific to InfluxProxy, this needs to be changed on case of
// alternative use in other projects.
type Request struct {
	Query url.Values
	Body  string
}

// ---------------------------------------------------------------------------------
// Response
// ---------------------------------------------------------------------------------

// Response describes the data that is send back to the orchestrator. Since this
// is specific to InfluxProxy, this needs to be changed on case of alternative use in
// other projects.
type Response struct {
	Series []*influxdb.Series // []*influxdb.Series is specific to InfluxProxy.
	Error  string             // Errors cannot be sent back, therfore an error string is used.
}
