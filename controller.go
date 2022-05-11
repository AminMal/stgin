package stgin

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type Controller struct {
	Name              string
	routes            []Route
	prefix            string
	requestListeners  []RequestListener
	responseListeners []ResponseListener
	apiListeners      []APIListener
}

func NewController(name string) *Controller {
	return &Controller{
		Name: name,
	}
}

func (controller *Controller) SetRoutePrefix(prefix string) {
	if strings.HasPrefix(prefix, "/") {
		controller.prefix = prefix
	} else {
		controller.prefix = fmt.Sprintf("%v%v", "/", prefix)
	}
}

func (controller *Controller) AddRoutes(routes ...Route) {
	for _, route := range routes {
		route.controller = controller
		controller.routes = append(controller.routes, route)
	}
}

func (controller *Controller) AddRequestListeners(listeners ...RequestListener) {
	controller.requestListeners = append(controller.requestListeners, listeners...)
}

func (controller *Controller) AddResponseListener(listeners ...ResponseListener) {
	controller.responseListeners = append(controller.responseListeners, listeners...)
}

func (controller *Controller) AddAPIListeners(listeners ...APIListener) {
	controller.apiListeners = append(controller.apiListeners, listeners...)
}

func (controller *Controller) hasPrefix() bool {
	return controller.prefix != ""
}

func (controller *Controller) executeInternal(request *http.Request) Status {
	body := RequestBody{
		underlying:      nil,
		underlyingBytes: []byte{},
		hasFilledBytes:  false,
	}

	rc := RequestContext{
		Url:           request.URL.Path,
		QueryParams:   request.URL.Query(),
		PathParams:    nil,
		Headers:       request.Header,
		Body:          &body,
		receivedAt:    time.Now(),
		Method:        request.Method,
		ContentLength: request.ContentLength,
		Host:          request.Host,
		MultipartForm: func() *multipart.Form {
			return request.MultipartForm
		},
		Scheme:        request.URL.Scheme,
		RemoteAddr:    request.RemoteAddr,
	}

	for _, modifier := range controller.requestListeners {
		rc = modifier(rc)
	}

	var done bool
	var result Status
	for _, route := range controller.routes {
		var r Route
		if controller.hasPrefix() {
			r = route.withPrefixPrepended(controller.prefix)
		} else {
			r = route
		}
		matches, pathParams := r.acceptsAndPathParams(request)
		if !matches {
			continue
		} else {
			rc.PathParams = pathParams
			done = true
			result = route.Action(rc)
			break
		}
	}
	if !done {
		result = NotFound(&msg{Message: "not found"})
	}

	for _, modifier := range controller.responseListeners {
		result = modifier(result)
	}

	for _, watcher := range controller.apiListeners {
		watcher(rc, result)
	}

	return result
}