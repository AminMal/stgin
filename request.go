package stgin

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// RequestContext holds all the information about incoming http requests.
type RequestContext struct {
	Url           string
	QueryParams   Queries
	PathParams    PathParams
	Headers       http.Header
	Trailer       http.Header
	Body          func() *RequestBody
	receivedAt    time.Time
	Method        string
	ContentLength int64
	Host          string
	MultipartForm func() *multipart.Form
	Scheme        string
	RemoteAddr    string
	underlying    *http.Request
	HttpPush      Push
}

func (request RequestContext) ReceivedAt() time.Time {
	return request.receivedAt
}

// Push is a struct that represents both the ability, and the functionality of http push inside this request.
type Push struct {
	IsSupported bool
	pusher      http.Pusher
}

// Pusher returns the actual http.Pusher instance, only if it's supported.
// Note that this will panic if it's not supported, so make sure to use IsSupported field before calling this function.
func (p Push) Pusher() http.Pusher {
	if !p.IsSupported {
		panic("pusher is not supported in the request")
	}
	return p.pusher
}

// Cookies returns the cookies that are attached to the request.
func (request RequestContext) Cookies() []*http.Cookie {
	return request.underlying.Cookies()
}

// Referer returns the value of referer header in http request, returns empty string if it does not exist.
func (request RequestContext) Referer() string { return request.underlying.Referer() }

// UserAgent returns the value of request's user agent, returns empty string if it does not exist.
func (request RequestContext) UserAgent() string { return request.underlying.UserAgent() }

// Cookie tries to find a cookie with the given name.
func (request RequestContext) Cookie(name string) (*http.Cookie, error) {
	return request.underlying.Cookie(name)
}

// FormValue is a shortcut to get a value by name inside the request form instead of parsing the whole form.
func (request RequestContext) FormValue(key string) string {
	return request.underlying.FormValue(key)
}

// FormFile is a shortcut to get a file with the given name from multipart form.
func (request RequestContext) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return request.underlying.FormFile(key)
}

// PostFormValue can get a value by the given name from request post-form.
func (request RequestContext) PostFormValue(key string) string {
	return request.underlying.PostFormValue(key)
}

// ParseMultipartForm is the manual approach to parse the request's entity to multipart form.
// Please read (*http.Request).ParseMultipartForm for more detailed information.
func (request RequestContext) ParseMultipartForm(maxMemory int64) error {
	return request.underlying.ParseMultipartForm(maxMemory)
}

// Form returns all the key-values inside the given request.
// It calls ParseForm itself.
func (request RequestContext) Form() (map[string][]string, error) {
	if err := request.underlying.ParseForm(); err != nil {
		return nil, err
	}
	return request.underlying.Form, nil
}

// PostForm returns all the key-values inside the given request's post-form.
func (request RequestContext) PostForm() (map[string][]string, error) {
	if err := request.underlying.ParseForm(); err != nil {
		return nil, err
	} else {
		return request.underlying.PostForm, nil
	}
}

// AddCookie adds the cookie to the request.
func (request RequestContext) AddCookie(cookie *http.Cookie) {
	request.underlying.AddCookie(cookie)
}

// RequestBody holds the bytes of the request's body entity.
type RequestBody struct {
	underlying      io.Reader
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

func bodyFromReadCloser(reader io.ReadCloser) (*RequestBody, error) {
	defer func(r io.ReadCloser) {
		err := r.Close()
		if err != nil {
			_ = stginLogger.Err("could not close reader stream from request")
		}
	}(reader)
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	} else {
		return bodyFromBytes(bytes), nil
	}
}

func requestContextFromHttpRequest(request *http.Request, writer http.ResponseWriter, pathParams Params) RequestContext {
	var pusher http.Pusher
	var isSupported bool
	if writer != nil {
		pusher, isSupported = writer.(http.Pusher)
	}
	var headers http.Header
	if request.Header == nil {
		headers = emptyHeaders
	} else {
		headers = request.Header
	}
	return RequestContext{
		Url:         request.URL.Path,
		QueryParams: Queries{request.URL.Query()},
		PathParams:  PathParams{pathParams},
		Headers:     headers,
		Trailer:     request.Trailer,
		Body: func() *RequestBody {
			var body *RequestBody
			if request.Body != nil {
				body, _ = bodyFromReadCloser(request.Body)
			}
			return body
		},
		receivedAt:    time.Now(),
		Method:        request.Method,
		ContentLength: request.ContentLength,
		Host:          request.Host,
		MultipartForm: func() *multipart.Form {
			return request.MultipartForm
		},
		Scheme:     request.URL.Scheme,
		RemoteAddr: request.RemoteAddr,
		underlying: request,
		HttpPush: Push{
			IsSupported: isSupported,
			pusher:      pusher,
		},
	}
}

func (body *RequestBody) fillAndGetBytes() ([]byte, *MalformedRequestContext) {
	if body.hasFilledBytes {
		return body.underlyingBytes, nil
	} else {
		bytes, err := io.ReadAll(body.underlying)
		if err != nil {
			return []byte{}, &MalformedRequestContext{details: err.Error()}
		}
		body.underlyingBytes = bytes
		body.hasFilledBytes = true
		return bytes, nil
	}
}

// SafeJSONInto receives a pointer to anything, and will try to parse the request bytes into it as JSON.
// if any error occurs, it is returned immediately by the function.
func (body *RequestBody) SafeJSONInto(a any) error {
	bytes, err := body.fillAndGetBytes()
	if err != nil {
		return *err
	}
	if unmarshalErr := json.Unmarshal(bytes, a); unmarshalErr != nil {
		return ParseError{
			tpe:     "JSON",
			details: unmarshalErr.Error(),
		}
	} else {
		return nil
	}
}

// SafeXMLInto receives a pointer to anything, and will try to parse the request bytes into it as JSON.
// if any error occurs, it is returned immediately by the function.
func (body *RequestBody) SafeXMLInto(a any) error {
	bytes, err := body.fillAndGetBytes()
	if err != nil {
		return *err
	}
	if unmarshalErr := xml.Unmarshal(bytes, a); unmarshalErr != nil {
		return ParseError{
			tpe:     "XML",
			details: unmarshalErr.Error(),
		}
	} else {
		return nil
	}
}

// JSONInto receives a pointer to anything, and will try to parse the request's JSON entity into it.
// It panics in case any error happens.
func (body *RequestBody) JSONInto(a any) {
	bytes, err := body.fillAndGetBytes()
	if err != nil {
		panic(err)
	}
	if unmarshalErr := json.Unmarshal(bytes, a); unmarshalErr != nil {
		panic(ParseError{details: err.Error(), tpe: "JSON"})
	}
}

// XMLInto receives a pointer to anything, and will try to parse the request's XML entity into it.
// It panics in case any error happens.
func (body *RequestBody) XMLInto(a any) {
	bytes, err := body.fillAndGetBytes()
	if err != nil {
		panic(err)
	}
	if unmarshalErr := xml.Unmarshal(bytes, a); unmarshalErr != nil {
		panic(ParseError{
			tpe:     "XML",
			details: err.Error(),
		})
	}
}
