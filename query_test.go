package stgin

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

type Q struct {
	Query    string `qp:"query"`
	Name     string `qp:"name"`
	Untagged string
}

func TestRequestContext_QueryToObj(t *testing.T) {
	uri, _ := url.Parse("/test/queries?query=search&name=John&Untagged=used")
	req := http.Request{
		Method:           "GET",
		URL:              uri,
		Header:           emptyHeaders,
		RequestURI:       "/test/queries?query=search&name=John&Untagged=used",
	}
	rc := requestContextFromHttpRequest(&req, nil, Params{})
	emptyQuery := Q{
		Query: "search",
		Name: "John",
	}
	err := rc.QueryToObj(&emptyQuery)
	if err != nil {
		t.Errorf("Failed creating query object: %s", err.Error())
	}
	if !reflect.DeepEqual(emptyQuery, Q{Name: "John", Query: "search", Untagged: "used"}) {
		t.Fatal("failed")
	}
}

func mkDummyRequest(path string) *http.Request {
	uri, _ := url.Parse(path)
	return &http.Request{
		Method:           "GET",
		URL:              uri,
		Header:           emptyHeaders,
		RequestURI:       path,
	}
}

func TestAcceptsAllQueries(t *testing.T) {
	pattern := "/test/queries?query:string&name&age:int"
	dummyRoute := GET(pattern, nil)
	regex, compileError := getPatternCorrespondingRegex(dummyRoute.Path)
	if compileError != nil {
		t.Fatalf("could not compile route pattern: %s", dummyRoute.Path)
	}
	dummyRoute.correspondingRegex = regex
	expectedQueries := queries{
		"query": "string",
		"name": "string",
		"age": "int",
	}
	if !reflect.DeepEqual(dummyRoute.expectedQueries, expectedQueries) {
		t.Errorf("query parser could not parse expected queries in route pattern")
	}
	shouldAccept := "/test/queries?query=search&name=John&age=23&support_extra=true"
	shouldNotAccept := "/test/queries?query=search&name=John&age=twenty_three"
	shouldAcceptRequest := mkDummyRequest(shouldAccept)
	shouldNotAcceptRequest := mkDummyRequest(shouldNotAccept)

	if !acceptsAllQueries(dummyRoute.expectedQueries, shouldAcceptRequest.URL.Query()) {
		t.Error("route did not accept a request which should've been accepted")
	}

	if acceptsAllQueries(dummyRoute.expectedQueries, shouldNotAcceptRequest.URL.Query()) {
		t.Error("route accepted a request which should not have been accepted")
	}
}
