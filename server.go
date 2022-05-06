package stgin

import (
	"encoding/json"
	"fmt"
	"github.com/AminMal/slogger/colored"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// This file is the part that underlying library changes are applied

type Server struct {
	port 					int
	Controllers				[]*Controller
	requestListeners 		[]RequestListener
	responseListeners 		[]ResponseListener
	journeyListeners    	[]ApiJourneyListener
	notFoundAction      	API
	methodNotAllowedAction 	API
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

func (server *Server) AddJourneyListeners(listeners ...ApiJourneyListener) {
	server.journeyListeners = append(server.journeyListeners, listeners...)
}

func (server *Server) NotFoundAction(action API) {
	server.notFoundAction = action
}

func (server *Server) MethodNowAllowedAction(action API) {
	server.methodNotAllowedAction = action
}

type msg struct {
	Message string `json:"message"`
}

// this is where the integration happens, having all the information from request context of underlying framework
// and API with all the other things, now combine the smallest unit of stgin (API) with gin (HandlerFunc)
func createHandlerFuncFromApi(
	api API,
	requestListeners []RequestListener,
	responseListeners []ResponseListener,
	journeyListeners []ApiJourneyListener,
	) gin.HandlerFunc {
	return func(context *gin.Context) {
		queryParams := make(map[string][]string, 10)
		for key, value := range context.Request.URL.Query() {
			queryParams[key] = value
		}

		pathParams := make(map[string]string, 10)

		for _, param := range context.Params {
			pathParams[param.Key] = param.Value
		}

		url := context.FullPath()
		headers := context.Request.Header
		body, err := bodyFromReadCloser(context.Request.Body)
		rc := RequestContext{
			Url:         url,
			QueryParams: queryParams,
			PathParams:  pathParams,
			Headers:     headers,
			Body:        body,
			receivedAt:  time.Now(),
			Method:      context.Request.Method,
		}
		for _, requestListener := range requestListeners {
			rc = requestListener(rc)
		}

		result := api(rc)
		for _, responseListener := range responseListeners {
			result = responseListener(result)
		}

		for _, journeyListener := range journeyListeners {
			journeyListener(rc, result)
		}

		if result.isRedirection() {
			http.Redirect(context.Writer, context.Request, fmt.Sprint(result.Entity), result.StatusCode)
			return
		}
		statusCode := result.StatusCode
		context.Status(result.StatusCode)
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
					context.Writer.Header().Add(key, value)
				}
			}
		}
		context.Status(statusCode)
		if context.Writer.Header().Get(contentTypeKey) == "" {
			context.Writer.Header().Add(contentTypeKey, applicationJsonType)
		}
		_, err = context.Writer.Write(responseBody)
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

func WatchAPIs() ApiJourneyListener {
	return func(request RequestContext, status Status) {
		now := time.Now()
		difference := fmt.Sprint(now.Sub(request.receivedAt))
		statusString := fmt.Sprintf("%v%d%v", getColor(status.StatusCode), status.StatusCode, colored.ResetPrevColor)
		stginLogger.InfoF("%v -> %v | %v | %v", request.Method, request.Url, statusString, difference)
	}
}
// todo, implement recovery func

type generalFailureMessage struct {
	StatusCode 		int 	`json:"status_code"`
	Path 			string 	`json:"path"`
	Message 		string  `json:"message"`
	Method 			string 	`json:"method"`
}

var notFoundDefaultAction API = func(request RequestContext) Status {
	return NotFound(&generalFailureMessage{
		StatusCode: 404,
		Path:       request.Url,
		Message:    "route not found",
		Method:		request.Method,
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

func (server *Server) Start() error {
	engine := gin.New()
	engine.Use(gin.Recovery())
	controllers := server.Controllers
	for _, controller := range controllers {
		for _, route := range controller.routes {
			requestListeners := append(server.requestListeners, controller.requestListeners...)
			responseListeners := append(server.responseListeners, controller.responseListeners...)
			journeyListeners := append(server.journeyListeners, controller.journeyListeners...)
			handlerFunc := createHandlerFuncFromApi(
				route.Action,
				requestListeners,
				responseListeners,
				journeyListeners,
				)
			var fullPath string
			if controller.hasPrefix() {
				fullPath = fmt.Sprintf("%v%v", controller.prefix, route.Path)
			} else {
				fullPath = route.Path
			}
			engine.Handle(route.Method, fullPath, handlerFunc)
		}
	}
	engine.NoRoute(createHandlerFuncFromApi(
		server.notFoundAction,
		server.requestListeners,
		server.responseListeners,
		server.journeyListeners,
		))
	engine.NoMethod(createHandlerFuncFromApi(
		server.methodNotAllowedAction,
		server.requestListeners,
		server.responseListeners,
		server.journeyListeners,
		))
	return engine.Run(fmt.Sprintf(":%d", server.port))
}

func NewServer(port int) *Server {
	return &Server{
		port: port,
		notFoundAction: notFoundDefaultAction,
		methodNotAllowedAction: methodNotAllowedDefaultAction,
	}
}

