package request

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// canonical casing avoids go-staticcheck flagging the constant with SA1008
const AcceptStatusHeader = "X-Blugnu-Http-Accept-Status"

func AcceptStatus(statusCodes ...int) func(*http.Request) error {
	return func(rq *http.Request) error {
		handle := func(err error) error {
			return fmt.Errorf("request.AcceptStatus: %w", err)
		}

		acc := []int{http.StatusOK}
		if h, ok := rq.Header[AcceptStatusHeader]; ok {
			if err := json.Unmarshal([]byte(h[0]), &acc); err != nil {
				return handle(fmt.Errorf("%w: %w", ErrInvalidJSON, err))
			}
		}

		acc = append(acc, statusCodes...)

		// we can safely ignore the returned error value as marshalling a
		// slice of int cannot error.  This avoids creating an irrelevant
		// and untestable code path
		h, _ := json.Marshal(acc)
		rq.Header[AcceptStatusHeader] = []string{string(h)}
		return nil
	}
}
