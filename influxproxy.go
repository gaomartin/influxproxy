package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/influxproxy/influxproxy/plugin"
)

func getConf(prefix string) *plugin.OrchestratorConfiguration {
	if prefix != "" {
		prefix = prefix + "_"
	}

	minport, _ := strconv.Atoi(os.Getenv(prefix + "MINPORT"))
	maxport, _ := strconv.Atoi(os.Getenv(prefix + "MAXPORT"))

	config := plugin.OrchestratorConfiguration{
		Address: os.Getenv(prefix + "ADDRESS"),
		MinPort: minport,
		MaxPort: maxport,
		Plugins: strings.Split(os.Getenv(prefix+"PLUGINS"), ","),
	}

	return &config
}

func main() {
	o, _ := plugin.NewOrchestrator(getConf("ORCH"))
	log.Print(o.Config.Print())
	o.Start()
	log.Print(o.Registry.Print())
}
