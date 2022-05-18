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
	expectedQueries    queryDecl
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
	if api == nil {
		printStacktrace("")
		panic("cannot use nil as an API action")
	}
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
	return mkRoute(stage.path, api, strings.ToUpper(stage.method))
}

func OnPath(path string) RouteCreationStage {
	return RouteCreationStage{path: path}
}

func (stage RouteCreationStage) WithMethod(method string) RouteCreationStage {
	stage.method = method
	return stage
}
