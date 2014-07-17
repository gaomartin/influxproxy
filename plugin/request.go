package plugin

import (
	"net/url"
)

type Request struct {
	Query url.Values
	Body  string
}
