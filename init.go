package stgin

import (
	"github.com/AminMal/slogger"
	"regexp"
	"runtime"
	"strings"
)

var stginLogger slogger.ConsoleLogger

var contentTypeKey = "Content-Type"
var applicationJsonType = "application/json"
var multipleSlashesRegex *regexp.Regexp = regexp.MustCompile("(/{2,})")

func init() {
	stginLogger = slogger.NewConsoleLogger("STGIN")
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
			if !strings.HasPrefix(f.Function, "github.com/AminMal/stgin.") {
				fs = append(fs, f)
			}
		} else {
			break
		}
	}
	return fs
}
