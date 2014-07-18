package main

import (
	"sync"

	"github.com/influxdb/influxdb-go"
)

type Dbs struct {
	Mutex    *sync.Mutex
	Settings *Influxdb
	Clients  map[string]*influxdb.Client
}

func NewDbs(settings *Influxdb) *Dbs {
	var clients map[string]*influxdb.Client
	return &Dbs{
		Mutex:    &sync.Mutex{},
		Settings: settings,
		Clients:  clients,
	}
}

func (d *Dbs) Get(name string) (*influxdb.Client, error) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	client, ok := d.Clients[name]

	if !ok {

		influx, err := influxdb.NewClient(&influxdb.ClientConfig{
			Username: d.Settings.Username,
			Password: d.Settings.Password,
			Database: name,
			Host:     d.Settings.Host,
		})

		if err != nil {
			return nil, err
		}

		//d.Clients[name] = influx
		client = influx
	}

	return client, nil
}
