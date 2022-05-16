package stgin

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	applicationJson = "application/json"
	applicationXml  = "application/xml"
	plainText       = "text/plain"
)

var getFileFormatRegex = regexp.MustCompile(".*\\.(.+)")

type ResponseEntity interface {
	ContentType() string
	Bytes() ([]byte, error)
}

func marshall(re ResponseEntity) (bytes []byte, contentType string, err error) {
	bytes, err = re.Bytes()
	contentType = re.ContentType()
	return
}

type emptyEntity struct {}

func (e emptyEntity) ContentType() string {
	return plainText
}

func (e emptyEntity) Bytes() ([]byte, error) {
	return []byte{}, nil
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

type fileContent struct {
	path string
}

func (f fileContent) ContentType() string {
	isValidPath := getFileFormatRegex.MatchString(f.path)
	if !isValidPath {
		panic(fmt.Sprintf("not a valid filepath: '%s'", f.path))
	}
	fileType := getFileFormatRegex.FindStringSubmatch(f.path)[1]
	switch strings.ToLower(fileType) {
	case "html":
		return "text/html"
	case "htm":
		return "text/html"
	case "jpeg":
		return "image/jpeg"
	case "jpg":
		return "image/jpeg"
	case "pdf":
		return "application/pdf"
	case "png":
		return "image/png"
	case "rar":
		return "application/vnd.rar"
	case "mp3":
		return "audio/mpeg"
	case "mp4":
		return "video/mp4"
	case "mpeg":
		return "video/mpeg"
	case "ppt":
		return "application/vnd.ms-powerpoint"
	case "svg":
		return "image/svg+xml"
	case "wav":
		return "audio/wav"
	case "txt":
		return "text/plain"
	default:
		return "text/plain"
	}
}

func (f fileContent) Bytes() ([]byte, error) {
	return os.ReadFile(f.path)
}

type dirPlaceholder struct {
	path string
}

func (d dirPlaceholder) ContentType() string {
	return "text/html" // will be overriden by http handler func by default
}

func (d dirPlaceholder) Bytes() ([]byte, error) {
	return []byte{}, nil // will be handled by go http handler func
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

func Empty() ResponseEntity { return emptyEntity{} }