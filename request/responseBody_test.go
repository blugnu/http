package request

import (
	"net/http"
	"testing"

	"github.com/blugnu/test"
)

func TestResponseBody(t *testing.T) {
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
				err := ResponseBodyRequired()(rq)

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, rq.Header[ResponseBodyRequiredHeader][0]).Equals("true")
			},
		},
		{scenario: "existing header/false",
			exec: func(t *testing.T) {
				// ARRANGE
				rq, _ := http.NewRequest(http.MethodGet, "", nil)
				rq.Header[ResponseBodyRequiredHeader] = []string{"false"}

				// ACT
				err := ResponseBodyRequired()(rq)

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, rq.Header[ResponseBodyRequiredHeader][0]).Equals("true")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
