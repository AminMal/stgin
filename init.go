package stgin

import (
	"github.com/AminMal/slogger"
	"regexp"
	"runtime"
)

var stginLogger slogger.Logger

var contentTypeKey = "Content-Type"
var applicationJsonType = "application/json"
var multipleSlashesRegex *regexp.Regexp

func init() {
	stginLogger = slogger.GetLogger("STGIN")
	multipleSlashesRegex = regexp.MustCompile("(/{2,})")
}

func normalizePath(path string) string {
	return multipleSlashesRegex.ReplaceAllString(path, "/")
}

func relevantCallers() []runtime.Frame {
	pc := make([]uintptr, 16)
	n := runtime.Callers(1, pc)
	frames := runtime.CallersFrames(pc[:n])
	var fs []runtime.Frame
	for {
		f, more := frames.Next()
		if more {
			fs = append(fs, f)
		} else {
			break
		}
	}
	return fs
}
