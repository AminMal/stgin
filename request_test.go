package stgin

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestRequestContext_GetQuery(t *testing.T) {
	rc := RequestContext{
		Url:           "/test",
		QueryParams: map[string][]string{"q": {"search"}, "date": {"2022-19:D"}},
		PathParams:    Params{},
		Headers:       emptyHeaders,
		receivedAt:    time.Now(),
		Method:        "GET",
	}
	q, found := rc.GetQuery("q")
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
		Method:           "GET",
		URL:              uri,
		Header:           emptyHeaders,
		RequestURI: "/test/JohnDoe/14",
	}
	accepts, pathParams := route.acceptsAndPathParams(&req)
	if !accepts {
		t.Fatal("route does not accept the input uri")
	}
	rc := requestContextFromHttpRequest(&req, nil, pathParams)
	username := rc.MustGetPathParam("username")
	uid, _ := strconv.Atoi(rc.MustGetPathParam("uid"))
	if username != "JohnDoe" || uid != 14 {
		t.Fatal("request context failed to load path params correctly")
	}
}
