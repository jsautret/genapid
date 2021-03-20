// +build !disable_log

package plugins

import logpredicate "github.com/jsautret/go-api-broker/predicates/log"

func init() {
	Add(logpredicate.Name, logpredicate.New)
}
