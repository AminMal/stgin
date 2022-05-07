package stgin

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

type RequestContext struct {
	Url           string
	QueryParams   map[string][]string
	PathParams    map[string]string
	Headers       http.Header
	Body          *RequestBody
	receivedAt    time.Time
	Method        string
	ContentLength int64
	Host          string
	MultipartForm func() *multipart.Form
	Scheme        string
	RemoteAddr    string
}

func (c RequestContext) GetPathParam(name string) (string, bool) {
	var res string
	var found bool
	for paramName, value := range c.PathParams {
		if paramName == name {
			found = true
			res = value
			break
		}
	}
	return res, found
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