package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/influxdb/influxdb-go"
	orch "github.com/influxproxy/influxproxy/orchestrator"
	"github.com/influxproxy/influxproxy/plugin"
)

type Configuration struct {
	Orchestrator *orch.OrchestratorConfiguration
	Influxdb     *influxdb.ClientConfig
	Proxy        string
}

func GetConf(prefix string) *Configuration {
	if prefix != "" {
		prefix = prefix + "_"
	}

	minport, _ := strconv.Atoi(os.Getenv(prefix + "PLUGIN_MINPORT"))
	maxport, _ := strconv.Atoi(os.Getenv(prefix + "PLUGIN_MAXPORT"))

	orch := &orch.OrchestratorConfiguration{
		Address:       os.Getenv(prefix + "PLUGIN_ADDRESS"),
		PluginMinPort: minport,
		PluginMaxPort: maxport,
		Plugins:       strings.Split(os.Getenv(prefix+"PLUGINS"), ","),
	}

	db := &influxdb.ClientConfig{
		Username: os.Getenv(prefix + "DB_USER"),
		Password: os.Getenv(prefix + "DB_PASSWORD"),
		Database: "",
		Host:     os.Getenv(prefix+"DB_ADDRESS") + ":" + os.Getenv(prefix+"DB_PORT"),
	}

	config := &Configuration{
		Orchestrator: orch,
		Influxdb:     db,
		Proxy:        os.Getenv(prefix+"ADDRESS") + ":" + os.Getenv(prefix+"PORT"),
	}

	return config

}

func getBodyAsString(body io.ReadCloser) (string, error) {
	out, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func main() {
	c := GetConf("INFLUXPROXY")
	o, err := orch.NewOrchestrator(c.Orchestrator)
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

	in := g.Group("/in")
	{

		in.GET("/:db/:plugin", func(c *gin.Context) {
			b := o.Registry.GetBrokerByName(c.Params.ByName("plugin"))
			if b != nil {
				reply, err := b.Describe()
				//TODO: Nice error handling
				text, err := json.Marshal(reply)
				if err == nil {
					c.String(200, string(text))
				} else {
					c.String(500, err.Error())
				}
			} else {
				c.String(404, c.Params.ByName("plugin")+" does not exist")
			}
		})

		in.POST("/:db/:plugin", func(c *gin.Context) {
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
					c.String(500, err.Error())
				} else if reply.Error != "" {
					c.String(500, reply.Error)
				} else {
					c.String(200, string(series))
				}
			} else {
				c.String(404, c.Params.ByName("plugin")+" does not exist")
			}
		})
	}

	admin := g.Group("/admin")
	{
		admin.GET("/brokers", func(c *gin.Context) {
			b, err := json.Marshal(o.Registry)
			if err == nil {
				c.String(200, string(b))
			} else {
				c.String(500, err.Error())
			}
		})

		admin.GET("/config", func(c *gin.Context) {
			b, err := json.Marshal(o.Config)
			if err == nil {
				c.String(200, string(b))
			} else {
				c.String(500, err.Error())
			}
		})
	}

	g.Run(c.Proxy)
}
