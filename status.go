package stgin

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)


type Status struct {
	StatusCode int
	Entity     any
	Headers    http.Header
}

func (status Status) isRedirection() bool {
	return status.StatusCode >= 300 && status.StatusCode < 400
}


func (status Status) WithHeaders(headers http.Header) Status {
	for key, value := range headers {
		status.Headers[key] = value
	}
	return status
}

var emptyHeaders http.Header = make(map[string][]string, 6)

func CreateResponse(statusCode int, body any) Status {
	return Status{
		StatusCode: statusCode,
		Entity:     body,
		Headers:    emptyHeaders,
	}
}

// 2xx Statuses here

func Ok(body any) Status {
	return CreateResponse(http.StatusOK, body)
}

func Created(body any) Status {
	return CreateResponse(http.StatusCreated, body)
}

// ------------------
// 3xx statuses here

func MovedPermanently(location string) Status {
	return Status{
		StatusCode: http.StatusMovedPermanently,
		Entity:     location,
		Headers:    emptyHeaders,
	}
}

func Found(location string) Status {
	return Status{
		StatusCode: http.StatusFound,
		Entity:     location,
		Headers:    emptyHeaders,
	}
}

func PermanentRedirect(location string) Status {
	return Status{
		StatusCode: http.StatusPermanentRedirect,
		Entity:     location,
		Headers:    emptyHeaders,
	}
}

// ------------------
// 4xx statuses here

func BadRequest(body any) Status {
	return CreateResponse(http.StatusBadRequest, body)
}

func Unauthorized(body any) Status {
	return CreateResponse(http.StatusUnauthorized, body)
}

func Forbidden(body any) Status {
	return CreateResponse(http.StatusForbidden, body)
}

func NotFound(body any) Status {
	return CreateResponse(http.StatusNotFound, body)
}

func MethodNotAllowed(body any) Status {
	return CreateResponse(http.StatusMethodNotAllowed, body)
}

// ------------------
// 5xx statuses here

func InternalServerError(body any) Status {
	return CreateResponse(http.StatusInternalServerError, body)
}

// ------------------



type API = func (c RequestContext) Status

type RequestBody struct {
	underlying io.Reader
	underlyingBytes []byte
	hasFilledBytes  bool
}

func bodyFromBytes(bytes []byte) *RequestBody {
	return &RequestBody{
		underlying:      nil,
		underlyingBytes: bytes,
		hasFilledBytes:  true,
	}
}

func bodyFromReader(reader io.Reader) *RequestBody {
	return &RequestBody{
		underlying:      reader,
		underlyingBytes: nil,
		hasFilledBytes:  false,
	}
}

func bodyFromReadCloser(reader io.ReadCloser) (*RequestBody, error) {
	defer func(r io.ReadCloser) {
		err := r.Close()
		if err != nil {
			_ = stginLogger.Err("could not close reader stream from request")
			return
		}
	}(reader)
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	} else {
		return bodyFromBytes(bytes), nil
	}
}

func (rb *RequestBody) WriteInto(a any) error {
	var bts []byte
	if !rb.hasFilledBytes {
		bytes, err := ioutil.ReadAll(rb.underlying)
		if err != nil {
			return err
		}
		rb.underlyingBytes = bytes
		bts = bytes
		rb.hasFilledBytes = true
	} else {
		bts = rb.underlyingBytes
	}
	return json.Unmarshal(bts, a)
}