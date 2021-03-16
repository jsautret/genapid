package tmpl

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"text/template"

	ctx "github.com/jsautret/go-api-broker/ctx"

	"github.com/Masterminds/sprig"
	"github.com/PaesslerAG/gval"
	jsonpathlib "github.com/PaesslerAG/jsonpath"
	fuzzysearch "github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rs/zerolog/log"
)

var t *template.Template

// GetTemplatedString processes templates in a string and returns the
// result
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
	t = template.New("string").Funcs(sprig.TxtFuncMap()).
		Funcs(template.FuncMap{
			"jsonpath": jsonpath,
			"fuzzy":    fuzzy,
		})
}

func jsonpath(path string, json interface{}) interface{} {
	builder := gval.Full(jsonpathlib.PlaceholderExtension())
	p, err := builder.NewEvaluable(path)
	if err != nil {
		log.Error().Err(err).Msg("Cannot evaluate gval")
		return `""`
	}
	res, err := p(context.Background(), json)
	if err != nil {
		log.Error().Err(err).Msg("Cannot evaluate jsonpath")
		return `""`
	}
	return res
}

func fuzzy(source string, targets []interface{}) string {
	matches := fuzzysearch.RankFindNormalizedFold(
		source, toStrings(targets))
	sort.Sort(matches)
	return matches[0].Target
}

func toStrings(l []interface{}) []string {
	ls := make([]string, len(l))
	for i := 0; i < len(l); i++ {
		e, ok := l[i].(string)
		if !ok {
			err := fmt.Errorf("element is not string: %v", l[i])
			log.Error().Err(err).Msg("")
			return []string{}
		}
		ls[i] = e
	}
	return ls
}
