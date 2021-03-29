// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package plugins

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

var logLevel = zerolog.FatalLevel

func TestPlugins(t *testing.T) {
	zerolog.SetGlobalLevel(logLevel)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().Caller().Timestamp().Logger()

	assert.Equal(t, len(available), len(List()), "List")
	assert.Greater(t, len(List()), 0, "No plugins registered")
	for n := range available {
		p := Get(n)
		assert.NotEmpty(t, p.Name(), "empty predicate name")
	}
}
