package log

import (
	"errors"
	"fmt"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/rs/zerolog/log"
)

// Predicate parameters
type params struct {
	Msg interface{}
}

// Evaluate predicate
func Call(ctx *ctx.Ctx, config conf.Params) bool {
	log := log.With().Str("predicate", "log").Logger()

	var p params
	if !conf.GetPredicateParams(ctx, config, &p) {
		log.Error().Err(errors.New("Invalid params, aborting")).Msg("")
		return false
	}

	log.Info().Str("msg", fmt.Sprintf("%v", p.Msg)).Msg("")

	return true
}
