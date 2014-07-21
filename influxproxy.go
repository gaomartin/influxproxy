package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/influxproxy/influxproxy/orchestrator"
)

func main() {
	conf := NewConfiguration("INFLUXPROXY")

	influxdbs := NewDbs(conf.Influxdb)

	o, err := orchestrator.NewOrchestrator(conf.Orchestrator)
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
			c.String(handleGetPlugin(c, o))
		})

		in.POST("/:db/:plugin", func(c *gin.Context) {
			c.String(handlePostPlugin(c, o, influxdbs))
		})
	}

	admin := g.Group("/admin")
	{
		admin.GET("/brokers", func(c *gin.Context) {
			c.String(handleGetBrokers(c, o))
		})

		admin.GET("/config", func(c *gin.Context) {
			c.String(handleGetConfig(c, conf))
		})
	}

	echo := g.Group("/echo")
	{
		echo.POST("/:plugin", func(c *gin.Context) {
			c.String(handleEchoPlugin(c, o))
		})
	}

	g.Run(conf.Proxy.Host)
}
