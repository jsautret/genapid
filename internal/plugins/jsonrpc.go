// +build !disable_jsonrpc

package plugins

import jsonrpcpredicate "github.com/jsautret/go-api-broker/predicates/jsonrpc"

func init() {
	Add(jsonrpcpredicate.Name, jsonrpcpredicate.New)
}
