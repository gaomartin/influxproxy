package plugin

import (
	"errors"
	"net/url"

	"github.com/influxdb/influxdb-go"
)

type Connector struct {
	e Exposer
}

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

func (c *Connector) Ping(in []*interface{}, pong *bool) error {
	*pong = true
	return nil
}

func (c *Connector) Describe(in []*interface{}, description *Description) error {
	*description = c.e.Describe()
	return nil
}

func (c *Connector) Run(in Request, out *Response) error {
	*out = c.e.Run(in)
	return nil
}

type Exposer interface {
	Describe() Description
	Run(in Request) Response
}

type Request struct {
	Query url.Values
	Body  string
}

type Response struct {
	Series []*influxdb.Series
	Error  string
}
