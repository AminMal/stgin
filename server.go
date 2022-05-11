package stgin

import (
	"encoding/json"
	"fmt"
	"github.com/AminMal/slogger/colored"
	"mime/multipart"
	"net/http"
	"path"
	"time"
)

type Server struct {
	port                   int
	Controllers            []*Controller
	requestListeners       []RequestListener
	responseListeners      []ResponseListener
	apiListeners           []APIListener
	notFoundAction         API
	methodNotAllowedAction API
	errorAction            ErrorHandler
}

func (server *Server) Register(controllers ...*Controller) {
	server.Controllers = append(server.Controllers, controllers...)
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

func (server *Server) MethodNowAllowedAction(action API) {
	server.methodNotAllowedAction = action
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

		url := request.URL.Path
		headers := request.Header
		body, err := bodyFromReadCloser(request.Body)
		rc := RequestContext{
			Url:           url,
			QueryParams:   queries,
			PathParams:    pathParams,
			Headers:       headers,
			Body:          body,
			receivedAt:    time.Now(),
			Method:        request.Method,
			ContentLength: request.ContentLength,
			Host:          request.Host,
			MultipartForm: func() *multipart.Form {
				return request.MultipartForm
			},
			Scheme:     request.URL.Scheme,
			RemoteAddr: request.RemoteAddr,
		}

		for _, requestListener := range requestListeners {
			rc = requestListener(rc)
		}
		defer func() {
			if err := recover(); err != nil {
				if recovery == nil {
					panic(err)
				} else {
					isr := recovery(rc, err)
					bodyBytes, contentType, _ := marshall(isr.Entity)
					writer.Header().Add(contentTypeKey, contentType)
					writer.WriteHeader(isr.StatusCode)
					_, _ = writer.Write(bodyBytes)
				}
			}
		}()

		result := api(rc)
		for _, responseListener := range responseListeners {
			result = responseListener(result)
		}

		for _, apiListener := range apiListeners {
			apiListener(rc, result)
		}

		if result.isRedirection() {
			location, _ := result.Entity.Bytes()
			http.Redirect(writer, request, string(location), result.StatusCode)
			return
		}
		statusCode := result.StatusCode
		writer.WriteHeader(result.StatusCode)
		responseBody, contentType, err := marshall(result.Entity)
		if err != nil {
			statusCode = http.StatusInternalServerError
			_ = stginLogger.ErrorF("error while marshalling request entity:\n\t%v", fmt.Sprintf("%s%s%s", colored.RED, err.Error(), colored.ResetPrevColor))
			ise := generalFailureMessage{
				Message:    "internal server error",
				Path:       request.URL.Path,
				StatusCode: statusCode,
				Method:     request.Method,
			}
			isrBytes, _ := json.Marshal(&ise)
			contentType = applicationJson
			responseBody = isrBytes
		} else {
			for _, cookie := range result.cookies {
				http.SetCookie(writer, cookie)
			}
			for key, values := range result.Headers {
				for _, value := range values {
					writer.Header().Add(key, value)
				}
			}
		}
		writer.WriteHeader(statusCode)
		if writer.Header().Get(contentTypeKey) == "" {
			writer.Header().Add(contentTypeKey, contentType)
		}
		_, err = writer.Write(responseBody)
		if err != nil {
			stginLogger.ErrorF("error while writing response to client:\n\t%s", fmt.Sprintf("%s%s%s", colored.RED, err.Error(), colored.ResetPrevColor))
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
	now := time.Now()
	difference := fmt.Sprint(now.Sub(request.receivedAt))
	statusString := fmt.Sprintf("%v%d%v", getColor(status.StatusCode), status.StatusCode, colored.ResetPrevColor)
	_ = stginLogger.InfoF("%v -> %v | %v | %v", request.Method, request.Url, statusString, difference)
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

var methodNotAllowedDefaultAction API = func(request RequestContext) Status {
	return MethodNotAllowed(Json(&generalFailureMessage{
		StatusCode: http.StatusMethodNotAllowed,
		Path:       request.Url,
		Message:    "method " + request.Method + " not allowed!",
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

type serverHandler struct {
	methodWithRoutes map[string][]Route // to optimize request matching time
	server           *Server
}

func (sh serverHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var done bool
	for method, routes := range sh.methodWithRoutes {
		if method != request.Method {
			continue
		}
		// method matches
		for _, route := range routes {
			accepts, pathParams := route.acceptsAndPathParams(request)
			if accepts {
				requestListeners := append(sh.server.requestListeners, route.controller.requestListeners...)
				responseListeners := append(sh.server.responseListeners, route.controller.responseListeners...)
				apiListeners := append(sh.server.apiListeners, route.controller.apiListeners...)
				handlerFunc := translate(
					route.Action,
					requestListeners,
					responseListeners,
					apiListeners,
					sh.server.errorAction,
					pathParams,
				)
				handlerFunc(writer, request)
				done = true
				break
			}
		}
	}
	if !done {
		body, _ := bodyFromReadCloser(request.Body)
		status := sh.server.notFoundAction(
			RequestContext{
				Url:           request.URL.Path,
				QueryParams:   request.URL.Query(),
				PathParams:    nil,
				Headers:       request.Header,
				Body:          body,
				receivedAt:    time.Now(),
				Method:        request.Method,
				ContentLength: request.ContentLength,
				Host:          request.Host,
				MultipartForm: func() *multipart.Form {
					return request.MultipartForm
				},
				Scheme:     request.URL.Scheme,
				RemoteAddr: request.RemoteAddr,
			},
		)
		statusCode := status.StatusCode
		bodyBtes, contentType, marshalErr := marshall(status.Entity)
		if marshalErr != nil {
			_ = stginLogger.ErrorF(
				"could not marshal not found action result:\n\t%v%v%v",
				colored.RED, fmt.Sprint(marshalErr), colored.ResetPrevColor,
			)
			bodyBtes, _ = json.Marshal(&generalFailureMessage{
				StatusCode: 404,
				Path:       request.URL.Path,
				Message:    "route not found",
				Method:     request.Method,
			})
			statusCode = 404
			contentType = applicationJson
		}
		writer.WriteHeader(statusCode)
		writer.Header().Add(contentTypeKey, contentType)
		writer.Write(bodyBtes)
	}
}

func (server *Server) handler() http.Handler {
	methodWithRoutes := make(map[string][]Route)
	for _, controller := range server.Controllers {
		for _, route := range controller.routes {
			var log string
			var r Route
			if controller.hasPrefix() {
				r = route.withPrefixPrepended(controller.prefix)
			} else {
				r = route
			}
			methodWithRoutes[route.Method] = append(methodWithRoutes[route.Method], r)
			log = fmt.Sprintf("Adding %v's API:\t%s%v%s -> %s%v%s",
				route.controller.Name,
				colored.CYAN, r.Method, colored.ResetPrevColor,
				colored.CYAN, r.Path, colored.ResetPrevColor,
			)

			_ = stginLogger.Info(log)
		}
	}
	return serverHandler{methodWithRoutes: methodWithRoutes, server: server}
}

func (server *Server) Start() error {
	_ = stginLogger.InfoF("starting server on port: %s%d%s", colored.YELLOW, server.port, colored.ResetPrevColor)
	return http.ListenAndServe(fmt.Sprintf(":%d", server.port), server.handler())
}

func (server *Server) Stop() {
	_ = stginLogger.Err("stopping server due to explicit stop call")
	panic("stopping server due to explicit stop call")
}

func NewServer(port int) *Server {
	return &Server{
		port:                   port,
		notFoundAction:         notFoundDefaultAction,
		methodNotAllowedAction: methodNotAllowedDefaultAction,
		errorAction:            nil,
	}
}

func DefaultServer(port int) *Server {
	server := NewServer(port)
	server.AddAPIListeners(WatchAPIs)
	server.errorAction = errorAction
	return server
}
