package plugin

import (
	"net/url"

	"github.com/influxdb/influxdb-go"
)

type Request struct {
	Query url.Values
	Body  string
}

type Response struct {
	Series []influxdb.Series
	Error  string
}
