package plugins

import (
	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
)

// Plugin is the type of the plugin entry point function
type Plugin func(*ctx.Ctx, conf.Params) bool

// Plugins store entrypoint functions of enabled plugins
type Plugins map[string]Plugin

var (
	available Plugins
)

// Get returns entrypoint of Plugin, if enabled
func Get(name string) (Plugin, bool) {
	if p, ok := available[name]; ok {
		return p, true
	}
	return nil, false
}

// List available plugins
func List() Plugins {
	return available
}

// Add a plugin entrypoint to the enabled plugins
func Add(name string, p Plugin) {
	if available == nil {
		available = make(map[string]Plugin, 20)
	}
	available[name] = p
}
