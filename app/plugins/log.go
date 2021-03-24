// +build !disable_log

package plugins

import logpredicate "github.com/jsautret/genapid/predicates/log"

func init() {
	Add(logpredicate.Name, logpredicate.New)
}
