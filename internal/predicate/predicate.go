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

// Process evaluate a predicate or a pipe from from conf file and
// current context
func Process(p conf.Predicate, c *ctx.Ctx) bool {
	var register, pluginName, name, stop string
	var plugin conf.Plugin
	var pipe conf.Pipe
	// Read all options and predicate name or pipe
	for k, v := range p {
		log.Trace().Str("key", k).Msgf("Found Key %v for predicate", k)
		switch k {
		case "register":
			assignOption("register", &register, v)
		case "set":
			processSet(c, v)
		case "stop":
			assignOption("stop", &stop, v)
		case "name":
			assignOption("name", &name, v)
		case "pipe":
			assignPipe(&pipe, v)
		default:
			// Try to check if option is a predicate
			assignPlugin(&plugin, &pluginName, k)
		}
	}
	if plugin != nil && pipe.Pipe != nil {
		log.Error().Err(fmt.Errorf("Both 'pipe' & '%v' declared",
			pluginName)).Msg("")
		return false
	}
	if pipe.Pipe != nil {
		log := log.With().Str("pipe", name).Logger()
		if register != "" {
			log.Warn().Err(
				errors.New("'register' set on a predicate," +
					"ignoring")).Msg("")
		}

		pipe.Name = name
		ProcessPipe(pipe, c)
		stopValue := false // Always continue after a pipe
		if stop != "" {
			if !conf.GetParams(c, stop, &stopValue) {
				log.Warn().Err(errors.New("'stop' is not boolean"))
			}
		}
		// Always continue after a pipe, unless stop is true
		return !stopValue
	}
	if plugin == nil {
		log.Error().Err(errors.New("No predicate found")).Msg("")
		return false
	}
	log := log.With().Str("predicate", pluginName).Logger()
	log.Debug().Msgf("Found predicate '%v'", pluginName)
	if stop != "" {
		log.Warn().Err(
			errors.New("'stop' set on a predicate, ignoring")).Msg("")
	}

	argsNode := p[pluginName]
	args := conf.Params{Name: pluginName}
	if err := argsNode.Decode(&(args.Conf)); err != nil {
		log.Error().Err(err).Msg("Parameters must be a dict")
		return false
	}
	c.Results = make(map[string]interface{})

	result := plugin.Call(c, args)
	log.Debug().Bool("value", result).Msg("End predicate")
	if register != "" {
		log.Debug().Str("register", register).
			Msgf("Register result to %v", register)
		c.R[register] = c.Results
		c.R[register]["result"] = result
	}
	return result
}

// Decode & store an option
func assignOption(name string, option *string, n yaml.Node) {
	if *option != "" {
		log.Error().Msgf("Several '%v' declared, "+
			"using the first found", name)
		return
	}
	if err := n.Decode(option); err != nil {
		log.Error().Err(err).Msgf("invalid '%v'", name)
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
func assignPlugin(plugin *conf.Plugin, pluginName *string, k string) {
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

func assignPipe(pipe *conf.Pipe, n yaml.Node) {
	if pipe.Pipe != nil {
		log.Error().Msg("Several 'pipe' declared, " +
			"using the first found")
		return
	}
	p := []conf.Predicate{}
	if err := n.Decode(&p); err != nil {
		log.Error().Err(err).Msg("invalid 'pipe'")
		return
	}
	pipe.Pipe = p
}
