package stgin

import (
	"reflect"
	"testing"
)

func TestMatchAndExtractPathParams(t *testing.T) {
	pattern := "/users/$username:string/purchases/$id:int?age"
	dummyRoute := GET(pattern, nil)
	regex, compileErr := getPatternCorrespondingRegex(dummyRoute.Path)
	if compileErr != nil {
		t.Errorf("could not compile '%s' as a valid uri pattern", normalizePath(pattern))
	}
	dummyRoute.correspondingRegex = regex
	uri := "/users/John/purchases/675?age=23"
	params, matches := MatchAndExtractPathParams(&dummyRoute, uri)
	expected := Params{
		"username": "John",
		"id":       "675",
	}
	if !matches || !reflect.DeepEqual(expected, params) {
		t.Error("path params do not follow the expected pattern")
	}
}
