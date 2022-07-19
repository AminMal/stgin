package stgin

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// RequestChangeable holds only changeable data for request that can be modified in pre-api time.
type RequestChangeable struct {
	queries		 	Queries
	pathParams 		PathParams
	headers 		http.Header
}

func (changeable *RequestChangeable) Queries() Queries { return changeable.queries }
func (changeable *RequestChangeable) AddQuery(key string, values []string) {
	changeable.queries.All[key] = append(changeable.queries.All[key], values...)
}
func (changeable *RequestChangeable) SetQueries(key string, values []string) {
	changeable.queries.All[key] = values
}

func (changeable *RequestChangeable) PathParams() PathParams { return changeable.pathParams }
func (changeable *RequestChangeable) SetPathParam(key, value string) {
	changeable.pathParams.All[key] = value
}

func (changeable *RequestChangeable) Headers() http.Header { return changeable.headers }
func (changeable *RequestChangeable) AddHeader(key, value string) {
	changeable.headers.Add(key, value)
}
func (changeable *RequestChangeable) AddRawHeader(key, value string) {
	changeable.headers[key] = append(changeable.headers[key], value)
}
func (changeable *RequestChangeable) SetHeader(key string, value string) {
	changeable.headers.Set(key, value)
}
func (changeable *RequestChangeable) SetRawHeader(key string, values []string) {
	changeable.headers[key] = values
}

type RequestContext struct {
	url 			string
	queryParams 	Queries
	pathParams  	PathParams
	headers  		http.Header
	trailer     	http.Header
	Body          	func() *RequestBody
	receivedAt  	time.Time
	method 			string
	contentLength 	int64
	host 			string
	MultipartForm 	func() *multipart.Form
	scheme 			string
	remoteAddr  	string
	underlying  	*http.Request
	httpPush    	Push
}

func (rc *RequestContext) Url() string               { return rc.url }
func (rc *RequestContext) QueryParams() Queries      { return rc.queryParams }
func (rc *RequestContext) PathParams() PathParams    { return rc.pathParams }
func (rc *RequestContext) Headers() http.Header      { return rc.headers }
func (rc *RequestContext) Trailer() http.Header      { return rc.trailer }
func (rc *RequestContext) ReceivedAt() time.Time     { return rc.receivedAt }
func (rc *RequestContext) Method() string            { return rc.method }
func (rc *RequestContext) ContentLength() int64      { return rc.contentLength }
func (rc *RequestContext) Host() string              { return rc.host }
func (rc *RequestContext) Scheme() string            { return rc.scheme }
func (rc *RequestContext) RemoteAddr() string        { return rc.remoteAddr }
func (rc *RequestContext) Underlying() *http.Request { return rc.underlying }
func (rc *RequestContext) Push() Push                { return rc.httpPush }

func (rc *RequestContext) affectChangeable(changeable *RequestChangeable) {
	if changeable.headers != nil { rc.headers = changeable.headers }
	if changeable.queries.All != nil { rc.headers = changeable.headers }
	if changeable.pathParams.All != nil { rc.pathParams = changeable.pathParams }
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
func (rc *RequestContext) Cookies() []*http.Cookie {
	return rc.underlying.Cookies()
}

// Referer returns the value of referer header in http request, returns empty string if it does not exist.
func (rc *RequestContext) Referer() string { return rc.underlying.Referer() }

// UserAgent returns the value of request's user agent, returns empty string if it does not exist.
func (rc *RequestContext) UserAgent() string { return rc.underlying.UserAgent() }

// Cookie tries to find a cookie with the given name.
func (rc *RequestContext) Cookie(name string) (*http.Cookie, error) {
	return rc.underlying.Cookie(name)
}

// FormValue is a shortcut to get a value by name inside the request form instead of parsing the whole form.
func (rc *RequestContext) FormValue(key string) string {
	return rc.underlying.FormValue(key)
}

// FormFile is a shortcut to get a file with the given name from multipart form.
func (rc *RequestContext) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return rc.underlying.FormFile(key)
}

// PostFormValue can get a value by the given name from request post-form.
func (rc *RequestContext) PostFormValue(key string) string {
	return rc.underlying.PostFormValue(key)
}

// ParseMultipartForm is the manual approach to parse the request's entity to multipart form.
// Please read (*http.Request).ParseMultipartForm for more detailed information.
func (rc *RequestContext) ParseMultipartForm(maxMemory int64) error {
	return rc.underlying.ParseMultipartForm(maxMemory)
}

// Form returns all the key-values inside the given request.
// It calls ParseForm itself.
func (rc *RequestContext) Form() (map[string][]string, error) {
	if err := rc.underlying.ParseForm(); err != nil {
		return nil, err
	}
	return rc.underlying.Form, nil
}

// PostForm returns all the key-values inside the given request's post-form.
func (rc *RequestContext) PostForm() (map[string][]string, error) {
	if err := rc.underlying.ParseForm(); err != nil {
		return nil, err
	} else {
		return rc.underlying.PostForm, nil
	}
}

// AddCookie adds the cookie to the request.
func (rc *RequestContext) AddCookie(cookie *http.Cookie) {
	rc.underlying.AddCookie(cookie)
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

func requestContextFromHttpRequest(request *http.Request, writer http.ResponseWriter, pathParams Params) *RequestContext {
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
	return &RequestContext{
		url:         request.URL.Path,
		queryParams: Queries{request.URL.Query()},
		pathParams:  PathParams{pathParams},
		headers:     headers,
		trailer:     request.Trailer,
		Body: func() *RequestBody {
			var body *RequestBody
			if request.Body != nil {
				body, _ = bodyFromReadCloser(request.Body)
			}
			return body
		},
		receivedAt:    time.Now(),
		method:        request.Method,
		contentLength: request.ContentLength,
		host:          request.Host,
		MultipartForm: func() *multipart.Form {
			return request.MultipartForm
		},
		scheme:     request.URL.Scheme,
		remoteAddr: request.RemoteAddr,
		underlying: request,
		httpPush: Push{
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
