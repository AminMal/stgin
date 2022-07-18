package stgin

import (
	"reflect"
	"testing"
)

func TestMatchAndExtractPathParams(t *testing.T) {
	pattern := "/users/$username:string/purchases/$id:int?age"
	var dummyAPI API = func(_ RequestContext) Status { return Ok(Empty()) }
	dummyRoute := GET(pattern, dummyAPI)
	regex, compileErr := getPatternCorrespondingRegex(dummyRoute.Path)
	if compileErr != nil {
		t.Errorf("could not compile '%s' as a valid uri pattern", normalizePath(pattern))
	}
	dummyRoute.correspondingRegex = regex
	uri := "/users/John/purchases/675?age=23"
	params, matches := matchAndExtractPathParams(&dummyRoute, uri)
	expected := Params{
		"username": "John",
		"id":       "675",
	}
	if !matches || !reflect.DeepEqual(expected, params) {
		t.Error("path params do not follow the expected pattern")
	}
}

func TestAddMatchingPattern(t *testing.T) {
	startsWithJohnRegex := "^john.*"
	err := AddMatchingPattern("john", startsWithJohnRegex)
	if err != nil {
		t.Fatal("compile error for valid regex")
	}
	validQueryOrPathParam := "johnDoe"
	acceptsValidParam := acceptsAllQueries(
		map[string]string {
			"test": "john",
		},
		map[string][]string {
			"test": {validQueryOrPathParam},
		},
	)
	if !acceptsValidParam {
		t.Fatal("failed to accept valid query/path param")
	}
	invalidQueryOrParam := "doeJohn"
	acceptsInvalidParam := acceptsAllQueries(
		map[string]string {
			"test": "john",
		},
		map[string][]string {
			"test": {invalidQueryOrParam},
		},
	)
	if acceptsInvalidParam {
		t.Fatal("query/path parameter got accepted while should not have")
	}
}

func TestAddMatchingPattern_Panic(t *testing.T) {
	invalidRegex := "^some[.d2"
	err := AddMatchingPattern("invalid", invalidRegex)
	if err == nil {
		t.Fatal("invalid regex got accepted in patterns")
	}
}
