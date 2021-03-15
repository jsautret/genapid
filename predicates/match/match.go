package match

import (
	"errors"
	"regexp"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"

	"github.com/rs/zerolog/log"
)

type params struct {
	String string
	Fixed  string
	Regexp string
}

func Call(ctx *ctx.Ctx, config conf.Params) bool {
	log := log.With().Str("predicate", "match").Logger()

	var p params
	if !conf.GetParams(ctx, config, &p) {
		log.Error().Err(errors.New("Invalid params, aborting")).Msg("")
		return false
	}

	log.Debug().Str("string", p.String).Msg("")
	if p.Fixed != "" {
		log.Debug().Str("fixed", p.Fixed).Msg("")
		return p.Fixed == p.String
	}
	if p.Regexp != "" {
		log.Debug().Str("regexp", p.Regexp).Msg("")

		if r, err := regexp.Compile(p.Regexp); err != nil {
			log.Error().Err(err).Msg("invalid 'regexp'")
			return false
		} else {
			if res := r.FindStringSubmatch(p.String); len(res) == 0 {
				return false
			} else {
				log.Debug().Msgf("'regexp' matched %v", res)
				ctx.Results["matches"] = res
				return true
			}
		}
	}
	log.Error().Err(errors.New("Missing one of 'fixed' or 'regexp' " +
		"for predicate 'match'"))
	return false
}
