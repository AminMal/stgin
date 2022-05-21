package stgin

import (
	"encoding/json"
	"fmt"
	"github.com/AminMal/slogger/colored"
	"net/http"
	"time"
)

var defaultController *Controller = NewController("Server", "")

// Server is the starting point of stgin applications, which holds the address, controllers, APIs and server-level listeners.
// Which can be run on the specified address.
type Server struct {
	addr              string
	Controllers       []*Controller
	requestListeners  []RequestListener
	responseListeners []ResponseListener
	apiListeners      []APIListener
	notFoundAction    API
	errorAction       ErrorHandler
	interrupts        []Interrupt
}

// Register appends given controllers to the server.
func (server *Server) Register(controllers ...*Controller) {
	server.Controllers = append(server.Controllers, controllers...)
}

// AddRoutes is an alternative to controller.AddRoutes, which adds the given routes to the server's default controller.
func (server *Server) AddRoutes(routes ...Route) {
	for _, c := range server.Controllers {
		if c == defaultController {
			c.AddRoutes(routes...)
			break
		}
	}
}

// CorsHandler function takes the responsibility to handle requests with "OPTIONS" method with the given headers in handler parameter.
func (server *Server) CorsHandler(handler CorsHandler) {
	server.AddRoutes(OPTIONS(Prefix(""), func(RequestContext) Status {
		return Ok(Empty()).WithHeaders(http.Header{
			"Access-Control-Allow-Origin":      handler.AllowOrigin,
			"Access-Control-Allow-Credentials": handler.AllowCredentials,
			"Access-Control-Allow-Headers":     handler.AllowHeaders,
			"Access-Control-Allow-Methods":     handler.AllowMethods,
		})
	}))
}

// AddRequestListeners adds the given request listeners to server-level
// listeners (which then will be applied to all the incoming requests).
func (server *Server) AddRequestListeners(listeners ...RequestListener) {
	server.requestListeners = append(server.requestListeners, listeners...)
}

// AddResponseListeners adds the given response listeners to server-level
// listeners (which then will be applied to all the outgoing responses).
func (server *Server) AddResponseListeners(listeners ...ResponseListener) {
	server.responseListeners = append(server.responseListeners, listeners...)
}

// AddAPIListeners adds the given api listeners to server-level
// listeners (which then will be applied to all the incoming requests and outgoing responses after they're finished).
func (server *Server) AddAPIListeners(listeners ...APIListener) {
	server.apiListeners = append(server.apiListeners, listeners...)
}

// NotFoundAction defines what server should do with the requests that match no routes.
func (server *Server) NotFoundAction(action API) {
	server.notFoundAction = action
}

// SetErrorHandler defines what server should do in case some api panics.
func (server *Server) SetErrorHandler(action ErrorHandler) {
	server.errorAction = action
}

// SetTimeout registers a timeout interrupt to the server
func (server *Server) SetTimeout(dur time.Duration) {
	server.RegisterInterrupts(TimeoutInterrupt(dur))
}

// RegisterInterrupts adds the given interrupts to the server's already existing interrupts
func (server *Server) RegisterInterrupts(interrupts ...Interrupt) {
	server.interrupts = append(server.interrupts, interrupts...)
}

func catchErrInto(errChan chan interface{}) {
	if err := recover(); err != nil {
		errChan <- err
	}
}

func executeInterrupts(interrupts []Interrupt, request RequestContext, completeWith chan *Status) {
	for _, interrupt := range interrupts {
		go interrupt.TriggerFor(request, completeWith)
	}
}

// translate is a function which takes stgin specifications about user defined APIs,
// and is responsible to translate it into the lower-level base package(currently net/http).
func translate(
	api API,
	requestListeners []RequestListener,
	responseListeners []ResponseListener,
	apiListeners []APIListener,
	recovery ErrorHandler,
	pathParams Params,
	interrupts []Interrupt,
) http.HandlerFunc {
	panicChannel := make(chan interface{}, 1)
	successfulResultChannel := make(chan *Status, 1)
	interruptChannel := make(chan *Status, 1)

	return func(writer http.ResponseWriter, request *http.Request) {
		queries := make(map[string][]string, 10)
		for key, value := range request.URL.Query() {
			queries[key] = value
		}

		rc := requestContextFromHttpRequest(request, writer, pathParams)

		for _, requestListener := range requestListeners {
			rc = requestListener(rc)
		}
		go executeInterrupts(interrupts, rc, interruptChannel)

		go func() {
			defer catchErrInto(panicChannel)

			result := api(rc)
			for _, responseListener := range responseListeners {
				result = responseListener(result)
			}
			result.doneAt = time.Now()
			successfulResultChannel <- &result
		}()

		select {
		case interrupt := <-interruptChannel:
			result := *interrupt
			interrupt.complete(request, writer)
			for _, apiListener := range apiListeners {
				apiListener(rc, result)
			}
		case success := <-successfulResultChannel:
			success.complete(request, writer)

			for _, apiListener := range apiListeners {
				go apiListener(rc, *success)
			}

		case err := <-panicChannel:
			if recovery == nil {
				panic(err)
			} else {
				status := recovery(rc, err)
				write(status, writer)
			}
		}
	}
}

// WatchAPIs is the default request and response logger for stgin.
// It logs the input request and the output response into the console.
func WatchAPIs(request RequestContext, status Status) {
	difference := fmt.Sprint(status.doneAt.Sub(request.receivedAt))
	if status.StatusCode == http.StatusRequestTimeout {
		_ = stginLogger.InfoF("%s -> %s\t\t|%s request timed out after %s%s",
			request.Method, request.Url, colored.YELLOW, difference, colored.ResetPrevColor,
		)
		return
	}
	statusString := fmt.Sprintf("%v%d%v", getColor(status.StatusCode), status.StatusCode, colored.ResetPrevColor)
	_ = stginLogger.InfoF("%s -> %s\t\t| %v | %v", request.Method, request.Url, statusString, difference)
}

type generalFailureMessage struct {
	StatusCode int    `json:"status_code"`
	Path       string `json:"path"`
	Message    string `json:"message"`
	Method     string `json:"method"`
}

var notFoundDefaultAction API = func(request RequestContext) Status {
	return NotFound(Json(&generalFailureMessage{
		StatusCode: http.StatusNotFound,
		Path:       request.Url,
		Message:    "route not found",
		Method:     request.Method,
	}))
}

var errorAction ErrorHandler = func(request RequestContext, err any) Status {
	printStacktrace(fmt.Sprintf("recovering following error: %v%v%v", colored.RED, fmt.Sprint(err), colored.ResetPrevColor))
	if parseErr, isParseError := err.(ParseError); isParseError {
		return BadRequest(Json(&generalFailureMessage{
			StatusCode: 400,
			Path:       request.Url,
			Message:    parseErr.Error(),
			Method:     request.Method,
		}))
	}

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
	for _, route := range handler.methodWithRoutes[request.Method] {
		accepts, pathParams := route.acceptsAndPathParams(request)
		if accepts && acceptsAllQueries(route.expectedQueries, request.URL.Query()) {
			requestListeners := append(handler.server.requestListeners, route.controller.requestListeners...)
			responseListeners := append(handler.server.responseListeners, route.controller.responseListeners...)
			apiListeners := append(handler.server.apiListeners, route.controller.apiListeners...)
			interrupts := append(handler.server.interrupts, route.controller.interrupts...)
			handlerFunc := translate(
				route.Action,
				requestListeners,
				responseListeners,
				apiListeners,
				handler.server.errorAction,
				pathParams,
				interrupts,
			)
			handlerFunc(writer, request)
			done = true
			break
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
				methodWithRoutes[route.Method] = append(methodWithRoutes[route.Method], route)
				log = routeAppendLog(controller.Name, route.Method, route.Path)
			} else {
				dir := route.dir
				log = bindStaticDirLog(route.Path, dir)
				mux.Handle(route.Path, http.StripPrefix(route.Path, http.FileServer(http.Dir(dir))))
			}
			_ = stginLogger.Info(log)
		}
	}
	mux.Handle("/", apiHandler{methodWithRoutes: methodWithRoutes, server: server})
	return mux
}

// Start executes the server over the specified address.
// In case any uncaught error or panic happens, and is not recovered in the server's error handler,
// the error value is returned as a result.
func (server *Server) Start() error {
	_ = stginLogger.InfoF("started server over address: %s%s%s", colored.YELLOW, server.addr, colored.ResetPrevColor)
	return http.ListenAndServe(server.addr, server.handler())
}

// NewServer returns a pointer to a basic stgin Server.
func NewServer(addr string) *Server {
	return &Server{
		addr:           addr,
		notFoundAction: notFoundDefaultAction,
		errorAction:    nil,
		Controllers:    []*Controller{defaultController},
	}
}

// DefaultServer is the recommended approach to get a new Server.
// It includes error handler and api logger by default.
func DefaultServer(addr string) *Server {
	server := NewServer(addr)
	server.AddAPIListeners(WatchAPIs)
	server.errorAction = errorAction
	return server
}
