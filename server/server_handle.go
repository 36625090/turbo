package server

import (
	"github.com/36625090/turbo/config"
	"github.com/36625090/turbo/logical"
	"github.com/36625090/turbo/transport"
	"github.com/go-various/consul"
	"github.com/hashicorp/go-hclog"
)

type TurboContext struct {
	Backends map[string]logical.Backend
	Config   *config.GlobalConfig
	Consul   consul.Client
	Logger   hclog.InterceptLogger
	Transport *transport.Transport
}
