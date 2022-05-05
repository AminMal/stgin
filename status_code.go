package stgin

type StatusCode = int

// todo, add more status codes
const (
	OK StatusCode = 200
	CREATED StatusCode = 201
	BAD_REQUEST StatusCode = 400
	UNAUTHORIZED StatusCode = 401
	FORBIDDEN StatusCode = 403
	INTERNAL_SERVER_ERROR StatusCode = 500
)
