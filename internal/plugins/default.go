// +build !disable_default

package plugins

import defaultpredicate "github.com/jsautret/go-api-broker/predicates/default"

func init() {
	Add("default", defaultpredicate.Call)
}
