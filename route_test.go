package stgin

import (
	"net/http"
	"testing"
)

func shouldHavePanicked(t *testing.T) {
	if err := recover(); err == nil {
		t.Error("make route should've panicked for nil api action")
	}
}

func TestMkRoute(t *testing.T) {
	defer shouldHavePanicked(t)
	GET("/test", nil)
}

func TestEmptyRouteCreationStage(t *testing.T) {
	defer shouldHavePanicked(t)
	OnPath("/test/$username").WithMethod("GET").Do(nil)
}

func TestGET(t *testing.T) {
	helloAPI := func(RequestContext) Status {return Ok(Text("hello"))}
	route := GET("/hello", helloAPI)
	if route.Method != http.MethodGet {
		t.Fatal("GET method could not set the correct http method")
	}
}

func TestPUT(t *testing.T) {
	helloAPI := func(RequestContext) Status {return Ok(Text("hello"))}
	route := PUT("/hello", helloAPI)
	if route.Method != http.MethodPut {
		t.Fatal("GET method could not set the correct http method")
	}
}

func TestPOST(t *testing.T) {
	helloAPI := func(RequestContext) Status {return Ok(Text("hello"))}
	route := POST("/hello", helloAPI)
	if route.Method != http.MethodPost {
		t.Fatal("GET method could not set the correct http method")
	}
}

func TestPATCH(t *testing.T) {
	helloAPI := func(RequestContext) Status {return Ok(Text("hello"))}
	route := PATCH("/hello", helloAPI)
	if route.Method != http.MethodPatch {
		t.Fatal("GET method could not set the correct http method")
	}
}

func TestDELETE(t *testing.T) {
	helloAPI := func(RequestContext) Status {return Ok(Text("hello"))}
	route := DELETE("/hello", helloAPI)
	if route.Method != http.MethodDelete {
		t.Fatal("GET method could not set the correct http method")
	}
}

func TestMakeRouteNilAction(t *testing.T) {
	defer shouldHavePanicked(t)
	_ = mkRoute("/hello", nil, "DELETE")
}
