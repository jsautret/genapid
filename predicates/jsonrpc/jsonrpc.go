package jsonrpc

import (
	"regexp"

	"github.com/jsautret/go-api-broker/context"

	"github.com/jsautret/go-api-broker/internal/tmpl"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/smallfish/simpleyaml"
)

func Call(ctx *context.Ctx, conf *simpleyaml.Yaml) bool {
	debug := func() *zerolog.Event {
		return log.Debug().Str("predicate",
			"jsonrpc")
	}
	error := func() *zerolog.Event {
		return log.Error().Str("predicate",
			"jsonrpc")
	}

	if !conf.IsMap() {
		error().Msg("Malformed conf for predicate 'jsonrpc'")
		return false
	}

	url, err := conf.Get("url").String()
	if err != nil {
		error().Str("missing", "url").
			Msgf("Missing 'url' field for predicate 'match': %v",
				err)
		return false
	}
	basic_auth, err := conf.Get("url").String()
	if err != nil {
		error().Str("missing", "url").
			Msgf("Missing 'url' field for predicate 'match': %v",
				err)
		return false
	}

	toMatch, err = tmpl.GetTemplatedString(ctx, "string", toMatch)
	if err != nil {
		return false
	}

	if fixed, err := conf.Get("fixed").String(); err == nil {
		debug().Str("fixed", fixed).Msg("fixed: " + fixed)
		return fixed == toMatch
	}
	if re, err := conf.Get("regexp").String(); err == nil {
		debug().Str("regexp", re).Msg("regexp: " + re)

		if r, err := regexp.Compile(re); err != nil {
			error().Err(err).Msgf("Bad regexp: %v", err)
			return false
		} else {
			if res := r.FindStringSubmatch(toMatch); len(res) == 0 {
				return false
			} else {
				debug().Msgf("Regexp Matched: %v", res)
				results := make(map[string]interface{})
				results["matches"] = res
				ctx.Results = results
				return true
			}
		}
	}

	error().Msg("Missing mandatory field(s) for predicate 'match'")
	return false
}
