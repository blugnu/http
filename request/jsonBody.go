package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// JSONBody sets the body of a request to the contents of a supplied value
// marshalled as JSON.  A Content-Type header is added with the value
// application/json.  The ContentLength is also set to the length of the
// JSON encoded bytes.
func JSONBody(v any) func(*http.Request) error {
	return func(rq *http.Request) error {
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("JSONBody: %w: %w", ErrMarshallingJSON, err)
		}

		rq.Body = io.NopCloser(bytes.NewReader(b))
		rq.ContentLength = int64(len(b))
		rq.Header.Set("Content-Type", "application/json")

		return nil
	}
}
