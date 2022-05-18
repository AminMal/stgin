package stgin

import (
	"testing"
)

func shouldHavePanicked(t *testing.T) {
	if err := recover(); err == nil {
		t.Error("make route should've panicked for nil api action")
	}
}

func TestMkRoute(t *testing.T) {
	defer shouldHavePanicked(t)
	GET("/test", nil)
}

func TestEmptyRouteCreationStage(t *testing.T) {
	defer shouldHavePanicked(t)
	OnPath("/test/$username").WithMethod("GET").Do(nil)
}
