package stgin

import "fmt"

type ParseError struct {
	tpe     string
	details string
}

func (ije ParseError) Error() string {
	return fmt.Sprintf("malformed %v context, %v", ije.tpe, ije.details)
}

type MalformedRequestContext struct {
	details string
}

func (mrc MalformedRequestContext) Error() string {
	return fmt.Sprintf("could not read from request, %v", mrc.details)
}
