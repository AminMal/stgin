package stgin

import "github.com/AminMal/slogger"

var stginLogger slogger.Logger

func init() {
	stginLogger = slogger.GetLogger("STGIN")
}
