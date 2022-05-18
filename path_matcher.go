package stgin

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	intRegexStr       = "[0-9]+"
	floatRegexStr     = "[+\\-]?(?:(?:0|[1-9]\\d*)(?:\\.\\d*)?|\\.\\d+)(?:\\d[eE][+\\-]?\\d+)?"
	stringRegexStr    = "[a-zA-Z0-9_!@#$%^&*()+=-]+"
	expectQueryParams = "(\\?.*)?"
)

var intRegex = regexp.MustCompile(intRegexStr)
var floatRegex = regexp.MustCompile(floatRegexStr)
var stringRegex = regexp.MustCompile(stringRegexStr)

var getPathParamSpecificationRegex = regexp.MustCompile("^(\\$[a-zA-Z0-9_-]+(:[a-z]{1,6})?)$")

type pathMatcher struct {
	key                string
	tpe                string
	correspondingRegex *regexp.Regexp
	rawRegex           string
}

type Params = map[string]string

func getMatcher(key, tpe string) *pathMatcher {
	var correspondingRegex string
	switch tpe {
	case "int":
		correspondingRegex = fmt.Sprintf("(?P<%s>%s)", key, intRegexStr)
	case "float":
		correspondingRegex = fmt.Sprintf("(?P<%s>%s)", key, floatRegexStr)
	default:
		correspondingRegex = fmt.Sprintf("(?P<%s>%s)", key, stringRegexStr)
	}
	return &pathMatcher{
		key:                key,
		tpe:                tpe,
		correspondingRegex: regexp.MustCompile(correspondingRegex),
		rawRegex:           correspondingRegex,
	}
}

func getPatternCorrespondingRegex(pattern string) (*regexp.Regexp, error) {
	pattern = normalizePath(pattern)
	portions := strings.Split(pattern, "/")
	rawPatternRegex := ""
	for i, portion := range portions {
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
		if i != len(portions)-1 {
			rawPatternRegex += "/"
		}
	}
	regex, compileErr := regexp.Compile("^" + rawPatternRegex + expectQueryParams + "$")
	if compileErr != nil {
		return nil, errors.New(fmt.Sprintf("could not compile '%s' as a valid uri pattern", pattern))
	} else {
		return regex, nil
	}
}

func MatchAndExtractPathParams(route *Route, uri string) (Params, bool) {
	regex := route.correspondingRegex
	if !regex.Match([]byte(uri)) {
		return nil, false
	} else {
		match := regex.FindStringSubmatch(uri)
		var res Params = make(map[string]string, 5)
		for i, name := range regex.SubexpNames() {
			if i != 0 && name != "" {
				res[name] = match[i]
			}
		}
		return res, true
	}
}
