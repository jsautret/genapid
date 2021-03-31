// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

// Package conf provides access and convert data from configuration file
package conf

import (
	"errors"
	"io"
	"os"
	"reflect"

	"github.com/jsautret/genapid/app/utils"
	"github.com/jsautret/genapid/ctx"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v3"
)

// Root maps the main conf file
type Root []Pipe

// Pipe maps a pipe in conf file
type Pipe struct {
	Name    string
	Pipe    []Predicate
	Init    []Predicate
	Default ctx.Default
}

// Predicate maps a predicate in conf file
type Predicate map[string]yaml.Node

// Params contains name of predicate and its parameters as set in conf file
type Params struct {
	Name string
	Conf map[string]interface{}
}

// ReadFile reads the YAML config file and return Root config
func ReadFile(filename string) Root {
	log.Info().Str("filename", filename).Msg("Reading configuration file")

	handle, err := os.Open(filename)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	defer utils.CloseQuietly(handle)
	return Read(handle)
}

// Read reads the reader as YAML and return Root config
func Read(r io.Reader) Root {
	conf := Root{}
	d := yaml.NewDecoder(r)
	if err := d.Decode(&conf); err != nil {
		log.Fatal().Err(err).Msg("")
	}
	return conf
}

// AddDefault adds predicate default parameters to context
func AddDefault(log zerolog.Logger, c *ctx.Ctx, defaultConf *ctx.DefaultParams) bool {
	log.Debug().Interface("default", defaultConf).Msg("Setting default fields")
	// for each predicate
	for predicate, value := range *defaultConf {
		log.Trace().Interface("default", value).
			Msg("Setting default fields for " + predicate)
		if _, ok := c.Default[predicate]; !ok {
			// no default value yet for that predicate
			c.Default[predicate] = make(map[string]interface{})
		}
		def := c.Default[predicate]
		if !GetParams(c, value, &def) {
			log.Error().
				Err(errors.New("Invalid 'default' value")).
				Str("predicate", predicate).Msg("")
			return false
		}
		log.Trace().Interface("default", c.Default[predicate]).
			Str("predicate", predicate).Msg("'default'")
	}
	return true
}

// GetPredicateParams from default and from the conf, with Gval
// expressions evaluated
func GetPredicateParams(ctx *ctx.Ctx, config *Params, params interface{}) bool {
	// set predicate default parameters
	if !GetParams(ctx, ctx.Default[config.Name], params) {
		log.Error().Msg("Incorrect 'default' fields")
	}
	// set predicate parameters
	return GetParams(ctx, config.Conf, params)
}

// GetParams from a map & evaluate Gval expressions in it
func GetParams(ctx *ctx.Ctx, config interface{}, params interface{}) bool {
	log.Trace().Interface("in", config).Msg("Params conversion")
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
	log.Trace().Interface("out", params).Msg("Params conversion")
	return true
}

// Evaluate Gval expressions while mapping data to params
func hookGval(c *ctx.Ctx) func(from, to reflect.Type, data interface{}) (interface{}, error) {
	return func(from, to reflect.Type, data interface{}) (interface{}, error) {
		log.Trace().Interface("hook data", data).Msg("")
		if from.Kind() == reflect.String {
			return evaluateGval(data.(string), c)
		}
		if to.Kind() == reflect.Interface &&
			(from.Kind() == reflect.Map || from.Kind() == reflect.Slice) {
			// This data will not be traversed by mapstructure,
			// so we do it here
			r := convert(data, c)
			log.Trace().Msgf("hook translated: %v", r)
			return r, nil
		}
		return data, nil
	}
}
