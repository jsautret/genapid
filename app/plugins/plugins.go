// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package plugins

import (
	"github.com/jsautret/genapid/genapid"
)

// Plugins store entrypoint functions of enabled plugins
type Plugins map[string]func() genapid.Predicate

var (
	available Plugins
)

// Get returns entrypoint of Plugin, if enabled
func Get(name string) genapid.Predicate {
	if new, ok := available[name]; ok {
		return new()
	}
	return nil
}

// List available plugins
func List() Plugins {
	return available
}

// Add a plugin entrypoint to the enabled plugins
func Add(name string, new func() genapid.Predicate) {
	if available == nil {
		available = make(Plugins, 20)
	}
	available[name] = new
}
