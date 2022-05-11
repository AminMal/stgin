package stgin

import (
	"encoding/json"
	"fmt"
	"github.com/AminMal/slogger/colored"
	"mime/multipart"
	"net/http"
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

		pp := make(map[string]string)

		for _, q := range pathParams {
			pp[q.key] = q.value
		}

		url := request.URL.Path
		headers := request.Header
		body, err := bodyFromReadCloser(request.Body)
		rc := RequestContext{
			Url:           url,
			QueryParams:   queries,
			PathParams:    pp,
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
				isr := recovery(rc, err)
				body, _ := json.Marshal(isr.Entity)
				writer.Header().Add(contentTypeKey, applicationJsonType)
				writer.WriteHeader(isr.StatusCode)
				_, _ = writer.Write(body)
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
			http.Redirect(writer, request, fmt.Sprint(result.Entity), result.StatusCode)
			return
		}
		statusCode := result.StatusCode
		writer.WriteHeader(result.StatusCode)
		responseBody, err := json.Marshal(result.Entity)
		if err != nil {
			statusCode = http.StatusInternalServerError
			// log error here
			isr := msg{Message: "internal server error"}
			isrBytes, _ := json.Marshal(InternalServerError(&isr))
			responseBody = isrBytes
		} else {
			for key, values := range result.Headers {
				for _, value := range values {
					writer.Header().Add(key, value)
				}
			}
		}
		writer.WriteHeader(statusCode)
		if writer.Header().Get(contentTypeKey) == "" {
			writer.Header().Add(contentTypeKey, applicationJsonType)
		}
		_, err = writer.Write(responseBody)
		if err != nil {
			// log failure here
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
	stginLogger.InfoF("%v -> %v | %v | %v", request.Method, request.Url, statusString, difference)
}

type generalFailureMessage struct {
	StatusCode int    `json:"status_code"`
	Path       string `json:"path"`
	Message    string `json:"message"`
	Method     string `json:"method"`
}

var notFoundDefaultAction API = func(request RequestContext) Status {
	return NotFound(&generalFailureMessage{
		StatusCode: 404,
		Path:       request.Url,
		Message:    "route not found",
		Method:     request.Method,
	})
}

var methodNotAllowedDefaultAction API = func(request RequestContext) Status {
	return MethodNotAllowed(&generalFailureMessage{
		StatusCode: http.StatusMethodNotAllowed,
		Path:       request.Url,
		Message:    "method " + request.Method + " not allowed!",
		Method:     request.Method,
	})
}

var errorAction ErrorHandler = func(request RequestContext, err any) Status {
	if parseErr, isParseError := err.(ParseError); isParseError {
		return BadRequest(&generalFailureMessage{
			StatusCode: 400,
			Path:       request.Url,
			Message:    parseErr.Error(),
			Method:     request.Method,
		})
	} // enough for now, user should be able to easily detect and provide actions for both client and server error
	errorRepr := fmt.Sprintf("%v%v%v", colored.RED, fmt.Sprint(err), colored.ResetPrevColor)
	stginLogger.ErrorF("Recovered following error:\n\t%v\n", errorRepr)
	return InternalServerError(&generalFailureMessage{
		StatusCode: 500,
		Path:       request.Url,
		Message:    "internal server error",
		Method:     request.Method,
	})
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
		writer.WriteHeader(status.StatusCode)
		writer.Header().Add(contentTypeKey, applicationJsonType)
		bytes, marshalErr := json.Marshal(status.Entity)
		if marshalErr != nil {
			stginLogger.ErrorF(
				"could not marshal not found action result:\n\t%v%v%v",
				colored.RED, fmt.Sprint(marshalErr), colored.ResetPrevColor,
			)
			bytes, _ = json.Marshal(&generalFailureMessage{
				StatusCode: 404,
				Path:       request.URL.Path,
				Message:    "route not found",
				Method:     request.Method,
			})
		}
		writer.Write(bytes)
	}
}

func (server *Server) handler() http.Handler {
	methodWithRoutes := make(map[string][]Route)
	for _, controller := range server.Controllers {
		for _, route := range controller.routes {
			var log string
			if controller.hasPrefix() {
				log = fmt.Sprintf("Adding %v's API:\t%v -> %v", route.controller.Name, route.Method, route.withPrefixPrepended(controller.prefix))
				methodWithRoutes[route.Method] = append(
					methodWithRoutes[route.Method],
					route.withPrefixPrepended(controller.prefix),
				)
			} else {
				log = fmt.Sprintf("Adding %v's API:\t%v -> %v", route.controller.Name, route.Method, route.Path)
				methodWithRoutes[route.Method] = append(
					methodWithRoutes[route.Method],
					route,
				)
			}

			stginLogger.Info(log)
		}
	}
	return serverHandler{methodWithRoutes: methodWithRoutes, server: server}
}

func (server *Server) Start() error {
	stginLogger.InfoF("starting server on port: %d", server.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", server.port), server.handler())
}

func NewServer(port int) *Server {
	return &Server{
		port:                   port,
		notFoundAction:         notFoundDefaultAction,
		methodNotAllowedAction: methodNotAllowedDefaultAction,
		errorAction:            errorAction,
	}
}
