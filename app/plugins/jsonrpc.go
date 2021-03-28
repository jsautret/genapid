// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

// +build !disable_jsonrpc

package plugins

import jsonrpcpredicate "github.com/jsautret/genapid/predicates/jsonrpc"

func init() {
	Add(jsonrpcpredicate.Name, jsonrpcpredicate.New)
}
