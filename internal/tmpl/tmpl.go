package tmpl

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"

	ctx "github.com/jsautret/go-api-broker/context"

	"github.com/Masterminds/sprig"
	"github.com/PaesslerAG/gval"
	jsonpathlib "github.com/PaesslerAG/jsonpath"
	"github.com/rs/zerolog/log"
)

var t *template.Template

func GetTemplatedString(ctx *ctx.Ctx, name, in string) (string, error) {
	log := log.With().Str("template", name).Logger()
	log.Trace().Str("in", in).Msg("")

	tmpl, err := t.Parse(in)
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
	log.Trace().Str("out", result).Msg("")
	return result, nil
}

func init() {
	t = template.New("string").Funcs(sprig.TxtFuncMap())
}

func jsonpath(json interface{}, path string) interface{} {
	builder := gval.Full(jsonpathlib.PlaceholderExtension())
	p, err := builder.NewEvaluable(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	res, err := p(context.Background(), json)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return res
}
