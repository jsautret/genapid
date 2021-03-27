// +build !disable_command

package plugins

import commandpredicate "github.com/jsautret/genapid/predicates/command"

func init() {
	Add(commandpredicate.Name, commandpredicate.New)
}
