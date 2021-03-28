// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

// +build !disable_command

package plugins

import commandpredicate "github.com/jsautret/genapid/predicates/command"

func init() {
	Add(commandpredicate.Name, commandpredicate.New)
}
