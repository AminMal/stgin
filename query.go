package stgin

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type queries = map[string]string

func getQueryMatcher(tpe string) *regexp.Regexp {
	switch tpe {
	case "int":
		return intRegex
	case "float":
		return floatRegex
	default:
		return stringRegex
	}
}

func acceptsQuery(q queries, key string, value string) bool {
	tpe := q[key]
	if tpe == "" { return true } else {
		return getQueryMatcher(tpe).Match([]byte(value))
	}
}

func acceptsAllQueries(q queries, qs map[string][]string) bool {
	var accepts = true
	for name, values := range qs {
		for _, v := range values {
			if !acceptsQuery(q, name, v) {
				accepts = false
				break
			}
		}
	}
	return accepts
}

func getQueryDefinitionsFromPattern(pattern string) queries {
	defs := strings.SplitN(pattern, "&", -1)
	qs := make(queries, 10)
	for _, def := range defs {
		if def != "" {
			arr := strings.SplitN(def, ":", 2)
			name := arr[0]
			var tpe = "string"
			if len(arr) == 2 {
				tpe = arr[1]
			}
			qs[name] = tpe
		}
	}
	return qs
}

func (c RequestContext) QueryToObj(a any) error {
	if reflect.TypeOf(a).Kind() != reflect.Ptr {
		return errors.New("passed raw type instead of value pointer to QueryToObj function, please use pointers instead")
	}
	if reflect.ValueOf(a).Kind() != reflect.Ptr {
		return errors.New("passed raw value instead of value pointer to QueryToObj function, please use pointers instead")
	}
	tpe := reflect.TypeOf(a).Elem()
	value := reflect.ValueOf(a).Elem()
	for i := 0; i < tpe.NumField(); i++ {
		tpeField := tpe.Field(i)
		if tpeField.IsExported() {
			queryTag, _ := tpeField.Tag.Lookup("qp")
			if queryTag == "" {
				queryTag = tpeField.Name
			}
			tags := strings.SplitN(queryTag, ",", -1)
			queryName := tags[0]
			//otherTags := tags[1:]  WIP
			query, found := c.GetQuery(queryName)
			valueField := value.Field(i)
			if found {
				switch valueField.Kind() {
				case reflect.String:
					valueField.Set(reflect.ValueOf(query))
				case reflect.Int:
					if queryInt, err := strconv.Atoi(query); err != nil {
						valueField.Set(reflect.ValueOf(queryInt))
					}
				case reflect.Float64:
					if queryFloat, err := strconv.ParseFloat(query, 64); err != nil {
						valueField.Set(reflect.ValueOf(queryFloat))
					}
				case reflect.Float32:
					if queryFloat, err := strconv.ParseFloat(query, 32); err != nil {
						valueField.Set(reflect.ValueOf(queryFloat))
					}
				}
			}
		}
	}
	return nil
}
