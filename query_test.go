package stgin

import (
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"
)

type Q struct {
	Query    string `qp:"query"`
	Name     string `qp:"name"`
	Untagged string
	Age      int `qp:"age"`
}

func TestRequestContext_QueryToObj(t *testing.T) {
	uri, _ := url.Parse("/test/queryDecl?query=search&name=John&Untagged=used&age=29")
	req := http.Request{
		Method:     "GET",
		URL:        uri,
		Header:     emptyHeaders,
		RequestURI: "/test/queryDecl?query=search&name=John&Untagged=used&age=29",
	}
	rc := requestContextFromHttpRequest(&req, nil, Params{})
	emptyQuery := Q{
		Query: "search",
		Name:  "John",
	}
	err := rc.QueryToObj(&emptyQuery)
	if err != nil {
		t.Errorf("Failed creating query object: %s", err.Error())
	}
	if !reflect.DeepEqual(emptyQuery, Q{Name: "John", Query: "search", Untagged: "used", Age: 29}) {
		t.Fatal("failed")
	}
}

func mkDummyRequest(path string) *http.Request {
	uri, _ := url.Parse(path)
	return &http.Request{
		Method:     "GET",
		URL:        uri,
		Header:     emptyHeaders,
		RequestURI: path,
	}
}

func TestAcceptsAllQueries(t *testing.T) {
	pattern := "/test/queryDecl?query:string&name&age:int&email"
	dummyRoute := GET(pattern, func(_ *RequestContext) Response { return Ok(Empty()) })
	regex, compileError := getPatternCorrespondingRegex(dummyRoute.Path)
	if compileError != nil {
		t.Fatalf("could not compile route pattern: %s", dummyRoute.Path)
	}
	dummyRoute.correspondingRegex = regex
	expectedQueries := queryDecl{
		"query": "string",
		"name":  "string",
		"age":   "int",
		"email": "string",
	}
	if !reflect.DeepEqual(dummyRoute.expectedQueries, expectedQueries) {
		t.Errorf("query parser could not parse expected queryDecl in route pattern")
	}
	shouldAccept := "/test/queryDecl?query=search&name=John&age=23&support_extra=true&email=john.doe@gmail.com"
	shouldNotAccept := "/test/queryDecl?query=search&name=John&age=twenty_three&email=john.doe@gmail.com"
	shouldNotAccept2 := "/test/queryDecl?query=search&age=twenty_three&email=john.doe@gmail.com"
	empty := "/test/queryDecl"

	shouldAcceptRequest := mkDummyRequest(shouldAccept)
	shouldNotAcceptRequest := mkDummyRequest(shouldNotAccept)
	shouldNotAcceptRequest2 := mkDummyRequest(shouldNotAccept2)
	emptyRequest := mkDummyRequest(empty)

	if acceptsAllQueries(dummyRoute.expectedQueries, emptyRequest.URL.Query()) {
		t.Fatal("route accepted a request which should not have been accepted")
	}

	if !acceptsAllQueries(dummyRoute.expectedQueries, shouldAcceptRequest.URL.Query()) {
		t.Error("route did not accept a request which should've been accepted")
	}

	if acceptsAllQueries(dummyRoute.expectedQueries, shouldNotAcceptRequest.URL.Query()) {
		t.Error("route accepted a request which should not have been accepted")
	}

	if acceptsAllQueries(dummyRoute.expectedQueries, shouldNotAcceptRequest2.URL.Query()) {
		t.Error("route accepted a request which should not have been accepted")
	}
}

func TestAcceptsAllQueries2(t *testing.T) {
	pattern := "/showall?uid:int&username"
	dummyRoute := GET(pattern, func(request *RequestContext) Response {
		return Ok(Text(strconv.Itoa(request.QueryParams().MustGetInt(`uid`))))
	})
	regex, compileError := getPatternCorrespondingRegex(dummyRoute.Path)
	if compileError != nil {
		t.Fatalf("could not compile route pattern: %s", dummyRoute.Path)
	}
	dummyRoute.correspondingRegex = regex
	expectedQueries := queryDecl{
		"uid":      "int",
		"username": "string",
	}
	if !reflect.DeepEqual(dummyRoute.expectedQueries, expectedQueries) {
		t.Errorf("query parser could not parse expected queryDecl in route pattern")
	}
	shouldNotAccept := "/showall?uid=a23&username=John"
	shouldNotAcceptReq := mkDummyRequest(shouldNotAccept)
	if acceptsAllQueries(dummyRoute.expectedQueries, shouldNotAcceptReq.URL.Query()) {
		t.Fatal("WTFFFF")
	}
}
