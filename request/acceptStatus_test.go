package request

import (
	"net/http"
	"testing"

	"github.com/blugnu/test"
)

func TestAcceptStatus(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(*testing.T)
	}{
		{scenario: "no header/add status",
			exec: func(t *testing.T) {
				// ARRANGE
				rq, _ := http.NewRequest(http.MethodGet, "", nil)

				// ACT
				err := AcceptStatus(http.StatusNotFound)(rq)

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, rq.Header[AcceptStatusHeader][0]).Equals("[200,404]")
			},
		},
		{scenario: "existing header/add status",
			exec: func(t *testing.T) {
				// ARRANGE
				rq, _ := http.NewRequest(http.MethodGet, "", nil)
				rq.Header[AcceptStatusHeader] = []string{"[200,401]"}

				// ACT
				err := AcceptStatus(http.StatusNotFound)(rq)

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, rq.Header[AcceptStatusHeader][0]).Equals("[200,401,404]")
			},
		},
		{scenario: "existing header/malformed",
			exec: func(t *testing.T) {
				// ARRANGE
				rq, _ := http.NewRequest(http.MethodGet, "", nil)
				rq.Header[AcceptStatusHeader] = []string{"this is not valid"}

				// ACT
				err := AcceptStatus(http.StatusNotFound)(rq)

				// ASSERT
				test.Error(t, err).Is(ErrInvalidJSON)
				test.That(t, rq.Header[AcceptStatusHeader][0]).Equals("this is not valid")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
