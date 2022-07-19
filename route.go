package stgin

import (
	"net/http"
	"regexp"
	"strings"
)

// API is the lowest-level functionality in stgin.
// It represents a function which takes a request, and generates an HTTP response.
type API = func(request *RequestContext) Response

// Route is a struct which specifies whether a request should be handled by the given Action inside the route.
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

// GET is a shortcut to define a route with http "GET" method.
func GET(pattern string, api API) Route {
	return mkRoute(pattern, api, "GET")
}

// PUT is a shortcut to define a route with http "PUT" method.
func PUT(pattern string, api API) Route {
	return mkRoute(pattern, api, "PUT")
}

// POST is a shortcut to define a route with http "POST" method.
func POST(pattern string, api API) Route {
	return mkRoute(pattern, api, "POST")
}

// DELETE is a shortcut to define a route with http "DELETE" method.
func DELETE(pattern string, api API) Route {
	return mkRoute(pattern, api, "DELETE")
}

// PATCH is a shortcut to define a route with http "PATCH" method.
func PATCH(pattern string, api API) Route {
	return mkRoute(pattern, api, "PATCH")
}

// OPTIONS is a shortcut to define a route with http "OPTIONS" method.
func OPTIONS(pattern string, api API) Route {
	return mkRoute(pattern, api, "OPTIONS")
}

// Prefix can be used as a pattern inside route definition, which matches all the requests that contain the given prefix.
// Note that this is appended to the corresponding controller's prefix in which the route is defined.
func Prefix(path string) string {
	return normalizePath("/" + path + "/.*")
}

// StaticDir can be used to server static directories.
// It's better to use StaticDir inside the server itself, or to have a dedicated controller for static directories you
// would want to serve.
func StaticDir(pattern string, dir string) Route {
	return Route{
		Path:   normalizePath("/" + pattern + "/"),
		Method: "GET",
		dir:    dir,
	}
}

// Handle is a generic function that can be used for other http methods that do not have a helper function in stgin (like GET).
func Handle(method string, pattern string, api API) Route {
	return mkRoute(pattern, api, method)
}

// RouteCreationStage is a struct that can make routes step by step.
// Is only returned after OnPath function is called.
type RouteCreationStage struct {
	method string
	path   string
}

// Do assign's the api action to the route creation stage, and returns the resulting route.
func (stage RouteCreationStage) Do(api API) Route {
	return mkRoute(stage.path, api, strings.ToUpper(stage.method))
}

// OnPath is the starting point of route creation stage, specifies the pattern.
func OnPath(path string) RouteCreationStage {
	return RouteCreationStage{path: path}
}

// WithMethod attaches the method to the route creation stage.
func (stage RouteCreationStage) WithMethod(method string) RouteCreationStage {
	stage.method = method
	return stage
}
