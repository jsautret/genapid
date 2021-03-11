package tmpl

import (
	"bytes"
	"text/template"

	"github.com/jsautret/go-api-broker/context"
	"github.com/rs/zerolog/log"
)

func GetTemplatedString(ctx *context.Ctx, name, in string) (string, error) {
	log := log.With().Str("template", name).Logger()
	log.Debug().Str("in", in).Msg("")

	tmpl, err := template.New(name).Parse(in)
	if err != nil {
		log.Error().Err(err).Msg("Cannot parse template")
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, ctx)
	if err != nil {
		log.Error().Err(err).Msg("Cannot execute template")
		return "", err

	}
	result := buf.String()
	log.Debug().Str("out", result).Msg("")
	return result, nil
}
