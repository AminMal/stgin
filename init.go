package stgin

import "github.com/AminMal/slogger"

var stginLogger slogger.Logger

var contentTypeKey = "Content-Type"
var applicationJsonType = "application/json"

func init() {
	stginLogger = slogger.GetLogger("STGIN")
}
