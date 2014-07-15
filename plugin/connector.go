package plugin

import ()

type Exposer interface {
	Ping() bool
	Describe() string
	Run(data string) string
}
