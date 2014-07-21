package main

import (
	"log"
	"sync"

	"github.com/influxdb/influxdb-go"
)

type Dbs struct {
	Mutex    *sync.Mutex
	Settings *Influxdb
	Clients  map[string]*influxdb.Client
}

func NewDbs(settings *Influxdb) *Dbs {
	return &Dbs{
		Mutex:    &sync.Mutex{},
		Settings: settings,
		Clients:  make(map[string]*influxdb.Client),
	}
}

func (dbs *Dbs) Get(name string) (*influxdb.Client, error) {
	client, ok := dbs.Clients[name]

	if !ok {

		influx, err := influxdb.NewClient(&influxdb.ClientConfig{
			Username: dbs.Settings.Username,
			Password: dbs.Settings.Password,
			Database: name,
			Host:     dbs.Settings.Host,
		})
		if err != nil {
			return nil, err
		}

		err = influx.Ping()
		if err != nil {
			return nil, err
		}

		dbs.Mutex.Lock()
		defer dbs.Mutex.Unlock()
		dbs.Clients[name] = influx
		log.Println("New database registered: " + name)
		client = influx
	}

	return client, nil
}
