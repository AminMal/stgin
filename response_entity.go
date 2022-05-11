package stgin

import (
	"encoding/json"
	"encoding/xml"
)

const (
	applicationJson = "application/json"
	applicationXml  = "application/xml"
	plainText       = "text/plain"
)

type ResponseEntity interface {
	ContentType() string
	Bytes() ([]byte, error)
}

func marshall(re ResponseEntity) (bytes []byte, contentType string, err error) {
	bytes, err = re.Bytes()
	contentType = re.ContentType()
	return
}

type jsonEntity struct {
	obj any
}

func (j jsonEntity) ContentType() string {
	return applicationJson
}

func (j jsonEntity) Bytes() ([]byte, error) {
	return json.Marshal(j.obj)
}

type xmlEntity struct {
	obj any
}

func (xe xmlEntity) ContentType() string {
	return applicationXml
}

func (xe xmlEntity) Bytes() ([]byte, error) {
	return xml.Marshal(xe.obj)
}

type textEntity struct {
	obj string
}

func (t textEntity) ContentType() string {
	return plainText
}

func (t textEntity) Bytes() ([]byte, error) {
	return []byte(t.obj), nil
}

func Json(a any) ResponseEntity {
	return jsonEntity{obj: a}
}

func Xml(a any) ResponseEntity {
	return xmlEntity{obj: a}
}

func Text(text string) ResponseEntity {
	return textEntity{obj: text}
}
