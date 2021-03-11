package predicate

import (
	"net/http"

	"github.com/jsautret/go-api-broker/context"

	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/jsautret/go-api-broker/internal/plugins"
	"github.com/rs/zerolog/log"
)

func Process(p conf.Predicate, ctx *context.Ctx, r *http.Request) bool {
	var register, pluginName string
	var plugin plugins.Plugin
	for k := range p {
		log.Trace().Str("key", k).Msgf("Found Key %v for predicate", k)
		if res, isPlugin := plugins.Get(k); isPlugin {
			if plugin != nil {
				log.Warn().
					Msg("Several types of predicate found, ignoring...")
			} else {
				plugin = res
				pluginName = k
			}
		} else {
			switch k {
			case "register":
				if r, ok := p[k].(string); !ok {
					log.Error().
						Msg("bad register, ignoring")
				} else {
					if register != "" {
						log.Warn().
							Msg("Several 'register' found, ignoring, the others")
					} else {
						register = r
					}
				}
			}
		}
	}
	if plugin == nil {
		log.Error().Msg("No plugin name found")
		return false
	}
	log := log.With().Str("predicate", pluginName).Logger()

	log.Debug().Msgf("Found predicate '%v'", pluginName)

	if args, ok := p[pluginName].(conf.Predicate); !ok {
		log.Error().Msg("Parameters must be a dict")
		return false
	} else {
		result := plugin(ctx, args)
		if register != "" {
			log.Debug().Str("register", register).
				Msgf("Register result to %v", register)
			ctx.R[register] = make(map[string]interface{})
			ctx.R[register]["result"] = result
			for k, v := range ctx.Results {
				ctx.R[register][k] = v
			}
		}
		return result
	}
}
