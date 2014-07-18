package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/influxdb/influxdb-go"
	"github.com/influxproxy/influxproxy/orchestrator"
)

type Configuration struct {
	Orchestrator *orchestrator.OrchestratorConfiguration
	Influxdb     *influxdb.ClientConfig
	Proxy        string
}

func NewConfiguration(prefix string) *Configuration {
	if prefix != "" {
		prefix = prefix + "_"
	}

	minport, _ := strconv.Atoi(os.Getenv(prefix + "PLUGIN_MINPORT"))
	maxport, _ := strconv.Atoi(os.Getenv(prefix + "PLUGIN_MAXPORT"))

	orch := &orchestrator.OrchestratorConfiguration{
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
