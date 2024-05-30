package request

import "net/http"

// ContentType sets the canonical Content-Type header on a request.
func ContentType(s string) func(*http.Request) error {
	return func(rq *http.Request) error {
		rq.Header.Set("Content-Type", s)
		return nil
	}
}
