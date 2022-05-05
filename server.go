package stgin

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// This file is the part that underlying library changes are applied

type Server struct {
	port 			int
	Controllers		[]*Controller
}

func (server *Server) Register(controller *Controller) {
	var found bool
	for _, c := range server.Controllers {
		if c == controller {
			found = true
			break
		}
	}
	if !found {
		server.Controllers = append(server.Controllers, controller)
	}
}

func (server *Server) RegisterAll(controllers ...*Controller) int {
	nonExistingControllers := make([]*Controller, len(controllers))
	for _, controller := range controllers {
		for _, existingController := range server.Controllers {
			if controller == existingController {
				continue
			}
			nonExistingControllers = append(nonExistingControllers, controller)
		}
	}
	server.Controllers = append(server.Controllers, nonExistingControllers...)
	return len(nonExistingControllers)
}

type msg struct {
	Message string `json:"message"`
}

// this is where the integration happens, having all the information from request context of underlying framework
// and API with all the other things, now combine the smallest unit of stgin (API) with gin (HandlerFunc)
func createHandlerFuncFromApi(api API) gin.HandlerFunc {
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
		}
		result := api(rc)
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
		}
		context.Status(statusCode)
		context.Writer.Header().Add(contentTypeKey, applicationJsonType)
		_, err = context.Writer.Write(responseBody)
		if err != nil {
			// log failure here
		}
	}
}

func (server *Server) Start() error {
	engine := gin.Default()
	controllers := server.Controllers
	for _, controller := range controllers {
		for _, api := range controller.routes {
			handlerFunc := createHandlerFuncFromApi(api.Action)
			var fullPath string
			if controller.hasPrefix() {
				fullPath = fmt.Sprintf("%v%v", controller.prefix, api.Path)
			} else {
				fullPath = api.Path
			}
			engine.Handle(api.Method, fullPath, handlerFunc)
		}
	}

	return engine.Run(fmt.Sprintf(":%d", server.port))
}

func NewServer(port int) *Server {
	return &Server{port: port}
}

