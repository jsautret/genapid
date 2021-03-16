package predicate

import (
	"errors"
	"fmt"

	"github.com/jsautret/go-api-broker/ctx"
	"gopkg.in/yaml.v3"

	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/jsautret/go-api-broker/internal/plugins"
	"github.com/rs/zerolog/log"
)

// Evaluate a predicate from its parameters (from conf file) and
// current context
func Process(p conf.Predicate, c *ctx.Ctx) bool {
	var register, pluginName string
	var plugin plugins.Plugin
	for k := range p {
		log.Trace().Str("key", k).Msgf("Found Key %v for predicate", k)
		switch k {
		case "register":
			assignRegister(&register, p[k])
		case "set":
			processSet(c, p[k])
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
		c.Results = make(map[string]interface{})

		result := plugin(c, args)
		log.Debug().Bool("value", result).Msg("End predicate")
		if register != "" {
			log.Debug().Str("register", register).
				Msgf("Register result to %v", register)
			c.R[register] = c.Results
			c.R[register]["result"] = result
		}
		return result
	}
}

// Store registered results with 'register' option
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

// Store variable values set by 'set' option
func processSet(c *ctx.Ctx, n yaml.Node) {
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
			if !conf.GetParams(c, arg, &field) {
				log.Error().
					Err(fmt.Errorf("Invalid value for %v", k)).
					Msg("")
				continue
			}
			log.Trace().Msgf("set %v='%v'", k, field[k])
			c.V[k] = field[k]
		}
	}
}

// Find a plugin corresponding to the predicate set in the conf file
// and set plugin & pluginName parameters accordingly
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
