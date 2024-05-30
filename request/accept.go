package request

import "net/http"

// Accept sets the canonical Accept header on a request
func Accept(contentType string) func(rq *http.Request) error {
	return func(rq *http.Request) error {
		rq.Header.Add("Accept", contentType)
		return nil
	}
}

// Accept sets the canonical Accept header with a value of "application/json"
func AcceptJSON() func(rq *http.Request) error {
	return func(rq *http.Request) error {
		rq.Header.Add("Accept", "application/json")
		return nil
	}
}
