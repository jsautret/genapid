package conf

import (
	"io/ioutil"
	"os"
	"reflect"

	"github.com/jsautret/go-api-broker/context"
	"github.com/jsautret/go-api-broker/internal/tmpl"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
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
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	conf := Root{}
	if err := yaml.Unmarshal(source, &conf); err != nil {
		log.Fatal().Err(err).Msg("")
	}
	log.Debug().Msgf("JJJJJ %v (%v)", conf[0].Pipe[0]["jsonrpc"], reflect.TypeOf(conf[0].Pipe[0]["jsonrpc"]))
	return conf
}

func init() {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	//log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func GetParams(ctx *context.Ctx, config Params, params interface{}) bool {
	c := mapstructure.DecoderConfig{
		DecodeHook: hookTemplate(ctx),
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

func hookTemplate(ctx *context.Ctx) func(from, to reflect.Type, data interface{}) (interface{}, error) {
	return func(from, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() == reflect.String {
			return tmpl.GetTemplatedString(ctx, from.Name(), data.(string))
		}
		return data, nil
	}
}
