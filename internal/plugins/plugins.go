package plugins

import (
	"github.com/jsautret/go-api-broker/internal/conf"
)

// Plugins store entrypoint functions of enabled plugins
type Plugins map[string]conf.Plugin

var (
	available Plugins
)

// Get returns entrypoint of Plugin, if enabled
func Get(name string) (conf.Plugin, bool) {
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
func Add(name string, p conf.Plugin) {
	if available == nil {
		available = make(Plugins, 20)
	}
	available[name] = p
}
