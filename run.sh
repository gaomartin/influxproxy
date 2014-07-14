#!/bin/bash
export ORCH_ADDRESS=0.0.0.0
export ORCH_MINPORT=4000
export ORCH_MAXPORT=5000
export ORCH_PLUGINDIR=/tmp/plugins
export ORCH_PLUGINS=/tmp/plugins/a,/tmp/plugins/bla

go run influxproxy.go
