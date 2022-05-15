package stgin

import (
	"fmt"
	"github.com/AminMal/slogger/colored"
	"net/http"
	"strings"
)

var emptyHeaders http.Header = make(map[string][]string, 5)

type Status struct {
	StatusCode int
	Entity     ResponseEntity
	Headers    http.Header
	cookies    []*http.Cookie
	isDir      bool
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

func write(status Status, rw http.ResponseWriter) {
	bytes, contentType, marshallErr := marshall(status.Entity)
	if marshallErr != nil {
		_ = stginLogger.ErrorF("error while marshalling request entity:\n\t%v", fmt.Sprintf("%s%s%s", colored.RED, marshallErr.Error(), colored.ResetPrevColor))
		panic(marshallErr)
	}
	for key, values := range status.Headers {
		for _, value := range values {
			rw.Header().Set(key, value)
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

func (status Status) complete(request *http.Request, writer http.ResponseWriter) {
	if status.isRedirection() {
		location, _ := status.Entity.Bytes()
		http.Redirect(writer, request, string(location), status.StatusCode)
		return
	} else {
		write(status, writer)
	}
}

func CreateResponse(statusCode int, body ResponseEntity) Status {
	return Status{
		StatusCode: statusCode,
		Entity:     body,
		Headers: emptyHeaders,
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
	}
}

func Found(location string) Status {
	return Status{
		StatusCode: http.StatusFound,
		Entity:     Text(location),
	}
}

func PermanentRedirect(location string) Status {
	return Status{
		StatusCode: http.StatusPermanentRedirect,
		Entity:     Text(location),
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

func Dir(fs string) Status {
	dir := dirPlaceholder{path: fs}
	return Status{
		StatusCode: 200,
		Entity:     dir,
		isDir:      true,
	}
}
