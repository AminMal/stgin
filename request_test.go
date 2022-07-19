package stgin

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestRequestContext_GetQuery(t *testing.T) {
	rc := &RequestContext{
		url:         "/test",
		queryParams: Queries{map[string][]string{"q": {"search"}, "date": {"2022-19:D"}}},
		pathParams:  PathParams{Params{}},
		headers:     emptyHeaders,
		receivedAt:  time.Now(),
		method:      http.MethodGet,
	}
	q, found := rc.QueryParams().GetOne("q")
	if !found {
		t.Fatal("request context .GetQuery method failed")
	}
	if q != "search" {
		t.Fatal("request context fetched the wrong value for query parameters")
	}
}

func TestRequestContext_GetPathParam(t *testing.T) {
	pattern := "/test/$username:string/$uid:int"
	route := Route{
		Path:               pattern,
		Method:             "GET",
		correspondingRegex: getRoutePatternRegexOrPanic(pattern),
		controller:         defaultController,
	}
	uri, _ := url.Parse("/test/JohnDoe/14")
	req := http.Request{
		Method:     "GET",
		URL:        uri,
		Header:     emptyHeaders,
		RequestURI: "/test/JohnDoe/14",
	}
	accepts, pathParams := route.acceptsAndPathParams(&req)
	if !accepts {
		t.Fatal("route does not accept the input uri")
	}
	rc := requestContextFromHttpRequest(&req, nil, pathParams)
	username := rc.PathParams().MustGet("username")
	uid := rc.PathParams().MustGetInt("uid")
	if username != "JohnDoe" || uid != 14 {
		t.Fatal("request context failed to load path params correctly")
	}
}

type person struct {
	Name     string `json:"name" xml:"name"`
	LastName string `json:"last_name" xml:"last_name"`
	Age      int    `json:"age" xml:"age"`
}

var mockBody = person{
	Name:     "John",
	LastName: "Doe",
	Age:      22,
}

var mockRoute Route = Route{
	Path:               "/users/$username/purchases/$pid:int",
	Method:             "GET",
	correspondingRegex: getRoutePatternRegexOrPanic("/users/$username/purchases/$pid:int"),
	controller:         defaultController,
}

func TestRequestBody_SafeJSONInto(t *testing.T) {
	jsonBytes, _ := json.Marshal(&mockBody)
	body := RequestBody{
		underlying:      nil,
		underlyingBytes: jsonBytes,
		hasFilledBytes:  true,
	}
	var result person
	err := body.SafeJSONInto(&result)
	if err != nil {
		t.Error(err.Error())
	}
	if result != mockBody {
		t.Fatal("objects do not match after serialization/deserialization")
	}
}

func TestRequestBody_SafeXMLInto(t *testing.T) {
	jsonBytes, _ := xml.Marshal(&mockBody)
	body := RequestBody{
		underlying:      nil,
		underlyingBytes: jsonBytes,
		hasFilledBytes:  true,
	}
	var result person
	err := body.SafeXMLInto(&result)
	if err != nil {
		t.Error(err.Error())
	}
	if result != mockBody {
		t.Fatal("objects do not match after serialization/deserialization")
	}
}

func TestRequestContext_MustGetPathParam(t *testing.T) {
	req := mkDummyRequest("/users/John/purchases/27")
	accepts, pathParams := mockRoute.acceptsAndPathParams(req)
	if !accepts {
		t.Error("route did not accept given uri")
	}

	requestContext := requestContextFromHttpRequest(req, nil, pathParams)
	if requestContext.PathParams().MustGet("username") != "John" ||
		requestContext.PathParams().MustGetInt("pid") != 27 {
		t.Fatal("path params parse failure")
	}
}
