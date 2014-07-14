package plugin

import (
	"errors"
	"fmt"
	"net"
)

func getListener(minPort int, maxPort int, address string) (net.Listener, int, error) {
	for port := minPort; port <= maxPort; port++ {
		address := fmt.Sprintf("%s:%d", address, port)
		listener, err := net.Listen("tcp", address)
		if err == nil {
			return listener, port, nil
		}
	}

	return nil, 0, errors.New("Couldn't bind orchestrator TCP listener")
}
