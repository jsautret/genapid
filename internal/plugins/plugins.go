package plugins

import (
	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
)

type Plugin func(*ctx.Ctx, conf.Params) bool
type Plugins map[string]Plugin

var (
	available Plugins
)

func Get(name string) (Plugin, bool) {
	if p, ok := available[name]; ok {
		return p, true
	}
	return nil, false
}

func List() Plugins {
	return available
}

func Add(name string, p Plugin) {
	if available == nil {
		available = make(map[string]Plugin, 20)
	}
	available[name] = p
}
