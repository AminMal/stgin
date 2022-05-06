package stgin

import (
	"net/http"
	"time"
)

type RequestContext struct {
	Url         string
	QueryParams map[string][]string
	PathParams  map[string]string
	Headers     http.Header
	Body        *RequestBody
	receivedAt  time.Time
	Method      string
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
	} else { return "", false }
}
