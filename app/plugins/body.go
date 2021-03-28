// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

// +build !disable_body

package plugins

import bodypredicate "github.com/jsautret/genapid/predicates/body"

func init() {
	Add(bodypredicate.Name, bodypredicate.New)
}
