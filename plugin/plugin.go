package plugin

import (
	"os"
	"path/filepath"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"strconv"
	"log"
)

type Fingerprint struct {
	Name string
	Port int
}

type PluginConfiguration struct {
	OrchConnString   string
	Address	         string
	MaxPort          int
	MinPort          int
}

type Plugin struct {
	Config      *PluginConfiguration
	Fingerprint *Fingerprint
}

func NewPlugin() (*Plugin, error) {
	// TODO: Recieve a Struct that implements the exposer interface and expose its fuctions via RCP
	max, _ := strconv.Atoi(os.Getenv("PLUGIN_MAX_PORT"))
	min, _ := strconv.Atoi(os.Getenv("PLUGIN_MIN_PORT"))
	address := os.Getenv("PLUGIN_ADDRESS")
	connString := os.Getenv("ORCHESTRATOR_CONN_STRING")

	name := filepath.Base(os.Args[0])

	if max != 0 && min != 0 && connString != "" && address != "" && name != "" {
		conf := &PluginConfiguration{
			OrchConnString: connString,
			Address:        address,
			MaxPort:        max,
			MinPort:        min,
		}
		fp := &Fingerprint{
			Name: name,
		}

		p := &Plugin{
			Config: conf,
			Fingerprint: fp,
		}
		return p, nil
	} else {
		return nil, errors.New("Not enough data to build plugin")
	}
}

func (p *Plugin) getListener() (net.Listener, int, error) {
	for port := p.Config.MinPort; port <= p.Config.MaxPort; port++ {
		connection := fmt.Sprintf("%s:%d", p.Config.Address, port)
		listener, err := net.Listen("tcp", connection)
		if err == nil {
			return listener, port, nil
		}
	}

	return nil, 0, errors.New("Could not get TCP listener, maybe all ports are already used")
}

func (p *Plugin) launch(c chan int) error {
	// api := NewPluginApi(this)
	// rpc.Register(api)
	ln, port, err := p.getListener()

	if err != nil {
		c <- 0
		return err
	}
	c <- port
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(c)
	}
	return nil
}

func (p *Plugin) Run() {
	c := make(chan int)
	go p.launch(c)
	p.Fingerprint.Port = <- c
	p.handshake()
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
	return reply
}