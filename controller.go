package stgin

import (
	"fmt"
	"strings"
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
