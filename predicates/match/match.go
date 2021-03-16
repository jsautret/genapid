package match

import (
	"errors"
	"regexp"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"

	"github.com/rs/zerolog/log"
)

// Predicate parameters
type params struct {
	String string
	Fixed  string
	Regexp string
}

// Call evaluate predicate
func Call(ctx *ctx.Ctx, config conf.Params) bool {
	log := log.With().Str("predicate", "match").Logger()

	var p params
	if !conf.GetPredicateParams(ctx, config, &p) {
		log.Error().Err(errors.New("Invalid params, aborting")).Msg("")
		return false
	}

	if p.String == "" {
		log.Error().Err(errors.New("'string' is missing or empty")).
			Msg("")
		return false
	}
	log.Debug().Str("string", p.String).Msg("")
	if p.Fixed != "" {
		log.Debug().Str("fixed", p.Fixed).Msg("")
		return p.Fixed == p.String
	}
	if p.Regexp != "" {
		log.Debug().Str("regexp", p.Regexp).Msg("")

		r, err := regexp.Compile(p.Regexp)
		if err != nil {
			log.Error().Err(err).Msg("invalid 'regexp'")
			return false
		}
		res := r.FindStringSubmatch(p.String)
		if len(res) == 0 {
			return false
		}
		log.Debug().Msgf("'regexp' matched %v", res)
		ctx.Results["matches"] = res
		return true
	}
	log.Error().Err(errors.New("Missing one of 'fixed' or 'regexp'")).Msg("")
	return false
}
