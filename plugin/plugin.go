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

type Fingerprint struct {
	Name string
	Port int
}

type PluginConfiguration struct {
	OrchConnString string
	MaxPort        int
	MinPort        int
}

type Plugin struct {
	Config      *PluginConfiguration
	Fingerprint *Fingerprint
	Client      *rpc.Client
}

func NewPlugin() (*Plugin, error) {
	max, _ := strconv.Atoi(os.Getenv("PLUGIN_MAX_PORT"))
	min, _ := strconv.Atoi(os.Getenv("PLUGIN_MIN_PORT"))
	connString := os.Getenv("ORCHESTRATOR_CONN_STRING")

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

func (p *Plugin) ping(c chan bool) {
	var reply bool
	call := new([]interface{})
	err := p.Client.Call("Connector.Ping", *call, &reply)
	if err != nil {
		c <- true
	}
}

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
