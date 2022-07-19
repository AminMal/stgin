package stgin

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type msg struct {
	Message string `json:"message"`
}

var ping Route = GET("/ping", func(_ *RequestContext) Response {
	return Ok(Json(&msg{Message: "PONG!"}))
})

func TestNewController(t *testing.T) {
	controller := NewController("TestSuite", "test")
	if controller.Name != "TestSuite" || controller.prefix != "/test/" {
		t.Error("controller configuration mismatch")
	}
}

func TestControllerListeners(t *testing.T) {
	var apiLogToTerminalString string
	var dummyQuery string
	var statusIncrementor = func(status Response) {
		status.StatusCode += 1
	}

	var addDummyQuery RequestModifier = func(request *RequestChangeable) {
		request.SetQueries("dummy", []string{"yes"})
	}
	var addApiLog ApiWatcher = func(request *RequestContext, status Response) {
		apiLogToTerminalString = fmt.Sprintf("request with path %s completed with status %d", request.Url(), status.StatusCode)
		dummyQuery = request.QueryParams().MustGet("dummy")
	}

	controller := NewController("TestSuite", "/test")
	controller.AddRoutes(ping)
	controller.AddRequestListeners(addDummyQuery)
	controller.AddResponseListener(statusIncrementor)
	controller.AddAPIListeners(addApiLog)

	uri, _ := url.Parse("/test/ping")
	rawRequest := http.Request{
		Method:     "GET",
		URL:        uri,
		RequestURI: "/test/ping",
	}
	res := controller.executeInternal(&rawRequest)
	if res.StatusCode != 201 {
		t.Fatalf("response listener could not mutate api response")
	}
	time.Sleep(200 * time.Millisecond) // since api listeners are now executed async, we should wait a little :)
	expectedLog := fmt.Sprintf("request with path %v completed with status %d", "/test/ping", 201)
	if apiLogToTerminalString != expectedLog {
		t.Fatalf("api listener did not work properly")
	}

	if dummyQuery != "yes" {
		t.Fatalf("request listener did not work properly")
	}

	if res.Entity.ContentType() != "application/json" {
		t.Fatal("content type is not as expected")
	}
}

func TestController_Timeout(t *testing.T) {
	controller := NewController("Timeout controller", "")
	timeConsumingTask := func(*RequestContext) Response {
		time.Sleep(1 * time.Second)
		return Ok(Empty())
	}
	timeout := 200 * time.Millisecond
	controller.AddRoutes(GET("/timeout", timeConsumingTask))
	controller.SetTimeout(timeout)
	result := controller.executeInternal(mkDummyRequest("/timeout"))
	if result.StatusCode != http.StatusRequestTimeout {
		t.Fatalf("controller could not interrupt")
	}
}

func BenchmarkPing(b *testing.B) {
	controller := NewController("Test", "")
	controller.AddRoutes(ping)
	uri, _ := url.Parse("/ping")
	req := http.Request{
		Method:     "GET",
		URL:        uri,
		RequestURI: "/ping",
	}
	for i := 0; i < b.N; i++ {
		controller.executeInternal(&req)
	}
}
