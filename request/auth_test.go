package request

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/blugnu/test"
)

func TestAuth(t *testing.T) {
	// ARRANGE
	tokenerr := errors.New("token error")

	testcases := []struct {
		scenario string
		act      func(*http.Request) error
		assert   func(*testing.T, *http.Request, error)
	}{
		// BearerToken tests
		{scenario: "BearerToken/token error",
			act: func(rq *http.Request) error {
				return BearerToken(func(context.Context) (string, error) { return "", tokenerr })(rq)
			},
			assert: func(t *testing.T, rq *http.Request, err error) {
				test.Error(t, err).Is(tokenerr)
				test.Value(t, rq.Header.Get("Authorisation")).Equals("")
			},
		},
		{scenario: "BearerToken/token ok",
			act: func(rq *http.Request) error {
				return BearerToken(func(context.Context) (string, error) { return "token-value", nil })(rq)
			},
			assert: func(t *testing.T, rq *http.Request, err error) {
				test.Error(t, err).IsNil()
				test.Value(t, rq.Header.Get("Authorization")).Equals("Bearer token-value")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			rq, err := http.NewRequest(http.MethodTrace, "notused", nil)
			test.Error(t, err).IsNil()

			tc.assert(t, rq, tc.act(rq))
		})
	}
}
