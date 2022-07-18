package stgin

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var getPathParamSpecificationRegex = regexp.MustCompile("^(\\$[a-zA-Z0-9_-]+(:[a-z]{1,6})?)$")

type regexHolder struct {
	rawRegex 		string
	compiledRegex   *regexp.Regexp
}

func AddMatchingPattern(key string, rawPattern string) error {
	if key == "int" || key == "string" || key == "float" || key == "uuid" {
		return errors.New("cannot modify basic matching matchers")
	}
	regex, regexCompileErr := regexp.Compile(rawPattern)
	if regexCompileErr != nil { return regexCompileErr }
	matchers[key] = &regexHolder{
		rawRegex:      rawPattern,
		compiledRegex: regex,
	}
	return nil
}

type Params = map[string]string

func getMatcherRawRegex(key, tpe string) string {
	pattern := matchers[tpe]
	regexStr := stringRegexStr
	if pattern != nil { regexStr = pattern.rawRegex }
	return fmt.Sprintf("(?P<%s>%s)", key, regexStr)
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
		params := make(map[string]string, 5)
		for i, name := range regex.SubexpNames() {
			if i != 0 && name != "" {
				params[name] = match[i]
			}
		}
		return params, true
	}
}
