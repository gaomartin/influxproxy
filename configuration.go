package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/influxproxy/influxproxy/orchestrator"
)

type Influxdb struct {
	Username string
	Password string
	Host     string
}

type Proxy struct {
	Host string
}

type Configuration struct {
	Orchestrator *orchestrator.OrchestratorConfiguration
	Influxdb     *Influxdb
	Proxy        *Proxy
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

	db := &Influxdb{
		Username: os.Getenv(prefix + "DB_USER"),
		Password: os.Getenv(prefix + "DB_PASSWORD"),
		Host:     os.Getenv(prefix+"DB_ADDRESS") + ":" + os.Getenv(prefix+"DB_PORT"),
	}

	proxy := &Proxy{
		Host: os.Getenv(prefix+"ADDRESS") + ":" + os.Getenv(prefix+"PORT"),
	}

	config := &Configuration{
		Orchestrator: orch,
		Influxdb:     db,
		Proxy:        proxy,
	}

	return config

}
