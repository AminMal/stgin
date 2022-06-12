package stgin

import (
	"github.com/AminMal/slogger"
	"regexp"
)

var stginLogger slogger.ConsoleLogger

var contentTypeKey = "Content-Type"
var multipleSlashesRegex *regexp.Regexp

const (
	intRegexStr       = "[-]?[0-9]+"
	floatRegexStr     = "[+\\-]?(?:(?:0|[1-9]\\d*)(?:\\.\\d*)?|\\.\\d+)(?:\\d[eE][+\\-]?\\d+)?"
	stringRegexStr    = "[a-zA-Z0-9_!@#$%^&*()+=-]+"
	expectQueryParams = "(\\?.*)?"
)

func init() {
	multipleSlashesRegex = regexp.MustCompile("(/{2,})")
	stginLogger = slogger.NewConsoleLogger("STGIN")
}
