package http

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/blugnu/test"
)

func TestMockRequest(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(*testing.T)
	}{
		// checkExpectations tests
		{scenario: "checkExpectations/not expected/no actual",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &MockRequest{}

				// ACT
				result := rq.checkExpectations()

				// ASSERT
				test.Strings(t, result).IsEmpty()
			},
		},
		{scenario: "checkExpectations/not expected/actual",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "http://hostname/path", nil)
				rq := &MockRequest{
					actual: a,
				}

				// ACT
				result := rq.checkExpectations()

				// ASSERT
				test.Strings(t, result).Equals([]string{
					"  got: GET http://hostname/path",
				})
			},
		},
		{scenario: "checkExpectations/expected/no actual",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &MockRequest{isExpected: true}

				// ACT
				result := rq.checkExpectations()

				// ASSERT
				test.Strings(t, result).Equals([]string{
					"  got: <no request>",
				})
			},
		},

		// checkMethodExpectation tests
		{scenario: "checkMethodExpectation/expect any method",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "", nil)
				rq := MockRequest{isExpected: true, actual: a}

				// ACT
				result := rq.checkMethodExpectation()

				// ASSERT
				test.That(t, result).IsNil()
			},
		},
		{scenario: "checkMethodExpectation/expect GET/got GET",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "", nil)
				m := http.MethodGet
				rq := MockRequest{isExpected: true, method: &m, actual: a}

				// ACT
				result := rq.checkMethodExpectation()

				// ASSERT
				test.That(t, result).IsNil()
			},
		},
		{scenario: "checkMethodExpectation/expect GET/got POST",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "", nil)
				m := http.MethodPost
				rq := MockRequest{isExpected: true, method: &m, actual: a}

				// ACT
				result := rq.checkMethodExpectation()

				// ASSERT
				test.That(t, result).Equals([]string{
					"expected method: POST",
					"   got         : GET",
				})
			},
		},
		// checkURLExpectation tests
		{scenario: "checkURLExpectation/url not set",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "http://hostname/path", nil)
				rq := MockRequest{isExpected: true, actual: a}

				// ACT
				result := rq.checkURLExpectation()

				// ASSERT
				test.That(t, result).Equals([]string{
					"expected url: <not specified>",
					"   got      : http://hostname/path",
				})
			},
		},
		{scenario: "checkURLExpectation/got unexpected",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "http://otherhost/path", nil)
				rq := MockRequest{isExpected: true, url: "http://hostname/path", actual: a}

				// ACT
				result := rq.checkURLExpectation()

				// ASSERT
				test.That(t, result).Equals([]string{
					"expected url: http://hostname/path",
					"   got      : http://otherhost/path",
				})
			},
		},

		// checkHeadersExpectation tests
		{scenario: "checkHeadersExpectation/any value/submitted",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "", nil)
				a.Header["key"] = []string{"any"}
				rq := MockRequest{isExpected: true, actual: a, headers: map[string]*string{"key": nil}}

				// ACT
				result := rq.checkHeadersExpectation()

				// ASSERT
				test.That(t, result).IsNil()
			},
		},
		{scenario: "checkHeadersExpectation/submitted with expected value",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "", nil)
				a.Header["key"] = []string{"value"}
				v := "value"
				rq := MockRequest{isExpected: true, actual: a, headers: map[string]*string{"key": &v}}

				// ACT
				result := rq.checkHeadersExpectation()

				// ASSERT
				test.That(t, result).IsNil()
			},
		},
		{scenario: "checkHeadersExpectation/any value/not present",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "", nil)
				a.Header["other"] = []string{"value"}
				rq := MockRequest{isExpected: true, actual: a, headers: map[string]*string{"key": nil}}

				// ACT
				result := rq.checkHeadersExpectation()

				// ASSERT
				test.That(t, result).Equals([]string{
					"header not set: key",
					"           got: [",
					"             other: value",
					"           ]",
				})
			},
		},
		{scenario: "checkHeadersExpectation/specific value/not present",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "", nil)
				a.Header["other"] = []string{"value"}
				v := "value"
				rq := MockRequest{isExpected: true, actual: a, headers: map[string]*string{"key": &v}}

				// ACT
				result := rq.checkHeadersExpectation()

				// ASSERT
				test.That(t, result).Equals([]string{
					"header not set: key: value",
					"           got: [",
					"             other: value",
					"           ]",
				})
			},
		},
		{scenario: "checkHeadersExpectation/present with wrong value",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "", nil)
				a.Header["key"] = []string{"other value"}
				v := "value"
				rq := MockRequest{isExpected: true, actual: a, headers: map[string]*string{"key": &v}}

				// ACT
				result := rq.checkHeadersExpectation()

				// ASSERT
				test.That(t, result).Equals([]string{
					"expected header: key: value",
					"   got         : key: other value",
				})
			},
		},

		// checkBodyExpectation tests
		{scenario: "checkBodyExpectation/any body/with body",
			exec: func(t *testing.T) {
				// ARRANGE
				b := bytes.NewReader([]byte("body"))
				a, _ := http.NewRequest(http.MethodGet, "", b)
				rq := MockRequest{isExpected: true, actual: a, body: nil}

				// ACT
				result := rq.checkBodyExpectation()

				// ASSERT
				test.That(t, result).IsNil()
			},
		},
		{scenario: "checkBodyExpectation/any body/with no body",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "", nil)
				rq := MockRequest{isExpected: true, actual: a, body: nil}

				// ACT
				result := rq.checkBodyExpectation()

				// ASSERT
				test.That(t, result).IsNil()
			},
		},
		{scenario: "checkBodyExpectation/no body/with no body",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "", http.NoBody)
				b := []byte{}
				rq := MockRequest{isExpected: true, actual: a, body: &b}

				// ACT
				result := rq.checkBodyExpectation()

				// ASSERT
				test.That(t, result).IsNil()
			},
		},
		{scenario: "checkBodyExpectation/no body/with body",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "", bytes.NewReader([]byte("body")))
				b := []byte{}
				rq := MockRequest{isExpected: true, actual: a, body: &b}

				// ACT
				result := rq.checkBodyExpectation()

				// ASSERT
				test.That(t, result).Equals([]string{
					"expected: <no body>",
					"   got  : 4 bytes",
				})
			},
		},
		{scenario: "checkBodyExpectation/body/with no body",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "", http.NoBody)
				b := []byte("body")
				rq := MockRequest{isExpected: true, actual: a, body: &b}

				// ACT
				result := rq.checkBodyExpectation()

				// ASSERT
				test.That(t, result).Equals([]string{
					"expected: 4 bytes",
					"   got  : <no body>",
				})
			},
		},
		{scenario: "checkBodyExpectation/body/with different body",
			exec: func(t *testing.T) {
				// ARRANGE
				a, _ := http.NewRequest(http.MethodGet, "", bytes.NewReader([]byte("other")))
				b := []byte("body")
				rq := MockRequest{isExpected: true, actual: a, body: &b}

				// ACT
				result := rq.checkBodyExpectation()

				// ASSERT
				test.That(t, result).Equals([]string{
					"request body differs from expected",
					"   got   :_________",
					"         |other",
					"   wanted:_________",
					"         |body",
				})
			},
		},

		// String tests
		{scenario: "String/no method/no url",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := MockRequest{}

				// ACT
				result := rq.String()

				// ASSERT
				test.That(t, result).Equals("<ANY> <any://hostname/and/path>")
			},
		},
		{scenario: "String/method/no url",
			exec: func(t *testing.T) {
				// ARRANGE
				m := http.MethodGet
				rq := MockRequest{method: &m}

				// ACT
				result := rq.String()

				// ASSERT
				test.That(t, result).Equals("GET <any://hostname/and/path>")
			},
		},
		{scenario: "String/no method, url set",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := MockRequest{url: "http://hostname/path"}

				// ACT
				result := rq.String()

				// ASSERT
				test.That(t, result).Equals("<ANY> http://hostname/path")
			},
		},
		{scenario: "String/with method and url",
			exec: func(t *testing.T) {
				// ARRANGE
				m := http.MethodGet
				rq := MockRequest{method: &m, url: "http://hostname/path"}

				// ACT
				result := rq.String()

				// ASSERT
				test.That(t, result).Equals("GET http://hostname/path")
			},
		},

		// Will... tests
		{scenario: "WillNotBeCalled",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &MockRequest{isExpected: true}

				// ACT
				rq.WillNotBeCalled()

				// ASSERT
				test.Bool(t, rq.isExpected).IsFalse()
			},
		},
		{scenario: "WillRespond",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &MockRequest{isExpected: true}

				// ACT
				rq.WillRespond()

				// ASSERT
				test.That(t, rq.Response).Equals(&mockResponse{
					headers: map[string]string{},
				})
			},
		},
		{scenario: "WillReturnError",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &MockRequest{isExpected: true}
				rqerr := errors.New("request error")

				// ACT
				rq.WillReturnError(rqerr)

				// ASSERT
				test.That(t, rq.Response.Err).Equals(rqerr)
			},
		},
		{scenario: "WithBody",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &MockRequest{isExpected: true}

				// ACT
				rq.WithBody([]byte("foo"))

				// ASSERT
				test.That(t, *rq.body).Equals([]byte("foo"))
			},
		},
		{scenario: "WithHeader/any value",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &MockRequest{
					headers:    map[string]*string{},
					isExpected: true,
				}

				// ACT
				rq.WithHeader("content-type")

				// ASSERT
				test.That(t, rq).Equals(&MockRequest{
					headers:    map[string]*string{"Content-Type": nil},
					isExpected: true,
				})
			},
		},
		{scenario: "WithHeader/specific value",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &MockRequest{
					headers:    map[string]*string{},
					isExpected: true,
				}
				ct := "application/json"

				// ACT
				rq.WithHeader("content-type", ct)

				// ASSERT
				test.That(t, rq).Equals(&MockRequest{
					headers:    map[string]*string{"Content-Type": &ct},
					isExpected: true,
				})
			},
		},
		{scenario: "WithNonCanonicalHeader/any value",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &MockRequest{
					headers:    map[string]*string{},
					isExpected: true,
				}

				// ACT
				rq.WithNonCanonicalHeader("sessionid")

				// ASSERT
				test.That(t, rq).Equals(&MockRequest{
					headers:    map[string]*string{"sessionid": nil},
					isExpected: true,
				})
			},
		},
		{scenario: "WithNonCanonicalHeader/specific value",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &MockRequest{
					headers:    map[string]*string{},
					isExpected: true,
				}
				id := "d29f0e6e-9100-4238-a2e6-4347116a8177"

				// ACT
				rq.WithNonCanonicalHeader("sessionid", id)

				// ASSERT
				test.That(t, rq).Equals(&MockRequest{
					headers:    map[string]*string{"sessionid": &id},
					isExpected: true,
				})
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
