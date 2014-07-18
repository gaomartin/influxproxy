package main

import (
	"sync"

	"github.com/influxdb/influxdb-go"
)

type Dbs struct {
	Settings *Influxdb
	Mutex    *sync.Mutex
	Clients  map[string]*influxdb.Client
}

func NewDbs(settings *Influxdb) *Dbs {
	return &Dbs{Mutex: &sync.Mutex{}, Settings: settings}
}

func (d *Dbs) Get(name string) (*influxdb.Client, error) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	client := d.Clients[name]

	if client == nil {

		influx, err := influxdb.NewClient(&influxdb.ClientConfig{
			Username: d.Settings.Username,
			Password: d.Settings.Password,
			Database: name,
			Host:     d.Settings.Host,
		})

		if err != nil {
			return nil, err
		}

		d.Clients[name] = influx
		client = d.Clients[name]
	}

	return client, nil
}
