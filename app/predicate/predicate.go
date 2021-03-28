// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package predicate

import (
	"errors"
	"fmt"

	"github.com/jsautret/genapid/ctx"
	"github.com/jsautret/genapid/genapid"
	"gopkg.in/yaml.v3"

	"github.com/jsautret/genapid/app/conf"
	"github.com/jsautret/genapid/app/plugins"
	"github.com/rs/zerolog"
)

type pOptions struct {
	register, name, result, when string
	p                            genapid.Predicate
	pipe                         conf.Pipe
	variable                     []map[string]interface{}
	def                          ctx.DefaultParams
}

func (o pOptions) hasPredicate() bool {
	return o.p != nil || o.pipe.Pipe != nil ||
		len(o.variable) != 0 || len(o.def) != 0
}

// Read all options and predicate or pipe
func getOptions(log zerolog.Logger, cfg *conf.Predicate, c *ctx.Ctx) (*pOptions, bool) {
	var o pOptions
	for k, node := range *cfg {
		log.Trace().Str("key", k).Msgf("Found Key %v for predicate", k)
		switch k {
		case "register":
			if !assignOption(log, "register", &o.register, node) {
				return nil, false
			}
		case "result":
			if !assignOption(log, "result", &o.result, node) {
				return nil, false
			}
		case "name":
			if !assignOption(log, "name", &o.name, node) {
				return nil, false
			}
		case "when":
			if !assignOption(log, "when", &o.when, node) {
				return nil, false
			}
		case "variable":
			if !assignVariable(log, &o, node) {
				return nil, false
			}
		case "default":
			if !assignDefault(log, &o, node) {
				return nil, false
			}
		case "pipe":
			if !assignPipe(log, &o, node) {
				return nil, false
			}
		default:
			// Try to check if it is a predicate
			if !assignPlugin(log, &o, k) {
				return nil, false
			}
		}
	}
	return &o, true
}

// Process evaluate a predicate or a pipe from from conf file and
// current context
func Process(log zerolog.Logger, cfg *conf.Predicate, c *ctx.Ctx) bool {
	o, ok := getOptions(log, cfg, c)
	if !ok {
		return false
	}
	var when bool
	if o.when != "" {
		if !conf.GetParams(c, o.when, &when) {
			log.Warn().Err(errors.New("'when' is not boolean")).Msg("")
			// we consider it false
			return true
		}
		if !when {
			// if when is false, we continue to next predicate
			return true
		}
	}

	if len(o.variable) > 0 {
		return processVariable(log, o, c)
	}
	if len(o.def) > 0 {
		return processDefault(log, o, c)
	}
	if o.pipe.Pipe != nil {
		return pipeHandling(log, c, o)
	}
	if o.p == nil {
		log.Error().Err(errors.New("No predicate found")).Msg("")
		return false
	}
	log = log.With().Str("predicate", o.p.Name()).Str("name", o.name).Logger()
	result := processPredicate(log, o, cfg, c)

	if o.register != "" { // Save predicate results
		log := log.With().Str("register", o.register).Logger()
		log.Debug().Msgf("Register result to %v", o.register)
		if r := o.p.Result(); r != nil {
			if val, ok := r["result"]; ok {
				// Predicate is no supposed to use 'result' field
				log.Warn().Msgf("Value is lost 'result':%v", val)
			}
			r["result"] = result // real predicate result, not 'result:' option
			c.R[o.register] = r

		} else {
			// Predicate doesn't any result data, we
			// just save its boolean evaluation
			c.R[o.register] = ctx.Result{"result": result}
		}
	}
	if o.result != "" {
		if !conf.GetParams(c, o.result, &result) {
			log.Error().Err(errors.New("'result' is not boolean")).Msg("")
			return false
		}
	}
	log.Debug().Bool("value", result).Msg("End predicate")
	return result
}

func processPredicate(log zerolog.Logger,
	o *pOptions, cfg *conf.Predicate, c *ctx.Ctx) bool {
	name := o.p.Name()
	log.Debug().Msgf("Found predicate '%v'", name)

	argsNode := (*cfg)[name]
	args := conf.Params{Name: name}
	if err := argsNode.Decode(&(args.Conf)); err != nil {
		log.Error().Err(err).Msgf("Parameters for %v must be a dict", name)
		return false
	}
	if !genapid.InitPredicate(log, c, o.p, &args) {
		return false
	}

	// Evaluate predicate
	return o.p.Call(log, c)
}

// Decode & store an option
func assignOption(log zerolog.Logger, name string, option *string, n yaml.Node) bool {
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

func assignVariable(log zerolog.Logger, o *pOptions, n yaml.Node) bool {
	log = log.With().Str("predicate", "variable").Logger()
	if o.variable != nil {
		log.Error().Err(errors.New("Several 'variable' declared")).Msg("")
		return false
	}
	if o.hasPredicate() {
		log.Error().Err(errors.New("'variable' declared with another predicate")).Msg("")
		return false
	}
	var args []map[string]interface{}
	if err := n.Decode(&(args)); err != nil {
		log.Error().Err(err).Msg("'variable' parameters must be a dict")
		return false
	}
	o.variable = args
	return true
}

// Store variable values set by 'variable' option
func processVariable(log zerolog.Logger, o *pOptions, c *ctx.Ctx) bool {
	log = log.With().Str("predicate", "variable").Logger()
	for i := 0; i < len(o.variable); i++ {
		for k := range o.variable[i] {
			var field map[string]interface{}
			arg := make(map[string]interface{})
			arg[k] = o.variable[i][k]
			if !conf.GetParams(c, arg, &field) {
				log.Error().
					Err(fmt.Errorf("Invalid value for %v", k)).
					Msg("")
				return false
			}
			log.Trace().Msgf("variable %v='%v'", k, field[k])
			c.V[k] = field[k]
		}
	}
	return true
}

// Store variable values set by 'default' option
func assignDefault(log zerolog.Logger, o *pOptions, n yaml.Node) bool {
	log = log.With().Str("predicate", "default").Logger()
	if o.hasPredicate() {
		log.Error().Err(errors.New("'default' declared with another predicate")).Msg("")
		return false
	}
	d := ctx.DefaultParams{}
	if err := n.Decode(&d); err != nil {
		log.Error().Err(err).Msg("'default' parameters must be a dict")
		return false
	}
	o.def = d
	return true
}

func processDefault(log zerolog.Logger, o *pOptions, c *ctx.Ctx) bool {
	log = log.With().Str("predicate", "default").Logger()
	return conf.AddDefault(log, c, &o.def)
}

// Find a plugin corresponding to the predicate set in the conf file
func assignPlugin(log zerolog.Logger, o *pOptions, name string) bool {
	if res := plugins.Get(name); res != nil {
		if o.p != nil {
			err := fmt.Errorf("Both '%v' & '%v' declared",
				(o.p).Name(), name)
			log.Error().Err(err).Msg("")
			return false
		}
		if o.hasPredicate() {
			err := fmt.Errorf("'%v' declared with another predicate",
				name)
			log.Error().Err(err).Msg("")
			return false
		}
		o.p = res
		return true
	}
	log.Error().
		Err(fmt.Errorf("Unknown predicate '%v'", name)).Msg("")
	return false
}

func assignPipe(log zerolog.Logger, o *pOptions, n yaml.Node) bool {
	if o.pipe.Pipe != nil {
		log.Error().Msg("Several 'pipe' declared")
		return false
	}
	if o.hasPredicate() {
		log.Error().Msg("'pipe' declared with another predicate")
		return false
	}
	p := []conf.Predicate{}
	if err := n.Decode(&p); err != nil {
		log.Error().Err(err).Msg("invalid 'pipe'")
		return false
	}
	o.pipe.Pipe = p
	return true
}

func pipeHandling(log zerolog.Logger, c *ctx.Ctx, o *pOptions) bool {
	log = log.With().Str("pipe", o.name).Logger()
	if o.register != "" {
		log.Error().Err(
			errors.New("Cannot set 'register' option on a " +
				" predicate")).Msg("")
		return false
	}
	o.pipe.Name = o.name
	ProcessPipe(&o.pipe, c)
	// Always continue after a pipe, unless result: option is set
	// and evaluate to false
	result := true
	if o.result != "" {
		if !conf.GetParams(c, o.result, &result) {
			log.Error().Err(errors.New("'result' is not boolean")).Msg("")
			return false
		}
	}
	return result
}
