package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
)

const (
	firstExpectedRequest = 0
	noExpectedRequests   = -1
)

var (
	writeBody = func(rw http.ResponseWriter, d []byte) (int, error) { return rw.Write(d) }
)

// MockClient is an interface that described the methods provided
// for mocking expectations on a client
type MockClient interface {
	Expect(method string, path string) *MockRequest
	ExpectDelete(path string) *MockRequest
	ExpectGet(path string) *MockRequest
	ExpectPatch(path string) *MockRequest
	ExpectPost(path string) *MockRequest
	ExpectPut(path string) *MockRequest
	ExpectationsWereMet() error
	Reset()
}

// mockClient implements the HttpClient interface, providing additional
// methods for configuring request and response expectations and
// verifying that those expectations have been met.
type mockClient struct {
	name         string
	hostname     string
	expectations []*MockRequest
	unexpected   []*http.Request
	next         int
}

// NewMockClient returns a new http.HttpClient to be used for making
// requests and an http.MockClient on which expected requests, and corresponding
// responses, may be configured.
//
// # params
//
//	name          // used to identify the mock client in test failure reports and errors
//	wrap          // optional function(s) to wrap the client with some other client
//	              // implementation, if required; nil functions are ignored
//
// # returns
//
//	HttpClient    // used to make requests; this should be injected into code under test
//	MockClient    // used to configure expected requests and provide details of responses
//	              // to be mocked for each request
//
// Note the use of an anonymous interface in the exported function signature.
// This avoids creating coupling modules thru a shared reference to an interface
// type.  In Go (currently at least) interfaces are fungible but interface types are not.
func NewMockClient(name string, wrap ...func(c interface {
	Do(*http.Request) (*http.Response, error)
}) interface {
	Do(*http.Request) (*http.Response, error)
}) (HttpClient, MockClient) {
	def := &mockClient{
		name:     name,
		hostname: "mock://hostname",
		next:     noExpectedRequests,
	}

	// internally we can use an interface type for brevity/clarity as this
	// is not part of the contract exported by the function
	var mock ClientInterface = def
	for _, wrap := range wrap {
		if wrap == nil {
			continue
		}
		mock = wrap(mock)
	}

	c, _ := NewClient(def.name,
		URL(def.hostname),
		Using(mock),
	)

	return c.(client), def
}

// defaultResponse provides the response configured as expected from the supplied
// expected request.  If no respond properties are configured, a simple OK response
// is returned.
func (mock *mockClient) defaultResponse(
	expected *MockRequest,
) (response *http.Response, err error) {
	var bodyerr error
	rec := httptest.NewRecorder()
	func(rw http.ResponseWriter, _ *http.Request) {
		if expected.Response == nil {
			rw.WriteHeader(http.StatusOK)
			return
		}

		if expected.Response.headers != nil {
			for k, v := range expected.Response.headers {
				rw.Header()[k] = []string{v}
			}
		}

		if expected.Response.statusCode != nil {
			rw.WriteHeader(*expected.Response.statusCode)
		}

		if len(expected.Response.body) > 0 {
			_, bodyerr = writeBody(rw, expected.Response.body)
		}

		err = expected.Response.Err
	}(rec, nil)

	// if there was an error writing the response body then the response is
	// unusable and we should return that error
	if bodyerr != nil {
		return nil, bodyerr
	}

	response = rec.Result()

	// if there is no configured response expectation or the expected
	// response has no body or an empty body then the response Body will be
	// http.NoBody
	if expected.Response == nil || len(expected.Response.body) == 0 {
		response.Body = http.NoBody
	}

	return
}

// Do implements the ClientInterface interface to perform any http request.
// The mock client takes the request to be performed, checks it against
// the next expected request and constructs any configured expected
// response either by passing it to a configured request handler or
// constructing a default response.
func (mock *mockClient) Do(rq *http.Request) (*http.Response, error) {
	if mock.next != noExpectedRequests && mock.next < len(mock.expectations) {
		expected := mock.expectations[mock.next]
		expected.actual = rq
		mock.next++

		switch {
		case !expected.isExpected:
			// NO-OP - the request will be recorded as unexpected

		default:
			return mock.defaultResponse(expected)
		}
	}

	mock.unexpected = append(mock.unexpected, rq)
	return nil, ErrUnexpectedRequest
}

// ExpectationsWereMet checks the expected requests against actual requests made
// and returns an error if any expectations were not met.
func (mock mockClient) ExpectationsWereMet() error {
	errs := []error{}

	for _, rq := range mock.expectations {
		rpt := rq.checkExpectations()
		if len(rpt) > 0 {
			m := "<ANY METHOD>"
			if rq.method != nil {
				m = *rq.method
			}
			errs = append(errs, fmt.Errorf("request #%d: expecting: %s %s", rq.index+1, m, rq.url))
			for _, s := range rpt {
				errs = append(errs, fmt.Errorf("   %s", s))
			}
		}
	}

	for ix, rq := range mock.unexpected {
		errs = append(errs, fmt.Errorf("request #%d: unexpected: %s %s",
			len(mock.expectations)+ix+1,
			rq.Method,
			rq.URL.String(),
		))
	}

	if len(errs) > 0 {
		return MockExpectationsError{mock.name, errs}
	}

	return nil
}

// Expect registers an expected request of an identified http method. The expected
// request is returned which may be used to configure additional properties of the
// expected request.
//
// If no additional properties are configured, the expectation will be satisfied by any
// request using the expected method, regardless of URL, body content or headers etc.
//
// This method will panic if called after a mock client has already received at least
// one request.
func (mock *mockClient) Expect(method string, path string) *MockRequest {
	if mock.next > 0 {
		msg := "requests have already been made"
		panic(fmt.Errorf("%s: %w: %s", mock.name, ErrCannotChangeExpectations, msg))
	}

	fqu, err := url.JoinPath(mock.hostname, path)
	if err != nil {
		msg := fmt.Sprintf("client url (%s) and/or request path (%s) are invalid",
			mock.hostname,
			path,
		)
		panic(fmt.Errorf("%w: %s: %w", ErrInvalidURL, msg, err))
	}

	rq := &MockRequest{
		index:      len(mock.expectations),
		method:     &method,
		url:        fqu,
		client:     mock,
		headers:    map[string]*string{},
		isExpected: true,
	}
	mock.expectations = append(mock.expectations, rq)

	if len(mock.expectations) == 1 {
		mock.next = 0
	}

	return rq
}

// ExpectDelete is a convenience method that returns a new expectation
// of a request, made using the DELETE method, with a specified url
// (appended to the base url as configured in the client).
func (mock *mockClient) ExpectDelete(path string) *MockRequest {
	return mock.Expect(http.MethodDelete, path)
}

// ExpectGet is a convenience method that returns a new expectation
// of a request, made using the GET method, with a specified url
// (appended to the base url as configured in the client).
func (mock *mockClient) ExpectGet(path string) *MockRequest {
	return mock.Expect(http.MethodGet, path)
}

// ExpectPatch is a convenience method that returns a new expectation
// of a request, made using the PATCH method, with a specified url
// (appended to the base url as configured in the client).
func (mock *mockClient) ExpectPatch(path string) *MockRequest {
	return mock.Expect(http.MethodPatch, path)
}

// ExpectPost is a convenience method that returns a new expectation
// of a request, made using the POST method, with a specified url
// (appended to the base url as configured in the client).
func (mock *mockClient) ExpectPost(path string) *MockRequest {
	return mock.Expect(http.MethodPost, path)
}

// ExpectPut is a convenience method that returns a new expectation
// of a request, made using the PUT method, with a specified url
// (appended to the base url as configured in the client).
func (mock *mockClient) ExpectPut(path string) *MockRequest {
	return mock.Expect(http.MethodPut, path)
}

// Reset clears all expectations in a mock client and prepares it to be
// configured with a new set of request expectations.
func (mock *mockClient) Reset() {
	mock.expectations = []*MockRequest{}
	mock.unexpected = []*http.Request{}
	mock.next = noExpectedRequests
}
