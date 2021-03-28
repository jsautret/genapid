// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

// +build !disable_log

package plugins

import logpredicate "github.com/jsautret/genapid/predicates/log"

func init() {
	Add(logpredicate.Name, logpredicate.New)
}
