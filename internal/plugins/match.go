// +build !disable_match

package plugins

import matchpredicate "github.com/jsautret/go-api-broker/predicates/match"

func init() {
	Add(matchpredicate.Name, matchpredicate.New)
}
