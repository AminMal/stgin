package stgin

import (
	"github.com/AminMal/slogger"
	"runtime"
	"strings"
)

var stginLogger slogger.Logger

var contentTypeKey = "Content-Type"
var applicationJsonType = "application/json"

func init() {
	stginLogger = slogger.GetLogger("STGIN")
}

func relevantCaller() []runtime.Frame {
	pc := make([]uintptr, 16)
	n := runtime.Callers(1, pc)
	frames := runtime.CallersFrames(pc[:n])
	var fs []runtime.Frame
	for {
		f, more := frames.Next()
		if more && !strings.HasPrefix(f.Function, "github.com/AminMal/stgin") {
			fs = append(fs, f)
		} else { break }
	}
	return fs
}