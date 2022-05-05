package stgin

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"fmt"
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
		queryParams := make(map[string]string, 10)
		for _, param := range context.Params {
			queryParams[param.Key] = param.Value
		}

		url := context.FullPath()
		headers := context.Request.Header
		body, err := bodyFromReadCloser(context.Request.Body)
		rc := RequestContext{
			Url:     url,
			Params:  queryParams,
			Headers: headers,
			Body:    body,
		}
		result := api(rc)
		context.Status(result.Status())
		responseBody, err := json.Marshal(result.Entity())
		if err != nil {
			context.Status(INTERNAL_SERVER_ERROR)
			// log error here
			isr := msg{Message: "internal server error"}
			isrBytes, _ := json.Marshal(InternalServerError{Body: &isr})
			responseBody = isrBytes
		} else {
			context.Status(result.Status())
		}
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
			}
			engine.Handle(api.Method, fullPath, handlerFunc)
		}
	}

	return engine.Run()
}

func NewServer(port int) *Server {
	return &Server{port: port}
}

