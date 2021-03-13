// +build !disable_log

package plugins

import "github.com/jsautret/go-api-broker/predicates/log"

func init() {
	Add("log", log.Call)
}
