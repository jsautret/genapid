// +build !disable_jsonrpc

package plugins

import jsonrpcpredicate "github.com/jsautret/go-api-broker/predicates/jsonrpc"

func init() {
	p := jsonrpcpredicate.Get()
	Add(p.Name(), p)
}
