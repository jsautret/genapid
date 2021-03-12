package plugins

import (
	"github.com/jsautret/go-api-broker/context"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/jsautret/go-api-broker/predicates/jsonrpc"
	"github.com/jsautret/go-api-broker/predicates/log"
	"github.com/jsautret/go-api-broker/predicates/match"
)

type Plugin func(*context.Ctx, conf.Params) bool

var (
	available map[string]Plugin
)

func init() {
	available = make(map[string]Plugin)
	available["match"] = match.Call
	available["jsonrpc"] = jsonrpc.Call
	available["log"] = log.Call
}

func Get(name string) (Plugin, bool) {
	if p, ok := available[name]; ok {
		return p, true
	}
	return nil, false
}
