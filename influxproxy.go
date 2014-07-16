package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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
		log.Print(message)
	}
	if err != nil {
		log.Panic(err)
	}
}
