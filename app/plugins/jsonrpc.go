// +build !disable_jsonrpc

package plugins

import jsonrpcpredicate "github.com/jsautret/genapid/predicates/jsonrpc"

func init() {
	Add(jsonrpcpredicate.Name, jsonrpcpredicate.New)
}
