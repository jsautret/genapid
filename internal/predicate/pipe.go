package predicate

import (
	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/rs/zerolog/log"
)

// ProcessPipe evaluate a pipe of predicates
func ProcessPipe(p *conf.Pipe, c *ctx.Ctx) bool {
	name := p.Name
	log := log.With().Str("pipe", name).Logger()

	log.Debug().Msgf("Processing pipe '%v'", name)

	var result bool
	for j := 0; j < len(p.Pipe); j++ {
		result = Process(log, &p.Pipe[j], c)
		if !result {
			break
		}

	}
	log.Debug().Bool("value", result).Msg("End pipe")
	// return result of last predicate
	// used for tests
	return result
}
