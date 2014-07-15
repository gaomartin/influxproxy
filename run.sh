#!/bin/bash
export ORCH_ADDRESS=0.0.0.0
export ORCH_MINPORT=4000
export ORCH_MAXPORT=5000
#export ORCH_PLUGINS=/tmp/plugins/c,/tmp/plugins/a,/tmp/plugins/b
export ORCH_PLUGINS=/tmp/plugins/a,/tmp/plugins/b

(cd ../influxproxy-collectd-plugin && go install)
mv $GOPATH/bin/influxproxy-collectd-plugin /tmp/plugins/a
go run influxproxy.go
