package predicate

import (
	"errors"
	"fmt"
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
		switch k {
		case "register":
			if r, ok := p[k].(string); !ok {
				log.Error().
					Msg("bad register, ignoring")
			} else if register != "" {
				log.Warn().
					Msg("Several 'register' found, " +
						"using the first one")
			} else {
				register = r
			}

		default:
			if res, isPlugin := plugins.Get(k); isPlugin {
				if plugin != nil {
					err := fmt.Errorf("Found both '%v' "+
						"& '%v', ignoring the later",
						pluginName, k)
					log.Error().Err(err).Msg("")
				} else {
					plugin = res
					pluginName = k
				}
			}

		}
	}
	if plugin == nil {
		log.Error().Err(errors.New("No plugin name found")).Msg("")
		return false
	}
	log := log.With().Str("predicate", pluginName).Logger()
	log.Debug().Msgf("Found predicate '%v'", pluginName)

	if args, ok := p[pluginName].(conf.Predicate); !ok {
		log.Error().Err(errors.New("Parameters must be a dict")).Msg("")
		return false
	} else {
		ctx.Results = make(map[string]interface{})
		result := plugin(ctx, args)
		if register != "" {
			log.Debug().Str("register", register).
				Msgf("Register result to %v", register)
			ctx.R[register] = ctx.Results
			ctx.R[register]["result"] = result
		}
		return result
	}
}
