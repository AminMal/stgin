package stgin

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

var ping Route = GET("/ping", func(_ RequestContext) Status {
	return Ok(&msg{Message: "PONG!"})
})

func TestNewController(t *testing.T) {
	controller := NewController("TestSuite")
	controller.SetRoutePrefix("test")

	if controller.Name != "TestSuite" || !controller.hasPrefix() || controller.prefix != "/test" {
		t.Error("controller configuration mismatch")
	}
}

func TestControllerListeners(t *testing.T) {
	var apiLogToTerminalString string
	var dummyQuery string
	var statusIncrementor ResponseListener = func(status Status) Status {
		status.StatusCode = status.StatusCode + 1
		return status
	}

	var addDummyQuery RequestListener = func(request RequestContext) RequestContext {
		request.QueryParams = map[string][]string{"dummy": {"yes"}}
		return request
	}
	var addApiLog APIListener = func(request RequestContext, status Status) {
		apiLogToTerminalString = fmt.Sprintf("request with path %v completed with status %d", request.Url, status.StatusCode)
		dummyQuery, _ = request.GetQuery("dummy")
	}

	controller := NewController("TestSuite")
	controller.SetRoutePrefix("test")
	controller.AddRoutes(ping)
	controller.AddRequestListeners(addDummyQuery)
	controller.AddResponseListener(statusIncrementor)
	controller.AddAPIListeners(addApiLog)

	uri, _ := url.Parse("/test/ping")
	rawRequest := http.Request{
		Method:           "GET",
		URL:              uri,
		RequestURI:       "/test/ping",
	}
	res := controller.executeInternal(&rawRequest)
	if res.StatusCode != 201 {
		t.Error("response listener could not mutate api response")
	}
	expectedLog := fmt.Sprintf("request with path %v completed with status %d", "/test/ping", 201)
	if apiLogToTerminalString != expectedLog {
		t.Fatalf("api listener did not work properly")
	}

	if dummyQuery != "yes" {
		t.Fatalf("request listener did not work properly")
	}
}

func BenchmarkPing(b *testing.B) {
	controller := NewController("Test")
	controller.AddRoutes(ping)
	uri, _ := url.Parse("/ping")
	req := http.Request{
		Method:           "GET",
		URL:              uri,
		RequestURI:       "/ping",
	}
	for i := 0; i < b.N; i++ {
		controller.executeInternal(&req)
	}
}