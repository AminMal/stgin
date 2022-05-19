package stgin

// RequestListener is a function which is applied to the incoming requests, before the api actually receives it.
// It can be defined both on the controller layer, or server layer.
type RequestListener = func(RequestContext) RequestContext

// ResponseListener is a function which is applied to the outgoing http responses, after they're evaluated by the api.
// It can be defined both on the controller layer, or server layer.
type ResponseListener = func(Status) Status

// APIListener is a function which is applied to both request and response, after the response is written to the client.
// It can be defined both on the controller layer, or server layer.
type APIListener = func(RequestContext, Status)

// ErrorHandler is a function which can decide what to do, based on the request and the error.
type ErrorHandler = func(request RequestContext, err any) Status

// CorsHandler is just a semantic wrapper over common CORS headers.
type CorsHandler struct {
	AllowOrigin      []string
	AllowCredentials []string
	AllowHeaders     []string
	AllowMethods     []string
}
