package request

import (
	"net/http"
)

// Header sets the value of a canonical header.
//
// Canonical header keys are normalised; normalising a non-canonical
// header may result in unintended changes to the case of the key.
//
// To set a non-canonical header use RawHeader() instead.
//
// Example:
//
//	// sets the canonical "Content-Type" header
//	Header("content-type", "application/json")
func Header(k, v string) func(*http.Request) error {
	return func(rq *http.Request) error {
		rq.Header.Set(k, v)
		return nil
	}
}

// NonCanonicalHeader sets the value of a header without canonicalising the key.
//
// When setting a NonCanonicalHeader() the key is applied exactly as-specified;
// this may be important for non-canonical keys but is undesirable for
// canonical headers.
//
// If setting a canonical header, use Header() instead.
//
// Example:
//
//	// sets a non-canonical "sessionid" header
//	NonCanonicalHeader("sessionid", id)
func NonCanonicalHeader(k, v string) func(*http.Request) error {
	return func(rq *http.Request) error {
		rq.Header[k] = []string{v}
		return nil
	}
}
