package plugins

import (
	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"

	"github.com/rs/zerolog/log"
)

type Plugin func(*ctx.Ctx, conf.Params) bool

var (
	available map[string]Plugin
)

func Get(name string) (Plugin, bool) {
	if p, ok := available[name]; ok {
		return p, true
	}
	return nil, false
}

func Add(name string, p Plugin) {
	if available == nil {
		available = make(map[string]Plugin, 20)
	}
	log.Info().Str("plugin", name).Msg("Plugin enabled")
	available[name] = p
}
