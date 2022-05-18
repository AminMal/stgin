package stgin

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

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
