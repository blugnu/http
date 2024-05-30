package http

import (
	"errors"
	"fmt"
)

var (
	ErrInitialisingClient   = errors.New("error initialising client")
	ErrInitialisingRequest  = errors.New("error initialising request")
	ErrInvalidJSON          = errors.New("invalid json")
	ErrInvalidRequestHeader = errors.New("invalid request headers")
	ErrInvalidURL           = errors.New("invalid url")
	ErrMaxRetriesExceeded   = errors.New("http retries exceeded")
	ErrNoResponseBody       = errors.New("response body was empty")
	ErrReadingResponseBody  = errors.New("error reading response body")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")

	// errors related to the mock client
	ErrCannotChangeExpectations = errors.New("expectations cannot be changed")
	ErrUnexpectedRequest        = errors.New("unexpected request")
)

// MockExpectationsError is the error returned by ExpectationsNotMet() when one or
// more configured expectations have not been met.  It wraps all errors
// representing the failed expectations.
type MockExpectationsError struct {
	name   string
	errors []error
}

// Error implements the error interface for MockExpectationsError by returning a
// string representation of the error, presenting each wrapped error indented
// under a summary identifying the mock client to which the failures relate.
func (err MockExpectationsError) Error() string {
	errs := ""
	for _, err := range err.errors {
		errs += fmt.Sprintf("   %s\n", err.Error())
	}
	return fmt.Sprintf("%s: expectations not met: [\n%s]", err.name, errs)
}
