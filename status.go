package stgin

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

type Status interface {
	Status() int
	Entity() any
	Headers() http.Header
	WithHeaders(headers http.Header) Status
}

// todo, add more statuses

type Ok struct {
	Body any
	headers http.Header
}

func (status Ok) WithHeaders(headers http.Header) Status {
	for key, value := range headers {
		status.headers[key] = value
	}
	return status
}

func (status Ok) Status() int {
	return OK
}

func (status Ok) Entity()  any {
	return status.Body
}

func (status Ok) Headers() http.Header {
	return status.headers
}

type Created struct {
	entity 	 any
	headers  http.Header
}

func (status Created) Status() int {
	return CREATED
}

func (status Created) Entity() any {
	return status.entity
}

func (status Created) Headers() http.Header {
	return status.headers
}

func (status Created) WithHeaders(headers http.Header) Status {
	for key, value := range headers {
		status.headers[key] = value
	}
	return status
}

type BadRequest struct {
	entity 	  any
	headers   http.Header
}

func (status BadRequest) Status() int {
	return BAD_REQUEST
}

func (status BadRequest) Entity() any {
	return status.entity
}

func (status BadRequest) Headers() http.Header {
	return status.headers
}

func (status BadRequest) WithHeaders(headers http.Header) Status {
	for key, value := range headers {
		status.headers[key] = value
	}
	return status
}

type Unauthorized struct {
	entity     any
	headers    http.Header
}

func (status Unauthorized) Status() int {
	return UNAUTHORIZED
}

func (status Unauthorized) Entity() any {
	return status.entity
}

func (status Unauthorized) Headers() http.Header {
	return status.headers
}

func (status Unauthorized) WithHeaders(headers http.Header) Status {
	for key, value := range headers {
		status.headers[key] = value
	}
	return status
}

type Forbidden struct {
	entity 		any
	headers 	http.Header
}

func (status Forbidden) Status() int {
	return FORBIDDEN
}

func (status Forbidden) Entity() any {
	return status.entity
}

func (status Forbidden) Headers() http.Header {
	return status.headers
}

func (status Forbidden) WithHeaders(headers http.Header) Status {
	for key, value := range headers {
		status.headers[key] = value
	}
	return status
}

type InternalServerError struct {
	Body    any
	headers http.Header
}

func (status InternalServerError) Entity() any {
	return status.Body
}

func (status InternalServerError) Status() int {
	return INTERNAL_SERVER_ERROR
}

func (status InternalServerError) Headers() http.Header {
	return status.headers
}

func (status InternalServerError) WithHeaders(headers http.Header) Status {
	for key, value := range headers {
		status.headers[key] = value
	}
	return status
}

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
			_ = stginLogger.Err("Could not close reader stream from request")
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

func (rb *RequestBody) ReadInto(a any) error {
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