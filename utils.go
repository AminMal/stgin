package stgin

import (
	"fmt"
	"github.com/AminMal/slogger/colored"
	"path"
	"runtime"
	"strings"
	"unicode/utf8"
)

func splitBy(str string, token string) (string, string) {
	arr := strings.SplitN(str, token, 2)
	switch len(arr) {
	case 2:
		return arr[0], arr[1]
	case 1:
		return arr[0], ""
	default:
		return "", ""
	}
}

func getColor(status int) colored.Color {
	switch {
	case status > 100 && status < 300:
		return colored.GREEN
	case status >= 300 && status < 500:
		return colored.YELLOW
	case status >= 500:
		return colored.RED
	default:
		return colored.CYAN
	}
}

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
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

func printStacktrace(firstLine string) {
	callers := relevantCallers()
	var stacktrace string
	if firstLine != "" {
		stacktrace += "\n"
	}
	for _, caller := range callers {
		stacktrace += fmt.Sprintf("\tIn: %s (%s:%d)\n", caller.Function, path.Base(caller.File), caller.Line)
	}
	fmt.Print(stacktrace)
}

func normalizePath(path string) string {
	return multipleSlashesRegex.ReplaceAllString(path, "/")
}
