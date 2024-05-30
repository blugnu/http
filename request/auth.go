package request

import (
	"context"
	"fmt"
	"net/http"

	"github.com/blugnu/errorcontext"
)

// BearerToken sets a canonical Authorisation header with a BearerToken value,
// using the result of a provided function.
//
// The token value is not supplied directly; instead, the provided function will
// be called to obtain a token, or an error if a token is not available.
func BearerToken(fn func(context.Context) (string, error)) func(*http.Request) error {
	return func(rq *http.Request) error {
		ctx := rq.Context()

		t, err := fn(ctx)
		if err != nil {
			return errorcontext.Errorf(ctx, "BearerToken: %w", err)
		}

		rq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t))

		return nil
	}
}
