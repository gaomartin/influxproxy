package main

import (
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
			reply := p.Name //, _ := p.Describe()
			c.String(200, reply)
		} else {
			c.String(404, c.Params.ByName("plugin")+" does not exist")
		}
	})

	g.GET("/plugins", func(c *gin.Context) {
		c.String(200, o.Registry.Print())
	})

	g.GET("/config", func(c *gin.Context) {
		c.String(200, o.Config.Print())
	})

	g.Run(":8080")
}
