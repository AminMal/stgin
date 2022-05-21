package stgin

import (
	"net/http"
	"time"
)

// Interrupt is n abstract semantic, that can be executed along-side the request,
// And if some event happens, can abort the request and complete with another response,
// which is filled within completeWith.
// A real world example would be timeouts; server.SetTimeout actually uses an interrupt to perform timeout operations.
type Interrupt interface {
	TriggerFor(request RequestContext, completeWith chan *Status)
}

func contextTimeoutExceededResponse() Status {
	return Status{
		StatusCode: http.StatusRequestTimeout,
		Entity:     Text("408 - request timed out"),
		doneAt:     time.Now(),
	}
}

type timeoutInterrupt struct {
	timeout time.Duration
}

func (t timeoutInterrupt) TriggerFor(_ RequestContext, completeWith chan *Status) {
	<-time.After(t.timeout)
	result := contextTimeoutExceededResponse()
	completeWith <- &result
}

func TimeoutInterrupt(timeout time.Duration) Interrupt {
	return timeoutInterrupt{timeout: timeout}
}
