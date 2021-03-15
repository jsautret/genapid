package jsonrpc

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/rs/zerolog/log"
	"github.com/ybbus/jsonrpc/v2"
)

type params struct {
	Url       string
	Procedure string
	Params    interface{} `mapstructure:",omitempty"`
	BasicAuth *basicAuth  `mapstructure:"basic_auth,omitempty"`
}
type basicAuth struct {
	Username, Password string
}

func Call(ctx *ctx.Ctx, config conf.Params) bool {
	log := log.With().Str("predicate", "jsonrpc").Logger()

	var p params
	if !conf.GetParams(ctx, config, &p) {
		log.Error().Err(errors.New("Invalid params")).Msg("")
		return false
	}
	if p.Url == "" || p.Procedure == "" {
		log.Error().Err(errors.New("Missing parameters")).Msg("")
		return false
	}
	opts := jsonrpc.RPCClientOpts{}
	if p.BasicAuth != nil {
		log.Debug().Msg("Enabling basic auth")
		auth := map[string]string{
			"Authorization": "Basic " +
				base64.StdEncoding.EncodeToString([]byte(
					p.BasicAuth.Username+":"+
						p.BasicAuth.Password)),
		}
		opts = jsonrpc.RPCClientOpts{
			CustomHeaders: auth,
		}
	}
	rpcClient := jsonrpc.NewClientWithOpts(p.Url, &opts)
	var result interface{}
	var err error
	if p.Params != nil {
		err = rpcClient.CallFor(&result, p.Procedure, getParams(p.Params))
	} else {
		err = rpcClient.CallFor(&result, p.Procedure)
	}
	if err != nil {
		log.Warn().Err(err).Msg("jsonrpc call error")
		return false
	}
	log.Debug().Interface("result", result).Msg("Server response")
	ctx.Results["response"] = result
	return true
}

// Try to convert params to something that be marshalled in json
func getParams(p interface{}) interface{} {
	if params, ok := p.(map[interface{}]interface{}); ok {
		mapString := make(map[string]interface{})
		for key, value := range params {
			strKey := fmt.Sprintf("%v", key)
			mapString[strKey] = value
		}
		return mapString
	}
	return p
}
