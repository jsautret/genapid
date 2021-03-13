// +build !disable_match

package plugins

import "github.com/jsautret/go-api-broker/predicates/match"

func init() {
	Add("match", match.Call)
}
