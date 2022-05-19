package stgin

import (
	"errors"
	"strconv"
)

// PathParams is a struct wrapped around the path parameters of an HTTP request.
// It provides some receiver functions which make it easier than ever to use them.
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

// GetInt tries to find a path parameter by the given key, and tries to convert it to integer.
// In case any error happens from finding to converting, this function returns immediately.
func (pp PathParams) GetInt(key string) (int, error) {
	value, err := pp.getOrErr(key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

// MustGetInt tries to find a path parameter by the given key, and tries to convert it to integer.
// This function should be used when you have defined your path parameter to be integer inside the route pattern
// (like "/users/$id:int"). This function will panic in case of any error.
func (pp PathParams) MustGetInt(key string) int {
	intValue, err := pp.GetInt(key)
	if err != nil {
		panic(err)
	}
	return intValue
}

// GetFloat tries to find a path parameter by the given key, and tries to convert it to float.
// In case any error happens from finding to converting, this function returns immediately.
func (pp PathParams) GetFloat(key string) (float64, error) {
	value, err := pp.getOrErr(key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(value, 64)
}

// MustGetFloat tries to find a path parameter by the given key, and tries to convert it to float.
// This function should be used when you have defined your path parameter to be floating point inside the route pattern
// (like "/geo/$lat:float"). This function will panic in case of any error.
func (pp PathParams) MustGetFloat(key string) float64 {
	value, err := pp.GetFloat(key)
	if err != nil {
		panic(err)
	}
	return value
}

// Get tries to find a path parameter by the given key.
func (pp PathParams) Get(key string) (string, bool) {
	value, found := pp.All[key]
	return value, found
}

// MustGet tries to find a path parameter by the given key.
// Returns empty string in case it couldn't be found.
func (pp PathParams) MustGet(key string) string {
	return pp.All[key]
}
