package logpredicate

import (
	"errors"
	"fmt"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/rs/zerolog/log"
)

// Name returns the name the predicate
func (*Predicate) Name() string {
	return "log"
}

// Get returns the plugin for the match predicate
func Get() *Predicate {
	return &Predicate{}
}

// Result returns nil as there is no result for this predicate
func (predicate *Predicate) Result() ctx.Result {
	return nil
}

// Predicate implements the conf.Plugin interface
type Predicate struct{}

// Predicate parameters
type params struct {
	Msg interface{}
}

// Call evaluate a predicate
func (*Predicate) Call(ctx *ctx.Ctx, config *conf.Params) bool {
	log := log.With().Str("predicate", "log").Logger()

	var p params
	if !conf.GetPredicateParams(ctx, config, &p) {
		log.Error().Err(errors.New("Invalid params, aborting")).Msg("")
		return false
	}

	log.Info().Str("log", fmt.Sprintf("%v", p.Msg)).Msg("")

	return true
}
