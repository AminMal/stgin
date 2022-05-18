package stgin

import (
	"errors"
	"strconv"
)

type PathParams struct {
	All map[string]string
}

func (pp PathParams) getOrErr(key string) (string, error) {
	value, found := pp.All[key]
	if !found {
		return "", errors.New("not found path parameter " + key)
	}
	return value, nil
}

func (pp PathParams) GetInt(key string) (int, error) {
	value, err := pp.getOrErr(key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

func (pp PathParams) MustGetInt(key string) int {
	intValue, err := pp.GetInt(key)
	if err != nil {
		panic(err)
	}
	return intValue
}

func (pp PathParams) GetFloat(key string) (float64, error) {
	value, err := pp.getOrErr(key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(value, 64)
}

func (pp PathParams) MustGetFloat(key string) float64 {
	value, err := pp.GetFloat(key)
	if err != nil {
		panic(err)
	}
	return value
}

func (pp PathParams) Get(key string) (string, bool) {
	value, found := pp.All[key]
	return value, found
}

func (pp PathParams) MustGet(key string) string {
	return pp.All[key]
}
