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
	var register, pluginName, name, stop string
	var plugin plugins.Plugin
	var pipe conf.Pipe
	for k := range p {
		log.Trace().Str("key", k).Msgf("Found Key %v for predicate", k)
		switch k {
		case "register":
			assignRegister(&register, p[k])
		case "set":
			processSet(c, p[k])
		case "pipe":
			assignPipe(&pipe, p[k])
		case "stop":
			assignStop(&stop, p[k])
		case "name":
			n := p[k]
			if err := n.Decode(&name); err != nil {
				log.Error().Err(err).Msg("invalid 'name'")
			}
		default:
			assignPlugin(&plugin, &pluginName, k)
		}
	}
	if plugin != nil && pipe.Pipe != nil {
		log.Error().Err(fmt.Errorf("Both 'pipe' & '%v' declared",
			pluginName)).Msg("")
		return false
	}
	if pipe.Pipe != nil {
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
		log.Error().Msg("Several 'register' declared, " +
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

// Store 'stop' option for pipe
func assignStop(stop *string, n yaml.Node) {
	if *stop != "" {
		log.Error().Msg("Several 'stop' declared, " +
			"using the first found")
		return
	}
	if err := n.Decode(stop); err != nil {
		log.Error().Err(err).Msg("invalid 'stop'")
	}
}
