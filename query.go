package stgin

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type queryDecl = map[string]string

type Queries struct {
	All map[string][]string
}

func (q Queries) Get(key string) ([]string, bool) {
	values, found := q.All[key]
	return values, found
}

func (q Queries) GetOne(key string) (string, bool) {
	all := q.All[key]
	if len(all) == 1 {
		return all[0], true
	} else {
		return "", false
	}
}

func (q Queries) MustGet(key string) string {
	if value, ok := q.GetOne(key); ok {
		return value
	} else {
		panic("query parameter entry " + key + " had either 0 or more than 1 value (must've been exactly one)")
	}
}

func (q Queries) GetInt(key string) (int, error) {
	stringed, ok := q.GetOne(key)
	if !ok {
		return 0, errors.New("query parameter entry " + key + " had either 0 or more than 1 value (must've been exactly one)")
	}
	return strconv.Atoi(stringed)
}

func (q Queries) MustGetInt(key string) int {
	value, err := q.GetInt(key)
	if err != nil {
		panic(err)
	}
	return value
}

func (q Queries) GetFloat(key string) (float64, error) {
	stringed, ok := q.GetOne(key)
	if !ok {
		return 0, errors.New("query parameter entry " + key + " had either 0 or more than 1 value (must've been exactly one)")
	}
	return strconv.ParseFloat(stringed, 64)
}

func (q Queries) MustGetFloat(key string) float64 {
	value, err := q.GetFloat(key)
	if err != nil {
		panic(err)
	}
	return value
}

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

func acceptsQuery(tpe string, value string) bool {
	if tpe == "" {
		return true
	} else {
		return getQueryMatcher(tpe).Match([]byte(value))
	}
}

func acceptsAllQueries(q queryDecl, qs map[string][]string) bool {
	var accepts = true
	for name, tpe := range q {
		value := qs[name]
		if len(value) == 0 {
			accepts = false
			break
		}
		for _, v := range value {
			if v == "" || !acceptsQuery(tpe, v) {
				accepts = false
				break
			}
		}
	}
	return accepts
}

func getQueryDefinitionsFromPattern(pattern string) queryDecl {
	defs := strings.SplitN(pattern, "&", -1)
	qs := make(queryDecl, 10)
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
			query, found := c.QueryParams.GetOne(queryName)
			valueField := value.Field(i)
			if found {
				switch valueField.Kind() {
				case reflect.String:
					valueField.Set(reflect.ValueOf(query))
				case reflect.Int:
					if queryInt, err := strconv.Atoi(query); err == nil {
						valueField.Set(reflect.ValueOf(queryInt))
					}
				case reflect.Float64:
					if queryFloat, err := strconv.ParseFloat(query, 64); err == nil {
						valueField.Set(reflect.ValueOf(queryFloat))
					}
				case reflect.Float32:
					if queryFloat, err := strconv.ParseFloat(query, 32); err == nil {
						valueField.Set(reflect.ValueOf(queryFloat))
					}
				}
			}
		}
	}
	return nil
}
