package stgin

import (
	"github.com/AminMal/slogger/colored"
	"net/http"
	"strings"
)

type Status struct {
	StatusCode int
	Entity     ResponseEntity
	Headers    http.Header
	cookies    []*http.Cookie
}

func (status Status) isRedirection() bool {
	return status.StatusCode >= 300 && status.StatusCode < 400
}

func (status Status) WithCookies(cookies ...*http.Cookie) Status {
	status.cookies = append(status.cookies, cookies...)
	return status
}

func (status Status) WithHeaders(headers http.Header) Status {
	for key, value := range headers {
		status.Headers[key] = value
	}
	return status
}

var emptyHeaders http.Header = make(map[string][]string, 6)

func CreateResponse(statusCode int, body ResponseEntity) Status {
	return Status{
		StatusCode: statusCode,
		Entity:     body,
		Headers:    emptyHeaders,
	}
}

// 2xx Statuses here

func Ok(body ResponseEntity) Status {
	return CreateResponse(http.StatusOK, body)
}

func Created(body ResponseEntity) Status {
	return CreateResponse(http.StatusCreated, body)
}

// ------------------
// 3xx statuses here

func MovedPermanently(location string) Status {
	return Status{
		StatusCode: http.StatusMovedPermanently,
		Entity:     Text(location),
		Headers:    emptyHeaders,
	}
}

func Found(location string) Status {
	return Status{
		StatusCode: http.StatusFound,
		Entity:     Text(location),
		Headers:    emptyHeaders,
	}
}

func PermanentRedirect(location string) Status {
	return Status{
		StatusCode: http.StatusPermanentRedirect,
		Entity:     Text(location),
		Headers:    emptyHeaders,
	}
}

// ------------------
// 4xx statuses here

func BadRequest(body ResponseEntity) Status {
	return CreateResponse(http.StatusBadRequest, body)
}

func Unauthorized(body ResponseEntity) Status {
	return CreateResponse(http.StatusUnauthorized, body)
}

func Forbidden(body ResponseEntity) Status {
	return CreateResponse(http.StatusForbidden, body)
}

func NotFound(body ResponseEntity) Status {
	return CreateResponse(http.StatusNotFound, body)
}

func MethodNotAllowed(body ResponseEntity) Status {
	return CreateResponse(http.StatusMethodNotAllowed, body)
}

// ------------------
// 5xx statuses here

func InternalServerError(body ResponseEntity) Status {
	return CreateResponse(http.StatusInternalServerError, body)
}

//------------------

func File(path string) Status {
	file := fileContent{path}
	_, err := file.Bytes()
	if err != nil {
		_ = stginLogger.Colored(colored.RED).ErrorF("error reading file '%s': %s", file.path, err.Error())
		if strings.Contains(err.Error(), "no such file or directory") {
			return NotFound(Text("404 not found"))
		} else {
			return InternalServerError(Text("internal server error"))
		}
	} else {
		return Ok(file)
	}
}

// todo, add support for directories
