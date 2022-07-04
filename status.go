package stgin

import (
	"fmt"
	"github.com/AminMal/slogger/colored"
	"net/http"
	"strings"
	"time"
)

var emptyHeaders http.Header = make(map[string][]string, 5)

// Status is the return type of stgin APIs.
// It represents an HTTP response with a status code, headers, body and cookies.
type Status struct {
	StatusCode int
	Entity     ResponseEntity
	Headers    http.Header
	cookies    []*http.Cookie
	doneAt     time.Time
}

func (status Status) DoneAt() time.Time { return status.doneAt }

func (status Status) isRedirection() bool {
	return status.StatusCode >= 300 && status.StatusCode < 400
}

// WithCookies returns a new Status, appended the given cookies.
func (status Status) WithCookies(cookies ...*http.Cookie) Status {
	status.cookies = append(status.cookies, cookies...)
	return status
}

// WithHeaders returns a new Status, appended the given headers.
func (status Status) WithHeaders(headers http.Header) Status {
	for key, value := range headers {
		status.Headers[key] = value
	}
	return status
}

func write(status Status, rw http.ResponseWriter) {
	bytes, contentType, marshallErr := marshall(status.Entity)
	if marshallErr != nil {
		_ = stginLogger.ErrorF("error while marshalling request entity:\n\t%v", fmt.Sprintf("%s%s%s", colored.RED, marshallErr.Error(), colored.ResetPrevColor))
		panic(marshallErr)
	}
	for key, values := range status.Headers {
		for _, value := range values {
			rw.Header().Add(key, value)
		}
	}
	for _, cookie := range status.cookies {
		http.SetCookie(rw, cookie)
	}
	rw.Header().Set(contentTypeKey, contentType)
	rw.WriteHeader(status.StatusCode)
	_, err := rw.Write(bytes)
	if err != nil {
		stginLogger.ErrorF("error while writing response to client:\n\t%s", fmt.Sprintf("%s%s%s", colored.RED, err.Error(), colored.ResetPrevColor))
		panic(err)
	}
}

func (status *Status) complete(request *http.Request, writer http.ResponseWriter) {
	if status.isRedirection() {
		location, _ := status.Entity.Bytes()
		http.Redirect(writer, request, string(location), status.StatusCode)
		return
	} else {
		write(*status, writer)
	}
}

// CreateResponse can be used in order to make responses that are not available in default functions in stgin.
// For instance a 202 http response.
func CreateResponse(statusCode int, body ResponseEntity) Status {
	return Status{
		StatusCode: statusCode,
		Entity:     body,
		Headers:    emptyHeaders,
	}
}

// 2xx Statuses here

// Ok represents a basic http 200 response with the given body.
func Ok(body ResponseEntity) Status {
	return CreateResponse(http.StatusOK, body)
}

// Created represents a basic http 201 response with the given body.
func Created(body ResponseEntity) Status {
	return CreateResponse(http.StatusCreated, body)
}

// ------------------
// 3xx statuses here

// MovedPermanently represents a basic http 301 redirect to the given location.
func MovedPermanently(location string) Status {
	return Status{
		StatusCode: http.StatusMovedPermanently,
		Entity:     Text(location),
	}
}

// Found represents a basic http 302 redirect to the given location.
func Found(location string) Status {
	return Status{
		StatusCode: http.StatusFound,
		Entity:     Text(location),
	}
}

// PermanentRedirect represents a basic http 308 redirect to the given location.
func PermanentRedirect(location string) Status {
	return Status{
		StatusCode: http.StatusPermanentRedirect,
		Entity:     Text(location),
	}
}

// ------------------
// 4xx statuses here

// BadRequest represents a basic http 400 response with the given body.
func BadRequest(body ResponseEntity) Status {
	return CreateResponse(http.StatusBadRequest, body)
}

// Unauthorized represents a basic http 401 response with the given body.
func Unauthorized(body ResponseEntity) Status {
	return CreateResponse(http.StatusUnauthorized, body)
}

// Forbidden represents a basic http 403 response with the given body.
func Forbidden(body ResponseEntity) Status {
	return CreateResponse(http.StatusForbidden, body)
}

// NotFound represents a basic http 404 response with the given body.
func NotFound(body ResponseEntity) Status {
	return CreateResponse(http.StatusNotFound, body)
}

// MethodNotAllowed represents a basic http 405 response with the given body.
func MethodNotAllowed(body ResponseEntity) Status {
	return CreateResponse(http.StatusMethodNotAllowed, body)
}

// ------------------
// 5xx statuses here

// InternalServerError represents a basic http 500 response with the given body.
func InternalServerError(body ResponseEntity) Status {
	return CreateResponse(http.StatusInternalServerError, body)
}

//------------------

// File is used to return a file itself as an HTTP response.
// If the file is not found, it returns 404 not found to the client.
// If there are issues reading file or anything else related, 500 internal server error is returned to the client.
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
