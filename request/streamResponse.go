package request

import "net/http"

// canonical casing avoids go-staticcheck flagging the constant with SA1008
const StreamResponseHeader = "X-Blugnu-Http-Stream-Response"

// StreamResponse adds a request header indicating that the client expects
// to stream the response body.  The header is removed 
//
// If specified, the usual reading of the response body prior to returning
// the response to the caller is skipped.
func StreamResponse() func(*http.Request) {
	return func(r *http.Request) {
		r.Header[StreamResponseHeader] = []string{"true"}
	}
}
