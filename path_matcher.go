package stgin

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var intRegex = "[0-9]+"
var floatRegex = "[+\\-]?(?:(?:0|[1-9]\\d*)(?:\\.\\d*)?|\\.\\d+)(?:\\d[eE][+\\-]?\\d+)?"
var stringRegex = "[a-zA-Z0-9_-]+"
var getPathParamSpecificationRegex = regexp.MustCompile("^(\\$[a-zA-Z0-9_-]+(:[a-z]{1,6})?)$")

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

type pathMatcher struct {
	key                string
	tpe                string
	correspondingRegex *regexp.Regexp
	rawRegex           string
}

type Params = []Param

type Param struct {
	key   string
	value string
}

func getMatcher(key, tpe string) *pathMatcher {
	var correspondingRegex string
	switch tpe {
	case "int":
		correspondingRegex = fmt.Sprintf("(?P<%s>%s)", key, intRegex)
	case "float":
		correspondingRegex = fmt.Sprintf("(?P<%s>%s)", key, floatRegex)
	default:
		correspondingRegex = fmt.Sprintf("(?P<%s>%s)", key, stringRegex)
	}
	return &pathMatcher{
		key:                key,
		tpe:                tpe,
		correspondingRegex: regexp.MustCompile(correspondingRegex),
		rawRegex:           correspondingRegex,
	}
}

func MatchAndExtractPathParams(pattern, uri string) ([]Param, bool) {
	portions := strings.Split(pattern, "/")
	rawPatternRegex := ""
	for i, portion := range portions {
		if portion != "" {
			isPathParamSpecification := getPathParamSpecificationRegex.Match([]byte(portion))
			if !isPathParamSpecification {
				rawPatternRegex += portion
			} else {
				keyAndType := strings.SplitN(portion, ":", 2)
				var key = trimFirstRune(keyAndType[0])
				var tpe string
				if len(keyAndType) == 1 {
					tpe = "string"
				} else {
					tpe = keyAndType[1]
				}
				matcher := getMatcher(key, tpe)
				rawPatternRegex += matcher.rawRegex
			}
		}
		if i != len(portions)-1 {
			rawPatternRegex += "/"
		}
	}
	regex, compileErr := regexp.Compile("^" + rawPatternRegex + "$")
	if compileErr != nil {
		return nil, false
	} else {
		if !regex.Match([]byte(uri)) {
			return nil, false
		} else {
			match := regex.FindStringSubmatch(uri)
			var res Params
			for i, name := range regex.SubexpNames() {
				if i != 0 && name != "" {
					res = append(res, Param{
						key:   name,
						value: match[i],
					})
				}
			}
			return res, true
		}
	}
}
