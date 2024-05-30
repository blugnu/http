package request

import (
	"net/http"
	"testing"

	"github.com/blugnu/test"
)

func TestMaxRetries(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(*testing.T)
	}{
		{scenario: "no header",
			exec: func(t *testing.T) {
				// ARRANGE
				rq, _ := http.NewRequest(http.MethodGet, "", nil)

				// ACT
				err := MaxRetries(3)(rq)

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, rq.Header[MaxRetriesHeader][0]).Equals("3")
			},
		},
		{scenario: "existing header",
			exec: func(t *testing.T) {
				// ARRANGE
				rq, _ := http.NewRequest(http.MethodGet, "", nil)
				rq.Header[MaxRetriesHeader] = []string{"10"}

				// ACT
				err := MaxRetries(3)(rq)

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, rq.Header[MaxRetriesHeader][0]).Equals("3")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
