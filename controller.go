package stgin

import (
	"fmt"
	"strings"
)

type ControllerRequestListener    = func(r RequestContext) RequestContext
type ControllerResponseListener   = func(response Status) Status
type ControllerApiJourneyListener = func(RequestContext, Status)

type Controller struct {
	Name             	string
	routes           	[]Route
	prefix           	string
	requestListeners 	[]ControllerRequestListener
	responseListeners 	[]ControllerResponseListener
	journeyListeners    []ControllerApiJourneyListener
}

func NewController(name string) *Controller {
	return &Controller{
		Name:   name,
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
	controller.routes = append(controller.routes, routes...)
}

func (controller *Controller) AddRequestListeners(listeners ...ControllerRequestListener) {
	controller.requestListeners = append(controller.requestListeners, listeners...)
}

func (controller *Controller) AddResponseListener(listeners ...ControllerResponseListener) {
	controller.responseListeners = append(controller.responseListeners, listeners...)
}

func (controller *Controller) AddJourneyListeners(listeners ...ControllerApiJourneyListener) {
	controller.journeyListeners = append(controller.journeyListeners, listeners...)
}

func (controller *Controller) hasPrefix() bool {
	return controller.prefix != ""
}