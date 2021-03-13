// +build !disable_jsonrpc

package plugins

import "github.com/jsautret/go-api-broker/predicates/jsonrpc"

func init() {
	Add("jsonrpc", jsonrpc.Call)
}
