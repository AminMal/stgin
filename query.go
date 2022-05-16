package stgin

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

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
