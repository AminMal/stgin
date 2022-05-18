package stgin

import (
	"mime/multipart"
	"net/http"
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

func NewController(name string, prefix string) *Controller {
	return &Controller{
		Name:   name,
		prefix: normalizePath("/" + prefix),
	}
}

func (controller *Controller) AddRoutes(routes ...Route) {
	for _, route := range routes {
		path := normalizePath(controller.prefix + route.Path)
		route.controller = controller
		route.correspondingRegex = getRoutePatternRegexOrPanic(path)
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

func (controller *Controller) executeInternal(request *http.Request) Status {
	var headers http.Header
	if request.Header == nil {
		headers = emptyHeaders
	} else {
		headers = request.Header
	}

	rc := RequestContext{
		Url:         request.URL.Path,
		QueryParams: request.URL.Query(),
		PathParams:  nil,
		Headers:     headers,
		Body: func() *RequestBody {
			return &RequestBody{
				underlying:      nil,
				underlyingBytes: []byte{},
				hasFilledBytes:  false,
			}
		},
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

	for _, modifier := range controller.requestListeners {
		rc = modifier(rc)
	}

	var done bool
	var result Status
	for _, route := range controller.routes {
		matches, pathParams := route.withPrefixPrepended(controller.prefix).acceptsAndPathParams(request)
		if matches && acceptsAllQueries(route.expectedQueries, request.URL.Query()) {
			rc.PathParams = pathParams
			done = true
			result = route.Action(rc)
			break
		}
	}
	if !done {
		result = NotFound(Json(&generalFailureMessage{
			StatusCode: 404,
			Path:       request.URL.Path,
			Message:    "not found",
			Method:     request.Method,
		}))
	}

	for _, modifier := range controller.responseListeners {
		result = modifier(result)
	}

	for _, watcher := range controller.apiListeners {
		watcher(rc, result)
	}

	return result
}
