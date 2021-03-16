package predicate

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jsautret/go-api-broker/ctx"
	"gopkg.in/yaml.v3"

	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/jsautret/go-api-broker/internal/plugins"
	"github.com/rs/zerolog/log"
)

func Process(p conf.Predicate, ctx *ctx.Ctx, r *http.Request) bool {
	var register, pluginName string
	var plugin plugins.Plugin
	for k := range p {
		log.Trace().Str("key", k).Msgf("Found Key %v for predicate", k)
		switch k {
		case "register":
			assignRegister(&register, p[k])
		case "set":
			processSet(ctx, p[k])
		default:
			assignPlugin(&plugin, &pluginName, k)
		}
	}
	if plugin == nil {
		log.Error().Err(errors.New("No plugin name found")).Msg("")
		return false
	}
	log := log.With().Str("predicate", pluginName).Logger()
	log.Debug().Msgf("Found predicate '%v'", pluginName)

	argsNode := p[pluginName]
	args := conf.Params{Name: pluginName}
	if err := argsNode.Decode(&(args.Conf)); err != nil {
		log.Error().Err(err).Msg("Parameters must be a dict")
		return false
	} else {
		ctx.Results = make(map[string]interface{})
		result := plugin(ctx, args)
		log.Debug().Bool("value", result).Msg("End predicate")
		if register != "" {
			log.Debug().Str("register", register).
				Msgf("Register result to %v", register)
			ctx.R[register] = ctx.Results
			ctx.R[register]["result"] = result
		}
		return result
	}
}

func assignRegister(register *string, n yaml.Node) {
	if *register != "" {
		log.Warn().Msg("Several 'register' declared, " +
			"using the first found")
		return
	}
	if err := n.Decode(register); err != nil {
		log.Error().Err(err).Msg("invalid 'register'")
	}
}

func processSet(ctx *ctx.Ctx, n yaml.Node) {
	var args []map[string]interface{}
	if err := n.Decode(&(args)); err != nil {
		log.Error().Err(err).Msg("'set' parameters must be a dict")
		return
	}
	for i := 0; i < len(args); i++ {
		for k := range args[i] {
			var field map[string]interface{}
			arg := make(map[string]interface{})
			arg[k] = args[i][k]
			if !conf.GetParams(ctx, arg, &field) {
				log.Error().
					Err(fmt.Errorf("Invalid value for %v", k)).
					Msg("")
				continue
			}
			log.Trace().Msgf("set %v='%v'", k, field[k])
			ctx.V[k] = field[k]
		}
	}
}

func assignPlugin(plugin *plugins.Plugin, pluginName *string, k string) {
	if *plugin != nil {
		err := fmt.Errorf("Found both '%v' & '%v', "+
			"ignoring the later",
			*pluginName, k)
		log.Error().Err(err).Msg("")
		return
	}
	if res, isPlugin := plugins.Get(k); isPlugin {
		*plugin = res
		*pluginName = k
	}
}
