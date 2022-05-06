package stgin

type RequestListener 		= func(RequestContext) RequestContext
type ResponseListener 		= func(Status) Status
type ApiJourneyListener 	= func (RequestContext, Status)
