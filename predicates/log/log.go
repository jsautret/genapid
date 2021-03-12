package log

import (
	"errors"

	"github.com/jsautret/go-api-broker/context"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/rs/zerolog/log"
)

type params struct {
	Msg string
}

func Call(ctx *context.Ctx, config conf.Params) bool {
	log := log.With().Str("predicate", "log").Logger()

	var p params
	if !conf.GetParams(ctx, config, &p) {
		log.Error().Err(errors.New("Invalid params, aborting")).Msg("")
		return false
	}

	log.Info().Str("msg", p.Msg).Msg("")

	return true
}
