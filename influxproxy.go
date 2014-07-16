package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	orch "github.com/influxproxy/influxproxy/orchestrator"
)

func getConf(prefix string) *orch.OrchestratorConfiguration {
	if prefix != "" {
		prefix = prefix + "_"
	}

	minport, _ := strconv.Atoi(os.Getenv(prefix + "MINPORT"))
	maxport, _ := strconv.Atoi(os.Getenv(prefix + "MAXPORT"))

	config := orch.OrchestratorConfiguration{
		Address: os.Getenv(prefix + "ADDRESS"),
		MinPort: minport,
		MaxPort: maxport,
		Plugins: strings.Split(os.Getenv(prefix+"PLUGINS"), ","),
	}

	return &config
}

func getBodyAsString(body io.ReadCloser) (string, error) {
	out, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func main() {
	c := getConf("INFLUXPROXY")
	o, err := orch.NewOrchestrator(c)
	if err != nil {
		log.Panic(err)
	}

	messages, err := o.Start()
	for _, message := range messages {
		log.Println(message)
	}
	if err != nil {
		log.Panic(err)
	}

	g := gin.Default()

	g.GET("in/:db/:queue/:plugin", func(c *gin.Context) {
		p := o.Registry.GetPluginByName(c.Params.ByName("plugin"))
		if p != nil {
			reply, err := p.Describe()
			//TODO: Nice error handling
			b, err := json.Marshal(reply)
			if err == nil {
				c.String(200, string(b))
			} else {
				c.String(500, err.Error())
			}
		} else {
			c.String(404, c.Params.ByName("plugin")+" does not exist")
		}
	})

	g.POST("in/:db/:queue/:plugin", func(c *gin.Context) {
		p := o.Registry.GetPluginByName(c.Params.ByName("plugin"))
		if p != nil {
			call, err := getBodyAsString(c.Req.Body)
			reply, err := p.Run(call)
			b, err := json.Marshal(reply)
			if err == nil {
				c.String(200, string(b))
			} else {
				c.String(500, err.Error())
			}
		} else {
			c.String(404, c.Params.ByName("plugin")+" does not exist")
		}
	})
	g.GET("/plugins", func(c *gin.Context) {
		b, err := json.Marshal(o.Registry)
		if err == nil {
			c.String(200, string(b))
		} else {
			c.String(500, err.Error())
		}
	})

	g.GET("/config", func(c *gin.Context) {
		b, err := json.Marshal(o.Config)
		if err == nil {
			c.String(200, string(b))
		} else {
			c.String(500, err.Error())
		}
	})

	g.Run(":8080")
}
