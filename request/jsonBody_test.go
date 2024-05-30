package request

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/blugnu/test"
)

type unmarshallable struct{}

func (unmarshallable) MarshalJSON() ([]byte, error) {
	return nil, errors.New("cannot be marshalled")
}

func TestJSONBody(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		act      func(*http.Request) error
		assert   func(*testing.T, *http.Request, error)
	}{
		// Body tests
		{scenario: "JSONBody/marshalling error",
			act: func(rq *http.Request) error {
				return JSONBody(unmarshallable{})(rq)
			},
			assert: func(t *testing.T, rq *http.Request, err error) {
				test.Error(t, err).Is(ErrMarshallingJSON)
				test.IsTrue(t, rq.Body == nil, "body is nil")
			},
		},
		{scenario: "JSONBody/int",
			act: func(rq *http.Request) error {
				return JSONBody(42)(rq)
			},
			assert: func(t *testing.T, rq *http.Request, err error) {
				body, _ := io.ReadAll(rq.Body)
				defer rq.Body.Close()

				test.Error(t, err).IsNil()
				test.Value(t, rq.Header["Content-Type"][0], "content type").Equals("application/json")
				test.Value(t, rq.ContentLength, "content length").Equals(2)
				test.Bytes(t, body).Equals([]byte("42"))
			},
		},
		{scenario: "JSONBody/string",
			act: func(rq *http.Request) error {
				return JSONBody("some string")(rq)
			},
			assert: func(t *testing.T, rq *http.Request, err error) {
				body, _ := io.ReadAll(rq.Body)
				defer rq.Body.Close()

				test.Error(t, err).IsNil()
				test.Value(t, rq.Header["Content-Type"][0], "content type").Equals("application/json")
				test.Value(t, rq.ContentLength, "content length").Equals(13)
				test.Bytes(t, body).Equals([]byte(`"some string"`))
			},
		},
		{scenario: "JSONBody/struct",
			act: func(rq *http.Request) error {
				return JSONBody(struct {
					Name string
					Age  int
				}{"Jane Smith", 32})(rq)
			},
			assert: func(t *testing.T, rq *http.Request, err error) {
				body, _ := io.ReadAll(rq.Body)
				defer rq.Body.Close()

				test.Error(t, err).IsNil()
				test.Value(t, rq.Header["Content-Type"][0], "content type").Equals("application/json")
				test.Value(t, rq.ContentLength, "content length").Equals(30)
				test.Bytes(t, body).Equals([]byte(`{"Name":"Jane Smith","Age":32}`))
			},
		},
		{scenario: "JSONBody/slice",
			act: func(rq *http.Request) error {
				return JSONBody([]int{1, 2, 3})(rq)
			},
			assert: func(t *testing.T, rq *http.Request, err error) {
				body, _ := io.ReadAll(rq.Body)
				defer rq.Body.Close()

				test.Error(t, err).IsNil()
				test.Value(t, rq.Header["Content-Type"][0], "content type").Equals("application/json")
				test.Value(t, rq.ContentLength, "content length").Equals(7)
				test.Bytes(t, body).Equals([]byte(`[1,2,3]`))
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
