package plugins

import (
	"github.com/jsautret/go-api-broker/context"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/jsautret/go-api-broker/predicates/match"
)

type Plugin func(*context.Ctx, conf.Predicate) bool

var (
	available map[string]Plugin
)

func init() {
	available = make(map[string]Plugin)
	available["match"] = match.Call
}

func Get(name string) (Plugin, bool) {
	if p, ok := available[name]; ok {
		return p, true
	}
	return nil, false
}
