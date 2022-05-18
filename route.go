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
	dir                string
	expectedQueries    queries
}

func (route Route) isStaticDir() bool { return route.dir != "" }

func (route Route) acceptsAndPathParams(request *http.Request) (bool, Params) {
	var ok bool
	var params Params
	if request.Method == route.Method {
		params, ok = matchAndExtractPathParams(&route, request.URL.Path)
	}

	return ok, params
}

func getRoutePatternRegexOrPanic(pattern string) *regexp.Regexp {
	regex, err := getPatternCorrespondingRegex(pattern)
	if err != nil {
		panic(err)
	}
	return regex
}

func mkRoute(pattern string, api API, method string) Route {
	path, queryDefs := splitBy(pattern, "?")
	return Route{
		Path:            path,
		Method:          method,
		Action:          api,
		expectedQueries: getQueryDefinitionsFromPattern(queryDefs),
	}
}

func GET(pattern string, api API) Route {
	return mkRoute(pattern, api, "GET")
}

func PUT(pattern string, api API) Route {
	return mkRoute(pattern, api, "PUT")
}

func POST(pattern string, api API) Route {
	return mkRoute(pattern, api, "POST")
}

func DELETE(pattern string, api API) Route {
	return mkRoute(pattern, api, "DELETE")
}

func Prefix(path string) string {
	return normalizePath("/" + path + "/.*")
}

func PATCH(pattern string, api API) Route {
	return mkRoute(pattern, api, "PATCH")
}

func OPTIONS(pattern string, api API) Route {
	return mkRoute(pattern, api, "OPTIONS")
}

func StaticDir(pattern string, dir string) Route {
	return Route{
		Path:   normalizePath("/" + pattern + "/"),
		Method: "GET",
		dir:    dir,
	}
}

func Handle(method string, pattern string, api API) Route {
	return mkRoute(pattern, api, method)
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
		path:   path,
	}
}

func OnPUT(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "PUT",
		path:   path,
	}
}

func OnPOST(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "POST",
		path:   path,
	}
}

func OnDelete(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "DELETE",
		path:   path,
	}
}

func OnPatch(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "PATCH",
		path:   path,
	}
}

func OnOptions(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "OPTIONS",
		path:   path,
	}
}

func OnPath(path string) RouteCreationStage {
	return RouteCreationStage{path: path}
}

func (stage RouteCreationStage) WithMethod(method string) RouteCreationStage {
	stage.method = method
	return stage
}
