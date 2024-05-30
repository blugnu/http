package request

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/blugnu/test"
)

func TestQuery(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "new",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{URL: &url.URL{}}

				// ACT
				err := Query(map[string]any{
					"foo": "bar",
					"g:g": nil})(rq)

				// ASSERT
				test.That(t, err).IsNil()

				// because map iteration order is not guaranteed either of the
				// possible permutations of the query string might be expected
				test.IsTrue(t,
					rq.URL.RawQuery == "foo=bar&g%3Ag" ||
						rq.URL.RawQuery == "g%3Ag&foo=bar")
			},
		},
		{scenario: "append",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{URL: &url.URL{RawQuery: "existing"}}

				// ACT
				err := Query(map[string]any{
					"foo": "bar",
					"g:g": nil})(rq)

				// ASSERT
				test.That(t, err).IsNil()

				// because map iteration order is not guaranteed either of the
				// possible permutations of the query string might be expected
				test.IsTrue(t,
					rq.URL.RawQuery == "existing&foo=bar&g%3Ag" ||
						rq.URL.RawQuery == "existing&g%3Ag&foo=bar")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
func TestQueryP(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "nil value",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{URL: &url.URL{}}

				// ACT
				err := QueryP("foo", nil)(rq)

				// ASSERT
				test.That(t, err).IsNil()
				test.That(t, rq.URL.RawQuery).Equals("foo")
			},
		},
		{scenario: "true value",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{URL: &url.URL{}}

				// ACT
				err := QueryP("foo", true)(rq)

				// ASSERT
				test.That(t, err).IsNil()
				test.That(t, rq.URL.RawQuery).Equals("foo=true")
			},
		},
		{scenario: "string value",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{URL: &url.URL{}}

				// ACT
				err := QueryP("foo", "http://google.com")(rq)

				// ASSERT
				test.That(t, err).IsNil()
				test.That(t, rq.URL.RawQuery).Equals("foo=http%3A%2F%2Fgoogle.com")
			},
		},
		{scenario: "append/nil value",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{URL: &url.URL{RawQuery: "existing"}}

				// ACT
				err := QueryP("foo", nil)(rq)

				// ASSERT
				test.That(t, err).IsNil()
				test.That(t, rq.URL.RawQuery).Equals("existing&foo")
			},
		},
		{scenario: "append/true value",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{URL: &url.URL{RawQuery: "existing"}}

				// ACT
				err := QueryP("foo", true)(rq)

				// ASSERT
				test.That(t, err).IsNil()
				test.That(t, rq.URL.RawQuery).Equals("existing&foo=true")
			},
		},
		{scenario: "append/string value",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{URL: &url.URL{RawQuery: "existing"}}

				// ACT
				err := QueryP("foo", "bar")(rq)

				// ASSERT
				test.That(t, err).IsNil()
				test.That(t, rq.URL.RawQuery).Equals("existing&foo=bar")
			},
		},
		{scenario: "url encoding",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{URL: &url.URL{}}

				// ACT
				err := QueryP("\"a map\"", "key=value")(rq)

				// ASSERT
				test.That(t, err).IsNil()
				test.That(t, rq.URL.RawQuery).Equals("%22a+map%22=key%3Dvalue")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}

func TestRawQuery(t *testing.T) {
	// ARRANGE
	rq := &http.Request{URL: &url.URL{RawQuery: "will be over-written"}}

	// ACT
	err := RawQuery("foo=bar")(rq)

	// ASSERT
	test.That(t, err).IsNil()
	test.That(t, rq.URL.RawQuery).Equals("foo=bar")
}
