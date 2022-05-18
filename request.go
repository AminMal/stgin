package stgin

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

type RequestContext struct {
	Url           string
	QueryParams   map[string][]string
	PathParams    Params
	Headers       http.Header
	Trailer       http.Header
	Body          *RequestBody
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

type Push struct {
	IsSupported bool
	pusher      http.Pusher
}

func (p Push) Pusher() http.Pusher {
	if !p.IsSupported {
		panic("pusher is not supported in the request")
	}
	return p.pusher
}

func (c RequestContext) Cookies() []*http.Cookie {
	return c.underlying.Cookies()
}

func (c RequestContext) Referer() string { return c.underlying.Referer() }

func (c RequestContext) UserAgent() string { return c.underlying.UserAgent() }

func (c RequestContext) Cookie(name string) (*http.Cookie, error) {
	return c.underlying.Cookie(name)
}

func (c RequestContext) FormValue(key string) string {
	return c.underlying.FormValue(key)
}

func (c RequestContext) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return c.underlying.FormFile(key)
}

func (c RequestContext) PostFormValue(key string) string {
	return c.underlying.PostFormValue(key)
}

func (c RequestContext) ParseMultipartForm(maxMemory int64) error {
	return c.underlying.ParseMultipartForm(maxMemory)
}

func (c RequestContext) Form() (map[string][]string, error) {
	if err := c.underlying.ParseForm(); err != nil {
		return nil, err
	}
	return c.underlying.Form, nil
}

func (c RequestContext) PostForm() (map[string][]string, error) {
	if err := c.underlying.ParseForm(); err != nil {
		return nil, err
	} else {
		return c.underlying.PostForm, nil
	}
}

func (c RequestContext) AddCookie(cookie *http.Cookie) {
	c.underlying.AddCookie(cookie)
}

func (c RequestContext) GetPathParam(name string) (string, bool) {
	var res string
	var found bool
	for key, value := range c.PathParams {
		if key == name {
			found = true
			res = value
			break
		}
	}
	return res, found
}

func (c RequestContext) MustGetPathParam(name string) string {
	value, found := c.GetPathParam(name)
	if !found {
		panic(fmt.Sprintf("used MustGetPathParam while path parameter %s does not exist", name))
	}
	return value
}

func (c RequestContext) GetQueries(name string) []string {
	return c.QueryParams[name]
}

func (c RequestContext) GetQuery(name string) (string, bool) {
	allValues := c.QueryParams[name]
	if len(allValues) == 1 {
		return allValues[0], true
	} else {
		return "", false
	}
}

func (c RequestContext) MustGetQuery(name string) string {
	query, found := c.GetQuery(name)
	if !found {
		panic(fmt.Sprintf("used MustGetQuery while query parameter %s does not exist, or has more than one value", name))
	}
	return query
}

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

func requestContextFromHttpRequest(request *http.Request, writer http.ResponseWriter, pathParams Params) RequestContext {
	var body *RequestBody
	if request.Body != nil {
		body, _ = bodyFromReadCloser(request.Body)
	} else {
		body = nil
	}
	pusher, isSupported := writer.(http.Pusher)
	var headers http.Header
	if request.Header == nil {
		headers = emptyHeaders
	} else {
		headers = request.Header
	}
	return RequestContext{
		Url:           request.URL.Path,
		QueryParams:   request.URL.Query(),
		PathParams:    pathParams,
		Headers:       headers,
		Trailer:       request.Trailer,
		Body:          body,
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
		bytes, err := ioutil.ReadAll(body.underlying)
		if err != nil {
			return []byte{}, &MalformedRequestContext{details: err.Error()}
		}
		body.underlyingBytes = bytes
		body.hasFilledBytes = true
		return bytes, nil
	}
}

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

func (body *RequestBody) JSONInto(a any) {
	bytes, err := body.fillAndGetBytes()
	if err != nil {
		panic(err)
	}
	if unmarshalErr := json.Unmarshal(bytes, a); unmarshalErr != nil {
		panic(ParseError{details: err.Error(), tpe: "JSON"})
	}
}

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
