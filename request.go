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

func (c RequestContext) Cookie(name string) (*http.Cookie, error) {
	return c.underlying.Cookie(name)
}

func (c RequestContext) AddCookie(cookie *http.Cookie) {
	c.underlying.AddCookie(cookie)
}

func (c RequestContext) GetPathParam(name string) (string, bool) {
	var res string
	var found bool
	for _, param := range c.PathParams {
		if param.key == name {
			found = true
			res = param.value
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
	var res []string
	for queryName, values := range c.QueryParams {
		if queryName == name {
			res = values
		}
	}
	return res
}

func (c RequestContext) GetQuery(name string) (string, bool) {
	allValues := c.GetQueries(name)
	if len(allValues) == 1 {
		return allValues[0], true
	} else {
		return "", false
	}
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
	body, _ := bodyFromReadCloser(request.Body)
	pusher, isSupported := writer.(http.Pusher)
	return RequestContext{
		Url:           request.URL.Path,
		QueryParams:   request.URL.Query(),
		PathParams:    pathParams,
		Headers:       request.Header,
		Body:          body,
		receivedAt:    time.Now(),
		Method:        request.Method,
		ContentLength: request.ContentLength,
		Host:          request.Host,
		MultipartForm: func() *multipart.Form {
			return request.MultipartForm
		},
		Scheme:        request.URL.Scheme,
		RemoteAddr:    request.RemoteAddr,
		underlying:    request,
		HttpPush:      Push{
			IsSupported: isSupported,
			pusher: pusher,
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
