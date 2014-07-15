package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	//"github.com/influxproxy/influxproxy/plugin"
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
	o, err := orch.NewOrchestrator(getConf("ORCH"))
	if err != nil {
		log.Print(err)
	}
	log.Print(o.Config.Print())
	o.Start()

	// wait a bit
	var input string
	fmt.Scanln(&input)

	//TODO: Spawned processes will life forever, needs some cleanup
	log.Print(o.Registry.Print())
}
