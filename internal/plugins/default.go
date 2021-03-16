// +build !disable_default

package plugins

import defaultPredicate "github.com/jsautret/go-api-broker/predicates/default"

func init() {
	Add("default", defaultPredicate.Call)
}
