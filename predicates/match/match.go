package match

import (
	"regexp"

	"github.com/jsautret/go-api-broker/context"

	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/jsautret/go-api-broker/internal/tmpl"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
)

type fields struct {
	String string
	Fixed  string
	Regexp string
}

func Call(ctx *context.Ctx, config conf.Predicate) bool {
	log := log.With().Str("predicate", "match").Logger()

	var f fields
	if err := mapstructure.Decode(config, &f); err != nil {
		log.Error().Msgf("Incorrect fields for predicate 'match': %v",
			err)
		return false
	}

	str, err := tmpl.GetTemplatedString(ctx, "string", f.String)
	if err != nil {
		return false // already logged in tmpl
	}

	if f.Fixed != "" {
		log.Debug().Str("fixed", f.Fixed).Msg("fixed: " + f.Fixed)
		return f.Fixed == str
	}
	if f.Regexp != "" {
		log.Debug().Str("regexp", f.Regexp).Msg("regexp: " + f.Regexp)

		if r, err := regexp.Compile(f.Regexp); err != nil {
			log.Error().Err(err).Msgf("Bad regexp: %v", err)
			return false
		} else {
			if res := r.FindStringSubmatch(str); len(res) == 0 {
				return false
			} else {
				log.Debug().Msgf("Regexp Matched: %v", res)
				results := make(map[string]interface{})
				results["matches"] = res
				ctx.Results = results
				return true
			}
		}
	}

	log.Error().Msg("Missing one of 'fixed' or 'regexp' for predicate 'match'")
	return false
}
