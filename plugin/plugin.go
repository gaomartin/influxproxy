// Package plugin is the counterpart of the orchestrator (github.com/influxproxy/influxproxy/orchestrator).
// Its purpose is to provide the most simple interface for programmers
// to implement a program that fulfils all requirements given by the orchestrator
// progamm to be called as a plugin.
//
// For more insights of the concept, read the documentation of the orchestrator
// (github.com/influxproxy/influxproxy/orchestrator).
package plugin

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	localhost = "127.0.0.1"
)

// ---------------------------------------------------------------------------------
// Plugin
// ---------------------------------------------------------------------------------

// Plugin is the core of the client side plugin infrastructure. It hold all information
// and provides all functionality used by the cumtom plugin implementation itself.
type Plugin struct {
	Config      *PluginConfiguration
	Fingerprint *Fingerprint
	Client      *rpc.Client
}

// NewPlugin reads the required configuration from the environment and returns an
// initialized plugin.
func NewPlugin() (*Plugin, error) {
	max, _ := strconv.Atoi(os.Getenv("PLUGIN_MAX_PORT"))
	min, _ := strconv.Atoi(os.Getenv("PLUGIN_MIN_PORT"))
	connString := os.Getenv("ORCHESTRATOR_CONN_STRING")

	// The name of the plugin is the name of the binary. This allows
	// to copy a binary or use symlinks to run the same plugin multiple times.
	name := filepath.Base(os.Args[0])

	if max != 0 && min != 0 && connString != "" {
		conf := &PluginConfiguration{
			OrchConnString: connString,
			MaxPort:        max,
			MinPort:        min,
		}
		fp := &Fingerprint{
			Name: name,
		}

		p := &Plugin{
			Config:      conf,
			Fingerprint: fp,
		}
		return p, nil
	} else {
		return nil, errors.New("Not enough data to build plugin. Do not start Plugin from command line.")
	}
}

// getListener allocates a port from an given range dynamically and returns a listener
// if any port was available in this range.
func (p *Plugin) getListener() (net.Listener, int, error) {
	for port := p.Config.MinPort; port <= p.Config.MaxPort; port++ {
		connection := fmt.Sprintf("%s:%d", localhost, port)
		listener, err := net.Listen("tcp", connection)
		if err == nil {
			return listener, port, nil
		}
	}

	return nil, 0, errors.New("Could not get TCP listener, maybe all ports are already used")
}

// Run starts the plugin and keeps it runnung until the orchestrator cannot be
// pinged anymore.
func (p *Plugin) Run(e Exposer) {
	keepalive := make(chan bool)
	c := make(chan int)
	go p.launch(c, e)
	p.Fingerprint.Port = <-c
	p.handshake()
	go func() {
		for {
			time.Sleep(10 * time.Second)
			p.ping(keepalive)
		}
	}()
	<-keepalive
}

// launch starts the RPC connection and keeps respondung to incoming requests
func (p *Plugin) launch(c chan int, e Exposer) error {
	api, err := NewConnector(e)
	if err != nil {
		c <- 0
		return err
	}
	rpc.Register(api)
	ln, port, err := p.getListener()

	if err != nil {
		c <- 0
		return err
	}
	c <- port
	for {
		con, err := ln.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(con)
	}
	return nil
}

// ping checks if the orchestrator is still rechable via its exposed Ping function
func (p *Plugin) ping(c chan bool) {
	var reply bool
	call := new([]interface{})
	err := p.Client.Call("Connector.Ping", *call, &reply)
	if err != nil {
		c <- true
	}
}

// handshake connects to the orchestrator and communicates the port that provides
// the RPC interface that allows the orchestrator to communicate with the plugin.
func (p *Plugin) handshake() bool {
	client, err := rpc.Dial("tcp", p.Config.OrchConnString)
	if err != nil {
		log.Fatal(err)
	}
	var reply bool
	err = client.Call("Connector.Handshake", p.Fingerprint, &reply)
	if err != nil {
		log.Fatal(err)
	}
	p.Client = client
	return reply
}

// ---------------------------------------------------------------------------------
// Fingerprint
// ---------------------------------------------------------------------------------

// Fingerprint provides all infromation to identify a plugin and perform an handshake
// with the orchestrator program
type Fingerprint struct {
	Name string
	Port int
}

// ---------------------------------------------------------------------------------
// PluginConfiguration
// ---------------------------------------------------------------------------------

// PluginConfiguration hold all configuration information provided by the orchestrator
// via environment variables at launch time.
type PluginConfiguration struct {
	OrchConnString string
	MaxPort        int
	MinPort        int
}
