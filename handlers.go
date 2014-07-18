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
		//TODO: Nice error handling
		text, err := json.Marshal(reply)
		if err == nil {
			return 200, string(text)
			return 200, string(text)
		} else {
			return 500, err.Error()
		}
	} else {
		return 404, c.Params.ByName("plugin") + " does not exist"
	}
}

func handlePostPlugin(c *gin.Context, o *orchestrator.Orchestrator) (int, string) {
	b := o.Registry.GetBrokerByName(c.Params.ByName("plugin"))
	if b != nil {
		body, err := getBodyAsString(c.Req.Body)
		query := c.Req.URL.Query()
		call := plugin.Request{
			Query: query,
			Body:  body,
		}
		reply, err := b.Run(call)
		series, err := json.Marshal(reply.Series)
		if err != nil {
			return 500, err.Error()
		} else if reply.Error != "" {
			return 500, reply.Error
		} else {
			return 200, string(series)
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
	out, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
