package conf

import (
	"io/ioutil"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v3"
)

type Root []Pipe
type Pipe struct {
	Name string
	Pipe []Predicate
}
type Predicate map[string]interface{}

func Read(filename string) Root {
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

func init() {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	//log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
