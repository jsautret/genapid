package jsonrpcpredicate

import (
	"encoding/base64"
	"fmt"

	"github.com/jsautret/genapid/ctx"
	"github.com/jsautret/genapid/genapid"
	"github.com/rs/zerolog"
	"github.com/ybbus/jsonrpc"
)

// Name of the predicate
var Name = "jsonrpc"

// Predicate is a genapid.Predicate interface that describes the predicate
type Predicate struct {
	name   string
	params struct { // Params accepted by the predicate
		URL       string      `validate:"required,url"`
		Procedure string      `validate:"required"`
		Params    interface{} `mapstructure:",omitempty"`
		BasicAuth *basicAuth  `mapstructure:"basic_auth,omitempty"`
	}
	results ctx.Result // response of of jsonrpc server
}

type basicAuth struct{ Username, Password string }

// Call evaluates the predicate
func (predicate *Predicate) Call(log zerolog.Logger, c *ctx.Ctx) bool {
	p := predicate.params
	log = log.With().Str("procedure", p.Procedure).Logger()
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
	rpcClient := jsonrpc.NewClientWithOpts(p.URL, &opts)
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
	predicate.results = ctx.Result{"response": result}

	return true
}

// Try to convert params to something that can be marshalled in json
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
