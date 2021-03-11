package tmpl

import (
	"bytes"
	"text/template"

	"github.com/jsautret/go-api-broker/context"
	"github.com/rs/zerolog/log"
)

func GetTemplatedString(ctx *context.Ctx, name, in string) (string, error) {
	log := log.With().Str("field", name).Logger()
	log.Debug().Str("template_in", in).
		Msgf("'%v' template: %v", name, in)

	tmpl, err := template.New(name).Parse(in)
	if err != nil {
		log.Error().Err(err).Msg("Cannot parse")
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, ctx)
	if err != nil {
		log.Error().Err(err).Msg("Cannot execute")
		return "", err

	}
	result := buf.String()
	log.Debug().Str("template_out", result).
		Msgf("'%v' value: %v", name, result)
	return result, nil
}
