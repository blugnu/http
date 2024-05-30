package request

import "errors"

var (
	ErrInvalidJSON      = errors.New("invalid json")
	ErrMarshallingJSON  = errors.New("error marshalling json")
	ErrSetBoundary      = errors.New("SetBoundary error")
	ErrTooManyArguments = errors.New("too many arguments")
	ErrInvalidQuery     = errors.New("invalid query")
)
