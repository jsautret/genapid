package conf

import (
	"io/ioutil"
	"reflect"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v3"
)

type Root []Pipe
type Pipe struct {
	Name string
	Pipe []Predicate
}
type Predicate map[string]yaml.Node
type Params map[string]interface{}

func Read(filename string) Root {
	log.Info().Str("filename", filename).Msg("Reading configuration file")
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	conf := Root{}
	if err := yaml.Unmarshal(source, &conf); err != nil {
		log.Fatal().Err(err).Msg("")
	}
	return conf
}

func GetParams(ctx *ctx.Ctx, config Params, params interface{}) bool {
	c := mapstructure.DecoderConfig{
		DecodeHook: hookGval(ctx),
		Result:     params,
	}
	if decode, err := mapstructure.NewDecoder(&c); err != nil {
		log.Error().Err(err).Msg("Decoder error")
		return false
	} else if err := decode.Decode(config); err != nil {
		log.Error().Err(err).Msg("Incorrect fields for predicate")
		return false
	}
	return true
}

func hookGval(c *ctx.Ctx) func(from, to reflect.Type, data interface{}) (interface{}, error) {
	return func(from, to reflect.Type, data interface{}) (interface{}, error) {
		log.Trace().Interface("hook data", data).Msg("")
		log.Trace().Msgf("hook from: %v", from.Kind())
		log.Trace().Msgf("hook to: %v", to.Kind())
		if from.Kind() == reflect.String {
			return convertGval(data.(string), c)
		}
		if to.Kind() == reflect.Interface &&
			(from.Kind() == reflect.Map || from.Kind() == reflect.Slice) {
			// data will not be traversed by mapstructure,
			// so we do it here
			r := convert(data, c)
			log.Trace().Msgf("hook translated: %v", r)
			return r, nil
		}
		return data, nil
	}
}
