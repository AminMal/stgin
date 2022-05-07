package stgin

type RequestListener 		= func(RequestContext) RequestContext
type ResponseListener 		= func(Status) Status
type APIListener			= func (RequestContext, Status)
type ErrorHandler			= func(request RequestContext, err any) Status
