package stgin

import (
	"fmt"
	"strings"
)

type Controller struct {
	Name    string
	routes 	[]Route
	prefix  string
}

func NewController(name string, routes ...Route) *Controller {
	return &Controller{
		Name:   name,
		routes: routes,
	}
}

func (controller *Controller) SetRoutePrefix(prefix string) {
	if strings.HasPrefix(prefix, "/") {
		controller.prefix = prefix
	} else {
		controller.prefix = fmt.Sprintf("%v%v", "/", prefix)
	}
}

func (controller *Controller) hasPrefix() bool {
	return controller.prefix == ""
}