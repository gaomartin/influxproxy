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
	orch "github.com/influxproxy/influxproxy/orchestrator"
	"github.com/influxproxy/influxproxy/plugin"
)

func getOrchestratorConf(prefix string) *orch.OrchestratorConfiguration {
	if prefix != "" {
		prefix = prefix + "_"
	}

	minport, _ := strconv.Atoi(os.Getenv(prefix + "PLUGINMINPORT"))
	maxport, _ := strconv.Atoi(os.Getenv(prefix + "PLUGINMAXPORT"))

	config := orch.OrchestratorConfiguration{
		Address:       os.Getenv(prefix + "ADDRESS"),
		PluginMinPort: minport,
		PluginMaxPort: maxport,
		Plugins:       strings.Split(os.Getenv(prefix+"PLUGINS"), ","),
	}

	return &config
}

func getConf(prefix string) string {
	if prefix != "" {
		prefix = prefix + "_"
	}
	connStr := fmt.Sprintf("%s:%s", os.Getenv(prefix+"ADDRESS"), os.Getenv(prefix+"PORT"))
	return connStr
}

func getBodyAsString(body io.ReadCloser) (string, error) {
	out, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func main() {
	c := getOrchestratorConf("INFLUXPROXY")
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

	g.Run(getConf("INFLUXPROXY"))
}
