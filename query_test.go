package stgin

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

type Q struct {
	Query    string `qp:"query"`
	Name     string `qp:"name"`
	Untagged string
}

func TestRequestContext_QueryToObj(t *testing.T) {
	uri, _ := url.Parse("/test/queries?query=search&name=John&Untagged=used")
	req := http.Request{
		Method:           "GET",
		URL:              uri,
		Header:           emptyHeaders,
		RequestURI:       "/test/queries?query=search&name=John&Untagged=used",
	}
	rc := requestContextFromHttpRequest(&req, nil, Params{})
	emptyQuery := Q{
		Query: "search",
		Name: "John",
	}
	err := rc.QueryToObj(&emptyQuery)
	if err != nil {
		t.Errorf("Failed creating query object: %s", err.Error())
	}
	if !reflect.DeepEqual(emptyQuery, Q{Name: "John", Query: "search", Untagged: "used"}) {
		t.Fatal("failed")
	}
}
