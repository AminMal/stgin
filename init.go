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
	uuidRegexStr      = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
)

var intQueryRegex = regexp.MustCompile("^" + intRegexStr + "$")
var floatQueryRegex = regexp.MustCompile("^" + floatRegexStr + "$")
var strQueryRegex = regexp.MustCompile(".*")
var uuidQueryRegex = regexp.MustCompile(uuidRegexStr)

var matchers map[string]*regexHolder

func init() {
	multipleSlashesRegex = regexp.MustCompile("(/{2,})")
	stginLogger = slogger.NewConsoleLogger("STGIN")
	matchers = map[string]*regexHolder {
		"int": {
			rawRegex:      		intRegexStr,
			compiledRegex: 		intQueryRegex,
		},
		"string": {
			rawRegex:		 	stringRegexStr,
			compiledRegex: 		strQueryRegex,
		},
		"float": {
			rawRegex: 			floatRegexStr,
			compiledRegex: 		floatQueryRegex,
		},
		"uuid": {
			rawRegex: uuidRegexStr,
			compiledRegex: uuidQueryRegex,
		},
	}
}
