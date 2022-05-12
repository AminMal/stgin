package stgin

import (
	"reflect"
	"testing"
)

func TestMatchAndExtractPathParams(t *testing.T) {
	pattern := "/users/$username:string/purchases/$id:int"
	regex, compileErr := getPatternCorrespondingRegex(pattern)
	if compileErr != nil {
		t.Errorf("could not compile '%s' as a valid uri pattern", normalizePath(pattern))
	}
	dummyRoute := Route{Path: pattern, correspondingRegex: regex}
	uri := "/users/John/purchases/675?q=search"
	params, matches := MatchAndExtractPathParams(&dummyRoute, uri)
	var expected Params
	expected = append(expected, Param{"username", "John"})
	expected = append(expected, Param{"id", "675"})
	var expectedSlice []Param = expected
	var resultSlice []Param = params
	if !matches || !reflect.DeepEqual(expectedSlice, resultSlice) {
		t.Error("path params do not follow the expected pattern")
	}
}
