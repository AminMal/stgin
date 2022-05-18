package stgin

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var getPathParamSpecificationRegex = regexp.MustCompile("^(\\$[a-zA-Z0-9_-]+(:[a-z]{1,6})?)$")

type Params = map[string]string

func getMatcherRawRegex(key, tpe string) string {
	var rawRegex string
	switch tpe {
	case "int":
		rawRegex = fmt.Sprintf("(?P<%s>%s)", key, intRegexStr)
	case "float":
		rawRegex = fmt.Sprintf("(?P<%s>%s)", key, floatRegexStr)
	default:
		rawRegex = fmt.Sprintf("(?P<%s>%s)", key, stringRegexStr)
	}
	return rawRegex
}

func getPatternCorrespondingRegex(pattern string) (*regexp.Regexp, error) {
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
			rawPatternRegex += getMatcherRawRegex(key, tpe)
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

func matchAndExtractPathParams(route *Route, uri string) (Params, bool) {
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
