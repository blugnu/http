package request

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

// cpy is a reference to a function to copy bytes from a src slice to a destination.
// It is a variable to facilitate testing of scenarios where a copy operation might
// fail.
var cpy = func(dst, src []byte) int { return copy(dst, src) }

// ErrCopyFailed is the error returned by the Body() RequestOption if the supplied
// byte slice cannot be completely copied to the request.Body.
var ErrCopyFailed = errors.New("copy() operation failed or was incomplete")

// Body sets the body of a request to the contents of a supplied byte slice
// and the ContentLength to the length of the slice.
//
// request.ErrCopyFailed is returned if the provided slice cannot be completely
// copied to the request Body.
func Body(data []byte) func(*http.Request) error {
	return func(rq *http.Request) error {
		b := make([]byte, len(data))
		if cpy(b, data) < len(data) {
			return ErrCopyFailed
		}

		rq.Body = io.NopCloser(bytes.NewReader(b))
		rq.ContentLength = int64(len(b))

		return nil
	}
}
