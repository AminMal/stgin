package stgin

// RequestModifier is a function that accepts a *RequestChangeable object and allows only some fields to be mutated
type RequestModifier = func(*RequestChangeable)

// ResponseModifier is a function that accepts a Response object and allows only some fields to be mutated
type ResponseModifier = func(Response)

// ApiWatcher is a function that gets executed after an http response is done, can be used for logging, etc.
type ApiWatcher = func(*RequestContext, Response)

// ErrorHandler is a function which can decide what to do, based on the request and the error.
type ErrorHandler = func(request *RequestContext, err any) Response

// CorsHandler is just a semantic wrapper over common CORS headers.
type CorsHandler struct {
	AllowOrigin      []string
	AllowCredentials []string
	AllowHeaders     []string
	AllowMethods     []string
}
