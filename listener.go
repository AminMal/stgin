package stgin

type RequestListener 		= func(RequestContext) RequestContext
type ResponseListener 		= func(Status) Status
type APIListener			 = func (RequestContext, Status)
