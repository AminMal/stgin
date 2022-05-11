package stgin

import (
	"reflect"
	"testing"
)

func TestMatchAndExtractPathParams(t *testing.T) {
	pattern := "/users/$username:string/purchases/$id:int"
	uri := "/users/John/purchases/675"
	params, matches := MatchAndExtractPathParams(pattern, uri)
	var expected Params
	expected = append(expected, Param{"username", "John"})
	expected = append(expected, Param{"id", "675"})
	var expectedSlice []Param = expected
	var resultSlice []Param = params
	if !matches || !reflect.DeepEqual(expectedSlice, resultSlice) {
		t.Error("path params do not follow the expected pattern")
	}
}


