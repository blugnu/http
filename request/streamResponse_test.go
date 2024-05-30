package request

import (
	"net/http"
	"testing"

	"github.com/blugnu/test"
)

func TestStreamResponse(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "no header",
			exec: func(t *testing.T) {
				// ARRANGE
				rq, _ := http.NewRequest(http.MethodGet, "", nil)

				// ACT
				StreamResponse()(rq)

				// ASSERT
				test.That(t, rq.Header[StreamResponseHeader][0]).Equals("true")
			},
		},
		{scenario: "existing header/false",
			exec: func(t *testing.T) {
				// ARRANGE
				rq, _ := http.NewRequest(http.MethodGet, "", nil)
				rq.Header[StreamResponseHeader] = []string{"false"}

				// ACT
				StreamResponse()(rq)

				// ASSERT
				test.That(t, rq.Header[StreamResponseHeader][0]).Equals("true")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
