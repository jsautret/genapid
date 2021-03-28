// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

// +build !disable_http

package plugins

import httppredicate "github.com/jsautret/genapid/predicates/http"

func init() {
	Add(httppredicate.Name, httppredicate.New)
}
