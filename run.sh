#!/bin/bash
export ORCH_ADDRESS=0.0.0.0
export ORCH_MINPORT=4000
export ORCH_MAXPORT=5000
export ORCH_PLUGINDIR=/tmp/plugins
export ORCH_PLUGINS=/tmp/plugins/test,/tmp/plugins/bla

go run orch.go