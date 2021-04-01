// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package main

import (
	"errors"

	"github.com/jsautret/genapid/app/conf"
	"github.com/jsautret/genapid/app/predicate"
	"github.com/jsautret/genapid/ctx"
	"github.com/rs/zerolog/log"
)

type initValues []conf.Predicate

func processInit(cfg *conf.Root, c *ctx.Ctx) {
	if len(*cfg) == 0 {
		return
	}
	i := (*cfg)[0]
	init, ok := i["init"]
	if !ok {
		// not an init statement
		return
	}
	*cfg = (*cfg)[1:] // remove init from the conf
	if len(i) > 1 {
		log.Error().Err(errors.New("'init' must be used alone")).Msg("")
		return
	}
	predicates := initValues{}
	if err := init.Decode(&predicates); err != nil {
		log.Error().Err(errors.New("Invalid values for 'init'")).Msg("")
		return
	}
	log.Info().Msg("Processing init")
	for j := 0; j < len(predicates); j++ {
		result := predicate.Process(log.Logger, &predicates[j], c)
		if !result {
			break
		}
	}
}
