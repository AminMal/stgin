package stgin

import (
	"encoding/json"
	"fmt"
	"github.com/AminMal/slogger/colored"
	"net/http"
	"path"
	"time"
)

var defaultController *Controller = NewController("Default", "")

type Server struct {
	port                   int
	Controllers            []*Controller
	requestListeners       []RequestListener
	responseListeners      []ResponseListener
	apiListeners           []APIListener
	notFoundAction         API
	errorAction            ErrorHandler
}

func (server *Server) Register(controllers ...*Controller) {
	server.Controllers = append(server.Controllers, controllers...)
}

func (server *Server) AddRoutes(routes ...Route) {
	for _, c := range server.Controllers {
		if c == defaultController {
			c.AddRoutes(routes...)
			break
		}
	}
}

func (server *Server) CorsHandler(handler CorsHandler) {
	server.AddRoutes(OPTIONS(Prefix(""), func(RequestContext) Status {
		return Ok(Empty()).WithHeaders(http.Header{
			"Access-Control-Allow-Origin": handler.AllowOrigin,
			"Access-Control-Allow-Credentials": handler.AllowCredentials,
			"Access-Control-Allow-Headers": handler.AllowHeaders,
			"Access-Control-Allow-Methods": handler.AllowMethods,
		})
	}))
}

func (server *Server) AddRequestListeners(listeners ...RequestListener) {
	server.requestListeners = append(server.requestListeners, listeners...)
}

func (server *Server) AddResponseListeners(listeners ...ResponseListener) {
	server.responseListeners = append(server.responseListeners, listeners...)
}

func (server *Server) AddAPIListeners(listeners ...APIListener) {
	server.apiListeners = append(server.apiListeners, listeners...)
}

func (server *Server) NotFoundAction(action API) {
	server.notFoundAction = action
}

func (server *Server) SetErrorHandler(action ErrorHandler) {
	server.errorAction = action
}

type msg struct {
	Message string `json:"message"`
}

func translate(
	api API,
	requestListeners []RequestListener,
	responseListeners []ResponseListener,
	apiListeners []APIListener,
	recovery ErrorHandler,
	pathParams Params,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		queries := make(map[string][]string, 10)
		for key, value := range request.URL.Query() {
			queries[key] = value
		}

		rc := requestContextFromHttpRequest(request, writer, pathParams)

		for _, requestListener := range requestListeners {
			rc = requestListener(rc)
		}
		defer func() {
			if err := recover(); err != nil {
				if recovery == nil {
					panic(err)
				} else {
					status := recovery(rc, err)
					write(status, writer)
				}
			}
		}()

		result := api(rc)
		for _, responseListener := range responseListeners {
			result = responseListener(result)
		}
		now := time.Now()
		result.doneAt = now

		result.complete(request, writer)

		for _, apiListener := range apiListeners {
			apiListener(rc, result)
		}
	}
}

func getColor(status int) colored.Color {
	switch {
	case status > 100 && status < 300:
		return colored.GREEN
	case status >= 300 && status < 500:
		return colored.YELLOW
	case status >= 500:
		return colored.RED
	default:
		return colored.CYAN
	}
}

var WatchAPIs APIListener = func(request RequestContext, status Status) {
	difference := fmt.Sprint(status.doneAt.Sub(request.receivedAt))
	statusString := fmt.Sprintf("%v%d%v", getColor(status.StatusCode), status.StatusCode, colored.ResetPrevColor)
	_ = stginLogger.InfoF("%v -> %v\t\t| %v | %v", request.Method, request.Url, statusString, difference)
}

type generalFailureMessage struct {
	StatusCode int    `json:"status_code"`
	Path       string `json:"path"`
	Message    string `json:"message"`
	Method     string `json:"method"`
}

var notFoundDefaultAction API = func(request RequestContext) Status {
	return NotFound(Json(&generalFailureMessage{
		StatusCode: 404,
		Path:       request.Url,
		Message:    "route not found",
		Method:     request.Method,
	}))
}

var errorAction ErrorHandler = func(request RequestContext, err any) Status {
	callers := relevantCallers()
	var stacktrace = fmt.Sprintf("recovering following error: %v%v%v\n", colored.RED, fmt.Sprint(err), colored.ResetPrevColor)
	for _, caller := range callers {
		stacktrace += fmt.Sprintf("\tIn: %s (%s:%d)\n", caller.Function, path.Base(caller.File), caller.Line)
	}
	if parseErr, isParseError := err.(ParseError); isParseError {
		return BadRequest(Json(&generalFailureMessage{
			StatusCode: 400,
			Path:       request.Url,
			Message:    parseErr.Error(),
			Method:     request.Method,
		}))
	}
	_ = stginLogger.Err(stacktrace)
	return InternalServerError(Json(&generalFailureMessage{
		StatusCode: 500,
		Path:       request.Url,
		Message:    "internal server error",
		Method:     request.Method,
	}))
}

type apiHandler struct {
	methodWithRoutes map[string][]Route // to optimize request matching time
	server           *Server
}

func (handler apiHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var done bool
	for method, routes := range handler.methodWithRoutes {
		if method != request.Method {
			continue
		}
		// method matches
		for _, route := range routes {
			accepts, pathParams := route.acceptsAndPathParams(request)
			if accepts && acceptsAllQueries(route.expectedQueries, request.URL.Query()) {
				requestListeners := append(handler.server.requestListeners, route.controller.requestListeners...)
				responseListeners := append(handler.server.responseListeners, route.controller.responseListeners...)
				apiListeners := append(handler.server.apiListeners, route.controller.apiListeners...)
				handlerFunc := translate(
					route.Action,
					requestListeners,
					responseListeners,
					apiListeners,
					handler.server.errorAction,
					pathParams,
				)
				handlerFunc(writer, request)
				done = true
				break
			}
		}
	}
	// no route matches the request
	if !done {
		rc := requestContextFromHttpRequest(request, writer, nil)
		status := handler.server.notFoundAction(rc)
		statusCode := status.StatusCode
		bodyBytes, contentType, marshalErr := marshall(status.Entity)
		if marshalErr != nil {
			_ = stginLogger.ErrorF(
				"could not marshal not found action result:\n\t%v%v%v",
				colored.RED, fmt.Sprint(marshalErr), colored.ResetPrevColor,
			)
			bodyBytes, _ = json.Marshal(&generalFailureMessage{
				StatusCode: http.StatusNotFound,
				Path:       request.URL.Path,
				Message:    "route not found",
				Method:     request.Method,
			})
			statusCode = http.StatusNotFound
			contentType = applicationJson
		}
		writer.Header().Set(contentTypeKey, contentType)
		writer.WriteHeader(statusCode)
		writer.Write(bodyBytes)
	}
}

func routeAppendLog(controllerName, method, path string) string {
	return fmt.Sprintf("Adding %v's API:\t%s%v%s\t\t-> %s%v%s",
		controllerName,
		colored.CYAN, method, colored.ResetPrevColor,
		colored.CYAN, path, colored.ResetPrevColor,
	)
}

func bindStaticDirLog(routePath string, dir string) string {
	return fmt.Sprintf("Binding route :\t%s%s%s\t\tto serve static directory -> %s%s%s",
		colored.CYAN, routePath, colored.ResetPrevColor,
		colored.GREEN, dir, colored.ResetPrevColor,
	)
}

func (server *Server) handler() http.Handler {
	mux := http.NewServeMux()
	methodWithRoutes := make(map[string][]Route)
	for _, controller := range server.Controllers {
		for _, route := range controller.routes {
			var log string
			if !route.isStaticDir() {
				route = route.withPrefixPrepended(controller.prefix)
				methodWithRoutes[route.Method] = append(methodWithRoutes[route.Method], route)
				log = routeAppendLog(controller.Name, route.Method, route.Path)
			} else {
				routePath := route.withPrefixPrepended(controller.prefix).Path
				dir := route.dir
				log = bindStaticDirLog(routePath, dir)
				mux.Handle(routePath, http.StripPrefix(routePath, http.FileServer(http.Dir(dir))))
			}
			_ = stginLogger.Info(log)
		}
	}
	mux.Handle("/", http.StripPrefix("", apiHandler{methodWithRoutes: methodWithRoutes, server: server}))
	return mux
}

func (server *Server) Start() error {
	_ = stginLogger.InfoF("starting server on port: %s%d%s", colored.YELLOW, server.port, colored.ResetPrevColor)
	return http.ListenAndServe(fmt.Sprintf(":%d", server.port), server.handler())
}

func NewServer(port int) *Server {
	return &Server{
		port:                   port,
		notFoundAction:         notFoundDefaultAction,
		errorAction:            nil,
		Controllers: 			[]*Controller{defaultController},
	}
}

func DefaultServer(port int) *Server {
	server := NewServer(port)
	server.AddAPIListeners(WatchAPIs)
	server.errorAction = errorAction
	return server
}
