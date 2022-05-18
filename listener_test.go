package stgin

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func welcomeAPI(RequestContext) Status {
	return Ok(Text("Welcome"))
}

func responseBodyModifier(response Status) Status {
	originalBody, _ := response.Entity.(textEntity)
	response.Entity = Text(originalBody.obj + " to the team!")
	return response
}

func responseStatusModifier(response Status) Status {
	response.StatusCode += 1
	return response
}

func responseAddHeadersListener(response Status) Status {
	response.Headers["x-test-listeners"] = []string{"true"}
	return response
}

func requestQueryModifier(request RequestContext) RequestContext {
	request.QueryParams.All["test"] = []string{"true"}
	return request
}

func requestHeaderModifier(request RequestContext) RequestContext {
	request.Headers["X-Test-Listeners"] = []string{"true"}
	return request
}

func TestResponseListeners(t *testing.T) {
	testController := NewController("Test", "/test")
	testController.AddRoutes(
		GET("/welcome", welcomeAPI),
	)
	testController.AddResponseListener(responseBodyModifier, responseStatusModifier, responseAddHeadersListener)
	uri, _ := url.Parse("/test/welcome")
	request := http.Request{
		Method:     "GET",
		URL:        uri,
		RequestURI: "/test/welcome",
	}
	result := testController.executeInternal(&request)
	if result.StatusCode != 201 {
		t.Fatal("response listener could not modify response status")
	}
	if text, ok := result.Entity.(textEntity); ok {
		if text.obj != "Welcome to the team!" {
			t.Fatal("response body modifier could not modify response body")
		}
	} else {
		t.Fatal("response content type changed in response body modifier")
	}
	// this is changed when actually writing headers by go http itself, to X-Test-Listeners (canonical name)
	if result.Headers["x-test-listeners"][0] != "true" {
		t.Fatal("response listener could not modify response headers")
	}
}

func TestRequestListeners(t *testing.T) {
	testController := NewController("Test", "/test")
	var testQueryValue string
	var testHeaderValue string
	testController.AddRoutes(GET("/req", func(request RequestContext) Status {
		fmt.Println("queryDecl: ", request.QueryParams, "headers: ", request.Headers)
		testQueryValue = request.QueryParams.MustGet("test")
		testHeaderValue = request.Headers.Get("X-Test-Listeners")
		return Ok(Text("Done"))
	}))
	testController.AddRequestListeners(requestHeaderModifier, requestQueryModifier)
	uri, _ := url.Parse("/test/req")
	request := http.Request{
		Method:     "GET",
		URL:        uri,
		RequestURI: "/test/req",
	}
	_ = testController.executeInternal(&request)
	if testQueryValue != "true" {
		fmt.Println(testQueryValue)
		t.Error("request listener failed to add query parameter")
	}
	if testHeaderValue != "true" {
		fmt.Println(testHeaderValue)
		t.Fatal("request listener could not append header")
	}
}
