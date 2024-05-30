package request

import (
	"io"
	"net/http"
	"testing"

	"github.com/blugnu/test"
)

func TestBody(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		act      func(*http.Request) error
		assert   func(*testing.T, *http.Request, error)
	}{
		// Body tests
		{scenario: "Body/copy error",
			act: func(rq *http.Request) error {
				og := cpy
				cpy = func(_, _ []byte) int { return 0 }
				defer func() { cpy = og }()

				return Body([]byte("body bytes"))(rq)
			},
			assert: func(t *testing.T, rq *http.Request, err error) {
				test.Error(t, err).Is(ErrCopyFailed)
				test.IsTrue(t, rq.Body == nil, "body is nil")
			},
		},
		{scenario: "Body/set successfully",
			act: func(rq *http.Request) error {
				return Body([]byte("body bytes"))(rq)
			},
			assert: func(t *testing.T, rq *http.Request, err error) {
				body, _ := io.ReadAll(rq.Body)
				defer rq.Body.Close()

				test.Error(t, err).IsNil()
				test.Value(t, rq.ContentLength, "content length").Equals(10)
				test.Bytes(t, body).Equals([]byte("body bytes"))
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
