package stgin

import (
	"github.com/AminMal/slogger"
	"regexp"
)

var stginLogger slogger.ConsoleLogger

var contentTypeKey = "Content-Type"
var multipleSlashesRegex *regexp.Regexp = regexp.MustCompile("(/{2,})")
var endsWithSlashRegex = regexp.MustCompile(".*/$")

const (
	intRegexStr       = "[0-9]+"
	floatRegexStr     = "[+\\-]?(?:(?:0|[1-9]\\d*)(?:\\.\\d*)?|\\.\\d+)(?:\\d[eE][+\\-]?\\d+)?"
	stringRegexStr    = "[a-zA-Z0-9_!@#$%^&*()+=-]+"
	expectQueryParams = "(\\?.*)?"
)

var intRegex *regexp.Regexp
var floatRegex *regexp.Regexp
var stringRegex *regexp.Regexp

func init() {
	intRegex = regexp.MustCompile(intRegexStr)
	floatRegex = regexp.MustCompile(floatRegexStr)
	stringRegex = regexp.MustCompile(stringRegexStr)

	stginLogger = slogger.NewConsoleLogger("STGIN")
}
