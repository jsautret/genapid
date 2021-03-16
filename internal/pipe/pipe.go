package pipe

import (
	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/jsautret/go-api-broker/internal/predicate"
	"github.com/rs/zerolog/log"
)

var defaultPipeName = "noname"

// Process a pipe of predicates
func Process(p conf.Pipe, c *ctx.Ctx) bool {
	name := p.Name
	if name == "" {
		name = defaultPipeName
	}
	log := log.With().Str("pipe", name).Logger()

	log.Debug().Msgf("Processing pipe '%v'", name)

	var result bool
	for j := 0; j < len(p.Pipe); j++ {
		result = predicate.Process(p.Pipe[j], c)
		if !result {
			break
		}

	}
	log.Debug().Bool("value", result).Msg("End pipe")
	return result
}
