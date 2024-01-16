package config

import (
	"sync/atomic"
	"tonysoft.com/comm/internal/config/client"
	"tonysoft.com/comm/internal/config/node"
	"tonysoft.com/comm/internal/config/server"
)

type Config interface {
	client.Config | server.Config | node.Config
}

type Configurable[T Config] interface {
	Config() T
	SetConfig(T)
}

type DefaultConfigurable[T Config] struct {
	_config atomic.Pointer[T]
}

func (c *DefaultConfigurable[T]) Config() T {
	return *c._config.Load()
}

func (c *DefaultConfigurable[T]) SetConfig(cfg T) {
	c._config.Store(&cfg)
}
