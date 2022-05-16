package stgin

import (
	"encoding/json"
	"encoding/xml"
	"reflect"
	"testing"
)

func TestText(t *testing.T) {
	text := "Hello, world!"
	response := Ok(Text("Hello, world!"))
	responseBytes, err := response.Entity.Bytes()
	if err != nil {
		t.Errorf("error while retrieving entity bytes, %s", err.Error())
	}
	if !reflect.DeepEqual(responseBytes, []byte(text)) {
		t.Fatal("text bytes and entity bytes are not equal")
	}
}

type testStruct struct {
	Message		string `json:"message"`
	Status 		bool `json:"status"`
}

func TestJson(t *testing.T) {
	obj := testStruct{
		Message: "Hello, world!",
		Status:  true,
	}
	jsonBytes, _ := json.Marshal(&obj)
	response := Ok(Json(&obj))
	responseBytes, err := response.Entity.Bytes()
	if err != nil {
		t.Errorf("error while retrieving entity bytes, %s", err.Error())
	}
	if !reflect.DeepEqual(jsonBytes, responseBytes) {
		t.Fatal("json bytes and response entity bytes are not equal")
	}
}

func TestXml(t *testing.T) {
	obj := testStruct{
		Message: "Hello, world!",
		Status:  true,
	}
	xmlBytes, _ := xml.Marshal(&obj)
	response := Ok(Xml(&obj))
	responseBytes, err := response.Entity.Bytes()
	if err != nil {
		t.Errorf("error while retrieving entity bytes, %s", err.Error())
	}
	if !reflect.DeepEqual(xmlBytes, responseBytes) {
		t.Fatal("xml bytes and response entity bytes are not equal")
	}
}