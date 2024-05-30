package http

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/blugnu/test"
)

func TestNewMockClient(t *testing.T) {
	// ARRANGE
	defer test.ExpectPanic(nil).Assert(t) // the nil wrapper func should not cause a panic
	wrappersAreApplied := false

	wrappers := []func(c interface {
		Do(*http.Request) (*http.Response, error)
	}) interface {
		Do(*http.Request) (*http.Response, error)
	}{
		func(c interface {
			Do(*http.Request) (*http.Response, error)
		}) interface {
			Do(*http.Request) (*http.Response, error)
		} {
			wrappersAreApplied = true
			return c.(ClientInterface)
		},
		nil,
	}

	// ACT
	c, m := NewMockClient("foo", wrappers...)

	// ASSERT
	if c, ok := test.IsType[client](t, c); ok {
		test.That(t, c.name).Equals("foo")
		test.That(t, c.url).Equals("mock://hostname")
	}
	if m, ok := test.IsType[*mockClient](t, m); ok {
		test.That(t, m.name).Equals("foo")
		test.That(t, m.hostname).Equals("mock://hostname")
	}
	test.IsTrue(t, wrappersAreApplied)
}

func TestMockClient(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(*testing.T)
	}{
		// defaultResponse tests
		{scenario: "defaultResponse/response configured",
			exec: func(t *testing.T) {
				// ARRANGE
				rsperr := errors.New("response error")
				sc := 400
				rq := MockRequest{
					Response: &mockResponse{
						headers: map[string]string{
							"header": "value",
						},
						body:       []byte("body"),
						statusCode: &sc,
						Err:        rsperr,
					},
				}
				c := mockClient{}

				// ACT
				result, err := c.defaultResponse(&rq)

				// ASSERT
				body, _ := io.ReadAll(result.Body)

				test.Error(t, err).Is(rsperr)
				test.That(t, body).Equals([]byte("body"))
				test.That(t, result.Header, "headers").Equals(http.Header{"header": []string{"value"}})
				test.That(t, result.StatusCode).Equals(400)
			},
		},
		{scenario: "defaultResponse/error writing response body",
			exec: func(t *testing.T) {
				// ARRANGE
				rwerr := errors.New("response writer error")
				c := mockClient{}

				og := writeBody
				defer func() { writeBody = og }()
				writeBody = func(rw http.ResponseWriter, d []byte) (int, error) { return 0, rwerr }

				// ACT
				result, err := c.defaultResponse(&MockRequest{
					Response: &mockResponse{
						body: []byte("non-empty"),
					},
				})

				// ASSERT
				test.That(t, result).IsNil()
				test.Error(t, err).Is(rwerr)
			},
		},
		{scenario: "defaultResponse/default",
			exec: func(t *testing.T) {
				// ARRANGE
				c := mockClient{}

				// ACT
				result, err := c.defaultResponse(&MockRequest{})

				// ASSERT
				body, _ := io.ReadAll(result.Body)

				test.Error(t, err).IsNil()
				test.That(t, body).Equals([]byte{})
				test.That(t, result.Header, "headers").Equals(http.Header{})
				test.That(t, result.StatusCode).Equals(http.StatusOK)
			},
		},

		// Do tests
		{scenario: "Do/no requests expected",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &mockClient{next: noExpectedRequests}
				rq, _ := http.NewRequest(http.MethodGet, "http://hostname/path", nil)

				// ACT
				response, err := client.Do(rq)

				// ASSERT
				test.Error(t, err).Is(ErrUnexpectedRequest)
				test.That(t, response).IsNil()

				test.Slice(t, client.unexpected).Equals([]*http.Request{
					{
						Method: http.MethodGet,
						URL:    &url.URL{Scheme: "http", Host: "hostname", Path: "path"},
					}},
					func(got, wanted *http.Request) bool {
						return got.Method == wanted.Method &&
							got.URL.String() == wanted.URL.String()
					})
			},
		},
		{scenario: "Do/request explictly not expected",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &mockClient{
					expectations: []*MockRequest{{isExpected: false}},
				}
				rq, _ := http.NewRequest(http.MethodGet, "http://hostname/path", nil)

				// ACT
				response, err := client.Do(rq)

				// ASSERT
				test.Error(t, err).Is(ErrUnexpectedRequest)
				test.That(t, response).IsNil()
			},
		},
		{scenario: "Do/default handling",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &mockClient{
					expectations: []*MockRequest{{isExpected: true}},
				}
				rq, _ := http.NewRequest(http.MethodGet, "http://hostname/path", nil)

				// ACT
				response, err := client.Do(rq)

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, response.StatusCode, "status code").Equals(http.StatusOK)
				test.That(t, response.Body, "body").Equals(http.NoBody)
			},
		},

		// ExpectationsWereMet tests
		{scenario: "ExpectationsWereMet/no requests expected/no requests made",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &mockClient{next: noExpectedRequests}

				// ACT
				err := client.ExpectationsWereMet()

				// ASSERT
				test.Error(t, err).IsNil()
			},
		},
		{scenario: "ExpectationsWereMet/no requests expected/requests made",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &mockClient{
					name:       "foo",
					next:       noExpectedRequests,
					unexpected: []*http.Request{{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "hostname", Path: "path"}}},
				}

				// ACT
				test := test.Helper(t, func(t *testing.T) {
					test.Error(t, client.ExpectationsWereMet()).IsNil()
				})

				// ASSERT
				test.Report.Contains([]string{
					"unexpected error: foo: expectations not met",
					"request #1: unexpected: GET http://hostname/path",
				})
			},
		},
		{scenario: "ExpectationsWereMet/one expected request/one unexpected",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &mockClient{
					name:         "foo",
					next:         0,
					expectations: []*MockRequest{{}},
					unexpected:   []*http.Request{{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "hostname", Path: "path"}}},
				}

				// ACT
				test := test.Helper(t, func(t *testing.T) {
					test.Error(t, client.ExpectationsWereMet()).IsNil()
				})

				// ASSERT
				test.Report.Contains([]string{
					"unexpected error: foo: expectations not met",
					"request #2: unexpected: GET http://hostname/path",
				})
			},
		},
		{scenario: "ExpectationsWereMet/expected request is made",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &mockClient{
					expectations: []*MockRequest{
						{
							isExpected: true,
							url:        "http://hostname/path",
							actual: &http.Request{
								Method: http.MethodGet,
								URL:    &url.URL{Scheme: "http", Host: "hostname", Path: "path"},
							},
						},
					},
				}

				// ACT
				err := client.ExpectationsWereMet()

				// ASSERT
				test.Error(t, err).IsNil()
			},
		},
		{scenario: "ExpectationsWereMet/unexpected method",
			exec: func(t *testing.T) {
				// ARRANGE
				m := http.MethodPost
				client := &mockClient{
					name: "foo",
					expectations: []*MockRequest{
						{
							isExpected: true,
							method:     &m,
							url:        "http://hostname/path",
							actual: &http.Request{
								Method: http.MethodGet,
								URL:    &url.URL{Scheme: "http", Host: "hostname", Path: "path"},
							},
						},
					},
				}

				// ACT
				test := test.Helper(t, func(t *testing.T) {
					test.Error(t, client.ExpectationsWereMet()).IsNil()
				})

				// ASSERT
				test.Report.Contains([]string{
					"unexpected error: foo: expectations not met",
					"request #1: expecting: POST http://hostname/path",
					"   expected method: POST",
					"      got         : GET",
				})
			},
		},
		{scenario: "ExpectationsWereMet/unexpected url",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &mockClient{
					name: "foo",
					expectations: []*MockRequest{
						{
							isExpected: true,
							url:        "https://other/path",
							actual: &http.Request{
								Method: http.MethodGet,
								URL:    &url.URL{Scheme: "http", Host: "hostname", Path: "path"},
							},
						},
					},
				}

				// ACT
				test := test.Helper(t, func(t *testing.T) {
					test.Error(t, client.ExpectationsWereMet()).IsNil()
				})

				// ASSERT
				test.Report.Contains([]string{
					"unexpected error: foo: expectations not met",
					"request #1: expecting: <ANY METHOD> https://other/path",
					"   expected url: https://other/path",
					"      got      : http://hostname/path",
				})
			},
		},

		// Expect tests
		{scenario: "Expect/initialises expected request",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &mockClient{
					hostname: "http://hostname",
					next:     noExpectedRequests,
				}

				// ACT
				result := client.Expect(http.MethodGet, "path")

				// ASSERT
				if result, ok := test.IsType[*MockRequest](t, result); ok {
					gm := http.MethodGet
					want := &MockRequest{
						index:      0,
						client:     client,
						method:     &gm,
						url:        "http://hostname/path",
						headers:    map[string]*string{},
						isExpected: true,
					}
					test.That(t, result).Equals(want)
				}
				test.That(t, client.next).Equals(0)
			},
		},
		{scenario: "Expect/when requests already made",
			exec: func(t *testing.T) {
				// ARRANGE
				defer test.ExpectPanic(ErrCannotChangeExpectations).Assert(t)
				client := &mockClient{next: 1}

				// ACT
				client.Expect("any", "any")
			},
		},
		{scenario: "Expect/when url is invalid",
			exec: func(t *testing.T) {
				// any path appended to a url will be escaped; it is impossible
				// to force an invalid url using a path, therefore, to exercise
				// validation of the request url we must use a mock client with
				// an invalid url

				// ARRANGE
				defer test.ExpectPanic(ErrInvalidURL).Assert(t)
				client := &mockClient{hostname: "\n"}

				// ACT
				client.Expect("any method", "any path")
			},
		},
		{scenario: "ExpectDelete",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &mockClient{next: noExpectedRequests}

				// ACT
				result := client.ExpectDelete(http.MethodGet)

				// ASSERT
				if result, ok := test.IsType[*MockRequest](t, result); ok {
					test.That(t, *result.method).Equals(http.MethodDelete)
				}
				test.That(t, client.next).Equals(0)
			},
		},
		{scenario: "ExpectGet",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &mockClient{next: noExpectedRequests}

				// ACT
				result := client.ExpectGet(http.MethodGet)

				// ASSERT
				if result, ok := test.IsType[*MockRequest](t, result); ok {
					test.That(t, *result.method).Equals(http.MethodGet)
				}
				test.That(t, client.next).Equals(0)
			},
		},
		{scenario: "ExpectPatch",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &mockClient{next: noExpectedRequests}

				// ACT
				result := client.ExpectPatch(http.MethodGet)

				// ASSERT
				if result, ok := test.IsType[*MockRequest](t, result); ok {
					test.That(t, *result.method).Equals(http.MethodPatch)
				}
				test.That(t, client.next).Equals(0)
			},
		},
		{scenario: "ExpectPost",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &mockClient{next: noExpectedRequests}

				// ACT
				result := client.ExpectPost(http.MethodGet)

				// ASSERT
				if result, ok := test.IsType[*MockRequest](t, result); ok {
					test.That(t, *result.method).Equals(http.MethodPost)
				}
				test.That(t, client.next).Equals(0)
			},
		},
		{scenario: "ExpectPut",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &mockClient{next: noExpectedRequests}

				// ACT
				result := client.ExpectPut(http.MethodGet)

				// ASSERT
				if result, ok := test.IsType[*MockRequest](t, result); ok {
					test.That(t, *result.method).Equals(http.MethodPut)
				}
				test.That(t, client.next).Equals(0)
			},
		},

		// Reset tests
		{scenario: "Reset",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &mockClient{
					next:         1,
					expectations: []*MockRequest{{}},
					unexpected:   []*http.Request{{}},
				}

				// ACT
				client.Reset()

				// ASSERT
				test.That(t, client.next).Equals(noExpectedRequests)
				test.Slice(t, client.expectations).IsEmpty()
				test.Slice(t, client.unexpected).IsEmpty()
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
