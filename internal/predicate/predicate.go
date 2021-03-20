package predicate

import (
	"errors"
	"fmt"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/genapid"
	"gopkg.in/yaml.v3"

	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/jsautret/go-api-broker/internal/plugins"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Process evaluate a predicate or a pipe from from conf file and
// current context
func Process(log zerolog.Logger, cfg *conf.Predicate, c *ctx.Ctx) bool {
	var register, name, stop string
	var p genapid.Predicate
	var pipe conf.Pipe
	cmd := false
	// Read all options and predicate name or pipe
	for k, node := range *cfg {
		log.Trace().Str("key", k).Msgf("Found Key %v for predicate", k)
		switch k {
		case "register":
			if !assignOption("register", &register, node) {
				return false
			}
		case "stop":
			if !assignOption("stop", &stop, node) {
				return false
			}
		case "name":
			if !assignOption("name", &name, node) {
				return false
			}
		case "set":
			if !processSet(c, node, cmd) {
				return false
			}
			cmd = true
		case "default":
			if !processDefault(c, node, cmd) {
				return false
			}
			cmd = true
		case "pipe":
			if !assignPipe(&pipe, node, cmd) {
				return false
			}
		default:
			// Try to check if option is a predicate
			assignPlugin(&p, k)
			if p == nil {
				log.Error().
					Err(fmt.Errorf("Unknown predicate '%v'", k))
				return false
			}
		}
	}
	if p != nil && cmd {
		log.Error().Err(fmt.Errorf("Cannot use '%v' and a command",
			p.Name())).Msg("")
		return false
	}
	if cmd && pipe.Pipe != nil {
		log.Error().
			Err(errors.New("Cannot use 'pipe' and a command")).Msg("")
		return false
	}
	if cmd { // a command was executed
		return true
	}
	if pipe.Pipe != nil {
		return pipeHandling(log, c, pipe, name, register, stop)
	}
	if p == nil {
		log.Error().Err(errors.New("No predicate found")).Msg("")
		return false
	}
	log = log.With().Str("predicate", p.Name()).Logger()
	log.Debug().Msgf("Found predicate '%v'", p.Name())
	if stop != "" {
		log.Error().Err(
			errors.New("'stop' set on a predicate")).Msg("")
		return false
	}

	argsNode := (*cfg)[p.Name()]
	args := conf.Params{Name: p.Name()}
	if err := argsNode.Decode(&(args.Conf)); err != nil {
		log.Error().Err(err).Msg("Parameters must be a dict")
		return false
	}
	if !genapid.InitPredicate(log, c, p, &args) {
		return false
	}

	// Evaluate predicate
	result := p.Call(log)

	if register != "" { // Save predicate results
		log := log.With().Str("register", register).Logger()
		log.Debug().Msgf("Register result to %v", register)
		if r := p.Result(); r != nil {
			if val, ok := r["result"]; ok {
				// Predicate is no supposed to use 'result' field
				log.Warn().Msgf("Value is lost 'result':%v", val)
			}
			r["result"] = result
			c.R[register] = r
		} else {
			// Predicate doesn't any result data, we
			// just save its boolean evaluation
			c.R[register] = ctx.Result{"result": result}
		}
	}
	log.Debug().Bool("value", result).Msg("End predicate")
	return result
}

// Decode & store an option
func assignOption(name string, option *string, n yaml.Node) bool {
	if *option != "" {
		log.Error().Msgf("Several '%v' declared", name)
		return false
	}
	if err := n.Decode(option); err != nil {
		log.Error().Err(err).Msgf("invalid '%v'", name)
		return false
	}
	return true
}

// Store variable values set by 'set' option
func processDefault(c *ctx.Ctx, n yaml.Node, cmd bool) bool {
	if cmd {
		log.Error().Msg("Cannot use 'default' with another command")
		return false
	}
	d := ctx.DefaultParams{}
	if err := n.Decode(&d); err != nil {
		log.Error().Err(err).Msg("'default' parameters must be a dict")
		return false
	}
	conf.AddDefault(c, &d)
	return true
}

// Store variable values set by 'set' option
func processSet(c *ctx.Ctx, n yaml.Node, cmd bool) bool {
	if cmd {
		log.Error().Msg("Cannot use 'set' with another command")
		return false
	}
	var args []map[string]interface{}
	if err := n.Decode(&(args)); err != nil {
		log.Error().Err(err).Msg("'set' parameters must be a dict")
		return false
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
				return false
			}
			log.Trace().Msgf("set %v='%v'", k, field[k])
			c.V[k] = field[k]
		}
	}
	return true
}

// Find a plugin corresponding to the predicate set in the conf file
// and set plugin & pluginName parameters accordingly
func assignPlugin(p *genapid.Predicate, name string) {
	if *p != nil {
		err := fmt.Errorf("Found both '%v' & '%v', "+
			"ignoring the later",
			(*p).Name(), name)
		log.Error().Err(err).Msg("")
		return
	}
	if res := plugins.Get(name); res != nil {
		*p = res
	}
}

func assignPipe(pipe *conf.Pipe, n yaml.Node, cmd bool) bool {
	if cmd {
		log.Error().Msg("Cannot use 'pipe' with another command")
		return false
	}
	if pipe.Pipe != nil {
		log.Error().Msg("Several 'pipe' declared")
		return false
	}
	p := []conf.Predicate{}
	if err := n.Decode(&p); err != nil {
		log.Error().Err(err).Msg("invalid 'pipe'")
		return false
	}
	pipe.Pipe = p
	return true
}

func pipeHandling(log zerolog.Logger, c *ctx.Ctx,
	pipe conf.Pipe, name, register, stop string) bool {
	log.With().Str("pipe", name).Logger()
	if register != "" {
		log.Error().Err(
			errors.New("Cannot set 'register' option on a " +
				" predicate")).Msg("")
		return false
	}
	pipe.Name = name
	ProcessPipe(&pipe, c)
	stopValue := false // Always continue after a pipe
	if stop != "" {
		if !conf.GetParams(c, stop, &stopValue) {
			log.Warn().Err(errors.New("'stop' is not boolean"))
		}
	}
	// Always continue after a pipe, unless stop is true
	return !stopValue
}
