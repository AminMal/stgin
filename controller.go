package stgin

import (
	"mime/multipart"
	"net/http"
	"time"
)

// Controller is a struct which groups some routes, and may have a path prefix.
// And might hold some request/response/api listeners.
type Controller struct {
	Name              string
	routes            []Route
	prefix            string
	requestListeners  []RequestListener
	responseListeners []ResponseListener
	apiListeners      []APIListener
	interrupts        []Interrupt
}

// NewController returns a pointer to a newly created controller with the given name and path prefixes.
func NewController(name string, prefix string) *Controller {
	return &Controller{
		Name:   name,
		prefix: "/" + prefix + "/",
	}
}

// AddRoutes normalizes, and evaluates path patterns for the given routes, and then adds them to the routes it contains.
func (controller *Controller) AddRoutes(routes ...Route) {
	for _, route := range routes {
		route.controller = controller
		route.Path = normalizePath(controller.prefix + route.Path)
		route.correspondingRegex = getRoutePatternRegexOrPanic(route.Path)
		controller.routes = append(controller.routes, route)
	}
}

// AddRequestListeners registers the given listeners to the controller.
// These listeners then will be applied to all the requests coming inside this controller.
func (controller *Controller) AddRequestListeners(listeners ...RequestListener) {
	controller.requestListeners = append(controller.requestListeners, listeners...)
}

// AddResponseListener registers the given listeners to the controller.
// These listeners then will be applied to all the outgoing responses from this controller.
func (controller *Controller) AddResponseListener(listeners ...ResponseListener) {
	controller.responseListeners = append(controller.responseListeners, listeners...)
}

// AddAPIListeners registers the given listeners to the controller.
// These listeners then will be applied to all the incoming/outgoing requests and responses after they're evaluated
// And returned to the client.
func (controller *Controller) AddAPIListeners(listeners ...APIListener) {
	controller.apiListeners = append(controller.apiListeners, listeners...)
}

// SetTimeout registers a timeout interrupt into the controller.
func (controller *Controller) SetTimeout(timeout time.Duration) {
	controller.RegisterInterrupts(TimeoutInterrupt(timeout))
}

// RegisterInterrupts adds the given interrupts to the controller's already existing interrupts.
func (controller *Controller) RegisterInterrupts(interrupts ...Interrupt) {
	controller.interrupts = append(controller.interrupts, interrupts...)
}

// executeInternal is just for testing purposes. This simulates executing an actual http request.
func (controller *Controller) executeInternal(request *http.Request) Status {
	var headers http.Header
	if request.Header == nil {
		headers = emptyHeaders
	} else {
		headers = request.Header
	}

	rc := RequestContext{
		Url:         request.URL.Path,
		QueryParams: Queries{request.URL.Query()},
		PathParams:  PathParams{nil},
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

	interruptChannel := make(chan *Status, 1)
	successfulOperationChannel := make(chan *Status, 1)
	// panicChannel should not be defined here, since error handling is on server layer

	for _, modifier := range controller.requestListeners {
		rc = modifier(rc)
	}

	for _, interrupt := range controller.interrupts {
		go interrupt.TriggerFor(rc, interruptChannel)
	}

	var done bool
	go func() {
		var result Status
		for _, route := range controller.routes {
			matches, pathParams := route.acceptsAndPathParams(request)
			if matches && acceptsAllQueries(route.expectedQueries, request.URL.Query()) {
				rc.PathParams = PathParams{pathParams}
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
		successfulOperationChannel <- &result
	}()

	select {
	case interrupted := <-interruptChannel:
		result := *interrupted
		for _, watcher := range controller.apiListeners {
			go watcher(rc, result)
		}
		return result
	case normal := <-successfulOperationChannel:
		result := *normal
		for _, watcher := range controller.apiListeners {
			go watcher(rc, result)
		}
		return result
	}
}
