// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package commandpredicate

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/jsautret/genapid/ctx"
	"github.com/jsautret/genapid/genapid"
	"github.com/rs/zerolog"
)

// Name of the predicate
var Name = "command"

// Predicate is a genapid.Predicate interface that describes the predicate
type Predicate struct {
	name   string
	params struct { // Params accepted by the predicate
		Command    string `validate:"required" mod:"path"`
		Chdir      string `mod:"path"`
		Args       []string
		Stdin      string
		Background bool
	}
	results ctx.Result // results of the command
}

// Call evaluates the predicate
func (predicate *Predicate) Call(log zerolog.Logger, c *ctx.Ctx) bool {
	p := predicate.params

	log.Debug().Str("Command", p.Command).Msg("")
	log.Debug().Str("Chdir", p.Chdir).Msg("")

	cmd := exec.Command(p.Command, p.Args...)

	if p.Chdir != "" {
		cmd.Dir = p.Chdir
	}

	if p.Background {
		if err := cmd.Start(); err != nil {
			log.Error().Err(err).Msg("Cannot start command")
			return false
		}
		return true
	}
	if p.Stdin != "" {
		cmd.Stdin = strings.NewReader(p.Stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	r := ctx.Result{"rc": 0}
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			rc := exitError.ExitCode()
			log.Debug().Int("rc", rc)
			r["rc"] = rc
		} else {
			log.Error().Err(err).Msg("Cannot run command")
			return false
		}
	}
	r["stdout"] = stdout.String()
	r["stderr"] = stderr.String()

	predicate.results = r
	return r["rc"] == 0
}

// Generic interface //

// Result returns data set by the predicate
func (predicate *Predicate) Result() ctx.Result {
	return predicate.results
}

// Name returns the name of the predicate
func (predicate *Predicate) Name() string {
	return predicate.name
}

// Params returns a reference to a struct params accepted by the predicate
func (predicate *Predicate) Params() interface{} {
	return &predicate.params
}

// New returns a new Predicate
func New() genapid.Predicate {
	return &Predicate{
		name: Name,
	}
}
