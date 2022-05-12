package stgin

import (
	"net/http"
	"regexp"
	"strings"
)

type API = func(c RequestContext) Status

type Route struct {
	Path               string
	Method             string
	Action             API
	correspondingRegex *regexp.Regexp
	controller         *Controller
}

func (route Route) withPrefixPrepended(controllerPrefix string) Route {
	route.Path = controllerPrefix + route.Path
	return route
}

func (route Route) acceptsAndPathParams(request *http.Request) (ok bool, params Params) {
	if request.Method == route.Method {
		params, ok = MatchAndExtractPathParams(&route, request.URL.Path)
	}
	return
}

func getRoutePatternRegexOrPanic(pattern string) *regexp.Regexp {
	regex, err := getPatternCorrespondingRegex(pattern)
	if err != nil {
		panic(err)
	}
	return regex
}

func GET(path string, api API) Route {
	return Route{
		Path:   normalizePath(path),
		Method: "GET",
		Action: api,
	}
}

func PUT(path string, api API) Route {
	return Route{
		Path:   normalizePath(path),
		Method: "PUT",
		Action: api,
	}
}

func POST(path string, api API) Route {
	return Route{
		Path:   normalizePath(path),
		Method: "POST",
		Action: api,
	}
}

func DELETE(path string, api API) Route {
	return Route{
		Path:   normalizePath(path),
		Method: "DELETE",
		Action: api,
	}
}

func PATCH(path string, api API) Route {
	return Route{
		Path:   normalizePath(path),
		Method: "PATCH",
		Action: api,
	}
}

func OPTIONS(path string, api API) Route {
	return Route{
		Path:   normalizePath(path),
		Method: "OPTIONS",
		Action: api,
	}
}

func Handle(method string, path string, api API) Route {
	return Route{
		Path:   normalizePath(path),
		Method: method,
		Action: api,
	}
}

type RouteCreationStage struct {
	method string
	path   string
}

func (stage RouteCreationStage) Do(api API) Route {
	switch strings.ToUpper(stage.method) {
	case "GET":
		return GET(stage.path, api)
	case "PUT":
		return PUT(stage.path, api)
	case "POST":
		return POST(stage.path, api)
	case "DELETE":
		return DELETE(stage.path, api)
	case "PATCH":
		return PATCH(stage.path, api)
	default:
		return GET(stage.path, api)
	}
}

func OnGET(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "GET",
		path:   normalizePath(path),
	}
}

func OnPUT(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "PUT",
		path:   normalizePath(path),
	}
}

func OnPOST(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "POST",
		path:   normalizePath(path),
	}
}

func OnDelete(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "DELETE",
		path:   normalizePath(path),
	}
}

func OnPatch(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "PATCH",
		path:   normalizePath(path),
	}
}

func OnOptions(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "OPTIONS",
		path:   normalizePath(path),
	}
}

func OnPath(path string) RouteCreationStage {
	return RouteCreationStage{path: normalizePath(path)}
}

func (stage RouteCreationStage) WithMethod(method string) RouteCreationStage {
	stage.method = method
	return stage
}
