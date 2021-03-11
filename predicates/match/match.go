package match

import (
	"errors"
	"regexp"

	"github.com/jsautret/go-api-broker/context"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/rs/zerolog/log"
)

type params struct {
	String string
	Fixed  string
	Regexp string
}

func Call(ctx *context.Ctx, config conf.Predicate) bool {
	log := log.With().Str("predicate", "match").Logger()

	var p params
	conf.GetParams(ctx, config, &p)

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
				results := make(map[string]interface{})
				results["matches"] = res
				ctx.Results = results
				return true
			}
		}
	}

	log.Error().Err(errors.New("Missing one of 'fixed' or 'regexp' " +
		"for predicate 'match'"))
	return false
}
