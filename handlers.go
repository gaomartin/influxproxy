package main

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/influxproxy/influxproxy/orchestrator"
	"github.com/influxproxy/influxproxy/plugin"
)

func handleGetPlugin(c *gin.Context, o *orchestrator.Orchestrator) (int, string) {
	b := o.Registry.GetBrokerByName(c.Params.ByName("plugin"))
	if b != nil {
		reply, err := b.Describe()
		if err != nil {
			return 500, err.Error()
		}

		text, err := json.Marshal(reply)
		if err != nil {
			return 500, err.Error()
		} else {
			return 200, string(text)
		}
	} else {
		return 404, c.Params.ByName("plugin") + " does not exist"
	}
}

func handleEchoPlugin(c *gin.Context, o *orchestrator.Orchestrator) (int, string) {
	b := o.Registry.GetBrokerByName(c.Params.ByName("plugin"))
	if b != nil {
		body, err := getBodyAsString(c.Req.Body)
		if err != nil {
			return 500, err.Error()
		}

		query := c.Req.URL.Query()
		call := plugin.Request{
			Query: query,
			Body:  body,
		}

		reply, err := b.Run(call)
		if err != nil {
			return 500, err.Error()
		} else if reply.Error != "" {
			return 500, reply.Error
		}

		b, err := json.Marshal(reply.Series)
		if err != nil {
			return 500, err.Error()
		} else {
			return 200, string(b)
		}
	} else {
		return 404, c.Params.ByName("plugin") + " does not exist"
	}
}

func handlePostPlugin(c *gin.Context, o *orchestrator.Orchestrator, influxdbs *Dbs) (int, string) {
	b := o.Registry.GetBrokerByName(c.Params.ByName("plugin"))
	if b != nil {
		db, err := influxdbs.Get(c.Params.ByName("db"))
		if err != nil {
			return 500, err.Error()
		}

		body, err := getBodyAsString(c.Req.Body)
		if err != nil {
			return 500, err.Error()
		}

		query := c.Req.URL.Query()
		call := plugin.Request{
			Query: query,
			Body:  body,
		}
		reply, err := b.Run(call)
		if err != nil {
			return 500, err.Error()
		} else if reply.Error != "" {
			return 500, reply.Error
		}

		err = db.WriteSeries(reply.Series)
		if err != nil {
			return 500, err.Error()
		} else {
			return 200, "Series are written to InfluxDB"
		}
	} else {
		return 404, c.Params.ByName("plugin") + " does not exist"
	}
}

func handleGetBrokers(c *gin.Context, o *orchestrator.Orchestrator) (int, string) {
	b, err := json.Marshal(o.Registry)
	if err == nil {
		return 200, string(b)
	} else {
		return 500, err.Error()
	}
}

func handleGetConfig(c *gin.Context, conf *Configuration) (int, string) {
	b, err := json.Marshal(conf)
	if err == nil {
		return 200, string(b)
	} else {
		return 500, err.Error()
	}
}

// ---------------------------------------------------------------------------------
// Helper Functions
// ---------------------------------------------------------------------------------

func getBodyAsString(body io.ReadCloser) (string, error) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}
	out := string(b)
	return out, nil
}
