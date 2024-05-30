package request

import (
	"net/http"
	"strconv"
)

// canonical casing avoids go-staticcheck flagging the constant with SA1008
const MaxRetriesHeader = "X-Blugnu-Http-Max-Retries"

// MaxRetries configures a maximum number of retries on a specific request.
// If set, this overrides any MaxRetries that may be configured on the client
// used to make the request.
//
// e.g. if the client is configured with MaxRetries == 5 and a request is
// submitted with MaxRetries == 3, then at most 4 attempts will be made: the
// initial request and at most 3 retry attempts
func MaxRetries(n uint) func(*http.Request) error {
	return func(rq *http.Request) error {
		rq.Header[MaxRetriesHeader] = []string{strconv.Itoa(int(n))}
		return nil
	}
}
