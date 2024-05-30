package request

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/blugnu/http/multipart"
)

// MultipartFormDataFromMap configures a multipart form data body by mapping
// the items in a map to the parts of the form data.
//
// A map must be provided together with a function to provide the field id,
// filename and content bytes of each part, given the key and value of an
// item in the map (or an error if a valid part is unable to be derived).
//
// The parts are added in order of map keys, after those keys have been sorted.
func MultipartFormDataFromMap[K comparable, V any](
	m map[K]V,
	opts ...func(multipart.Options),
) func(*http.Request) error {
	return func(rq *http.Request) error {
		handle := func(err error) error {
			rq.Body = nil
			return fmt.Errorf("MultipartFormDataFromMap: %w", err)
		}

		ct, body, err := multipart.BodyFromMap(m, opts...)
		if err != nil {
			return handle(err)
		}

		rq.Header.Set("Content-Type", ct)
		rq.Body = io.NopCloser(bytes.NewReader(body))
		rq.ContentLength = int64(len(body))

		return nil
	}
}
