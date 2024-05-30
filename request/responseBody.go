package request

import "net/http"

// canonical casing avoids go-staticcheck flagging the constant with SA1008
const ResponseBodyRequiredHeader = "X-Blugnu-Http-Response-Body-Required"

// ResponseBodyRequired establishes that a non-empty response body is expected
// in response to this request.  If the response provides an empty body the
// client will return an http.ErrNoResponseBody error, together with the
// response
func ResponseBodyRequired() func(*http.Request) error {
	return func(rq *http.Request) error {
		rq.Header[ResponseBodyRequiredHeader] = []string{"true"}
		return nil
	}
}
