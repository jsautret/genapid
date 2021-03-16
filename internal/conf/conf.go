package conf

import (
	"errors"
	"io/ioutil"
	"reflect"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v3"
)

type Root []Pipe
type Pipe struct {
	Name    string
	Pipe    []Predicate
	Default ctx.Default
}
type Predicate map[string]yaml.Node
type Params struct {
	Name string
	Conf map[string]interface{}
}

// Read the YAML config file and return Root config
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

// Add predicate default parameters to context
func AddDefault(c *ctx.Ctx, defaultConf Params) {
	log.Debug().Interface("default", defaultConf).Msg("Setting default fields")
	// for each predicate
	for predicate, value := range defaultConf.Conf {
		if conf, ok := value.(map[string]interface{}); !ok {
			log.Error().
				Err(errors.New("'default' in not a dict")).
				Str("predicate", predicate).Msg("")
		} else {
			if _, ok = c.Default[predicate]; !ok {
				// no default value yet for that predicate
				c.Default[predicate] = make(map[string]interface{})
			}
			def := c.Default[predicate]
			if !GetParams(c, conf, &def) {
				log.Error().
					Err(errors.New("Invalid 'default' value")).
					Str("predicate", predicate).Msg("")
			}
		}
		log.Trace().Interface("default", c.Default[predicate]).
			Str("predicate", predicate).Msg("'default'")
	}
}

// Get predicate parameters from default and from the conf
func GetPredicateParams(ctx *ctx.Ctx, config Params, params interface{}) bool {
	// set predicate default parameters
	if !GetParams(ctx, ctx.Default[config.Name], params) {
		log.Error().Msg("Incorrect 'default' fields")
	}
	// set predicate parameters
	return GetParams(ctx, config.Conf, params)
}

func GetParams(ctx *ctx.Ctx, config map[string]interface{}, params interface{}) bool {
	c := mapstructure.DecoderConfig{
		DecodeHook: hookGval(ctx),
		ZeroFields: false, // needed for 'default' field
		Result:     params,
	}
	if decode, err := mapstructure.NewDecoder(&c); err != nil {
		log.Error().Err(err).Msg("Decoder error")
		return false
	} else if err := decode.Decode(config); err != nil {
		log.Error().Err(err).Msg("Incorrect fields")
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
