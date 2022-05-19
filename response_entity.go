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

var getFileFormatRegex = regexp.MustCompile(".*\\.(.+)$")

/*
	ResponseEntity is an interface representing anything that can be sent through http response body.
	Structs implementing ResponseEntity must have a content type (which is written directly in the response),
	And also a function which can provide entity bytes, or any error if exists.
	So for instance if you wanted to define a custom PDF ResponseEntity,

	type PDF struct {
        filepath string
    }
	func (pdf PDF) ContentType() string { return "application/pdf" }
	func (pdf PDF) Bytes() ([]byte, error) { ... read file ... }

	And simply use it inside your APIs.

	return stgin.Ok(PDF{filePath})
*/
type ResponseEntity interface {
	// ContentType represents *HTTP* response content type of the entity.
	ContentType() string
	// Bytes function is responsible to provide response entity bytes, and eny error if exists.
	Bytes() ([]byte, error)
}

func marshall(re ResponseEntity) (bytes []byte, contentType string, err error) {
	bytes, err = re.Bytes()
	contentType = re.ContentType()
	return
}

type emptyEntity struct{}

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

// Json is a shortcut to convert any object into a JSON ResponseEntity.
func Json(a any) ResponseEntity {
	return jsonEntity{obj: a}
}

// Xml is a shortcut to convert any object into an XML ResponseEntity.
func Xml(a any) ResponseEntity {
	return xmlEntity{obj: a}
}

// Text is a shortcut to convert any object into a text ResponseEntity.
func Text(text string) ResponseEntity {
	return textEntity{obj: text}
}

// Empty is used when you want to return empty responses to the client.
// There are situations where status codes talk, and there is no need to populate response body with non-meaningful data.
func Empty() ResponseEntity { return emptyEntity{} }
