package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
)

// MockRequest holds details of a request expected by a MockClient
type MockRequest struct {
	// index of the request in the associated client
	index int

	// reference to the client expected to make the request, from which the base url
	// for each request is obtained
	client *mockClient

	// expected method (optional; any method is acceptable if nil, otherwise the
	// methods must match)
	method *string

	// expected body (optional; if nil, any body is acceptable, otherwise the
	// bodies must match)
	body *[]byte

	// expected url (required; the url must match exactly including any query parameters)
	url string

	// expected headers (optional; a key with a nil value indicates a header which
	// must be present regardless of value; a key with a non-nil value indicates
	// a header that must have a specific value)
	headers map[string]*string

	// records the actual request made
	actual *http.Request

	// indicates whether the request is expected or not
	isExpected bool

	// configuration of the response to be mocked in response to the request
	Response *mockResponse
}

// analyse performs expectation analysis for a request and returns a
// report identifying any unmet expectations.  If all expectations were
// met nil is returned.
func (rq *MockRequest) checkExpectations() []string {
	result := []string{}
	switch {
	case !rq.isExpected:
		if rq.actual == nil {
			return nil
		}
		result = append(result, fmt.Sprintf("  got: %s %s", rq.actual.Method, rq.actual.URL.String()))

	case rq.actual == nil:
		result = append(result, "  got: <no request>")

	default:
		result = append(result, rq.checkMethodExpectation()...)
		result = append(result, rq.checkURLExpectation()...)
		result = append(result, rq.checkHeadersExpectation()...)
		result = append(result, rq.checkBodyExpectation()...)
	}
	return result
}

// checkMethod returns a report describing any exception if the method
// expected to be used by a request was not the method used by the
// corresponding actual request
func (rq *MockRequest) checkMethodExpectation() []string {
	if rq.method != nil && *rq.method != rq.actual.Method {
		return ([]string{
			fmt.Sprintf("expected method: %s", *rq.method),
			fmt.Sprintf("   got         : %s", rq.actual.Method),
		})
	}
	return nil
}

// checkURL returns a report describing any exception if the URL
// expected to be used by a request was not the URL used by the
// corresponding actual request
func (rq *MockRequest) checkURLExpectation() []string {
	u := rq.url
	if u == "" {
		u = "<not specified>"
	}
	if rq.url != rq.actual.URL.String() {
		return []string{
			fmt.Sprintf("expected url: %s", u),
			fmt.Sprintf("   got      : %s", rq.actual.URL.String()),
		}
	}
	return nil
}

// checkHeaders returns a report describing any exception if the headers
// expected to be submitted with a request were not submitted with the
// corresponding actual request
func (rq *MockRequest) checkHeadersExpectation() (rpt []string) {
	for k, v := range rq.headers {
		avs := ""
		present := false
		if av, ok := rq.actual.Header[k]; ok {
			present = true
			avs = av[0]
		}

		switch {
		case !present && v == nil:
			rpt = append(rpt, fmt.Sprintf("header not set: %s", k), "           got: [")
			for k, av := range rq.actual.Header {
				rpt = append(rpt, fmt.Sprintf("             %s: %s", k, av[0]))
			}
			rpt = append(rpt, "           ]")

		case !present && v != nil:
			rpt = append(rpt, fmt.Sprintf("header not set: %s: %s", k, *v), "           got: [")
			for k, av := range rq.actual.Header {
				rpt = append(rpt, fmt.Sprintf("             %s: %s", k, av[0]))
			}
			rpt = append(rpt, "           ]")

		case v != nil && avs != *v:
			rpt = append(rpt,
				fmt.Sprintf("expected header: %s: %s", k, *v),
				fmt.Sprintf("   got         : %s: %s", k, avs),
			)
		default:
			// NO-OP: header expectations are satisfied
		}
	}
	return rpt
}

// checkMethod returns report describing any exception if the method
// expected to be used by a request was not the method used by the
// corresponding actual request
func (rq *MockRequest) checkBodyExpectation() []string {
	// check the request body vs expected
	if rq.body == nil {
		return nil
	}

	expected := *rq.body
	actual, _ := io.ReadAll(rq.actual.Body)
	defer rq.actual.Body.Close()
	if bytes.Equal(expected, actual) {
		return nil
	}

	switch {
	case len(expected) == 0:
		return []string{
			"expected: <no body>",
			fmt.Sprintf("   got  : %d bytes", len(actual)),
		}
	case len(actual) == 0:
		return []string{
			fmt.Sprintf("expected: %d bytes", len(expected)),
			"   got  : <no body>",
		}
	default:
		rpt := []string{
			"request body differs from expected",
			"   got   :_________",
		}
		for _, b := range bytes.Split(actual, []byte("\n")) {
			rpt = append(rpt, fmt.Sprintf("         |%s", b))
		}
		rpt = append(rpt, "   wanted:_________")
		for _, b := range bytes.Split(expected, []byte("\n")) {
			rpt = append(rpt, fmt.Sprintf("         |%s", b))
		}
		return rpt
	}
}

// String implements the stringer interface for a MockRequest, returning a
// string consisting of the request method (or <ANY> if not specified) and
// url (or <any://hostname/and/path> if not specified)
func (rq MockRequest) String() string {
	m := "<ANY>"
	u := "<any://hostname/and/path>"
	if rq.method != nil {
		m = *rq.method
	}
	if rq.url != "" {
		u = rq.url
	}
	return fmt.Sprintf("%s %s", m, u)
}

// WillNotBeCalled indicates that the request is not expected to be made.  If a
// corresponding request is made by the client, this will be reflected as a failed
// expectation.
func (mock *MockRequest) WillNotBeCalled() {
	mock.isExpected = false
}

// WillRespond establishes a default response for the request, returning a mock
// response to be used to provide details of the response such as status code,
// headers or a body etc.
func (mock *MockRequest) WillRespond() *mockResponse {
	mock.Response = &mockResponse{
		headers: map[string]string{},
	}
	return mock.Response
}

// WillReturnError establishes an error to be returned by the client when
// attempting to perform this request.  Any other response configuration is
// discarded if a request is configured to return an error.
func (mock *MockRequest) WillReturnError(err error) {
	mock.Response = &mockResponse{Err: err}
}

// WithBody identifies the expected body to be sent with the request.
func (mock *MockRequest) WithBody(b []byte) *MockRequest {
	mock.body = &b
	return mock
}

// WithHeader identifies a header expected to be included with the request. The key (k)
// is normalised using textproto.CanonicalMIMEHeaderKey.  An option value (v) may be
// specified; if no value is specified then the header only needs to be present; if a
// value is also specified then the header must be present with the specified value.
//
// If multiple values are specified only the first is significant; additional values
// are discarded.
//
// To configured a non-canonical header, use WithNonCanonicalHeader().
func (mock *MockRequest) WithHeader(k string, v ...string) *MockRequest {
	k = textproto.CanonicalMIMEHeaderKey(k)
	return mock.WithNonCanonicalHeader(k, v...)
}

// WithNonCanonicalHeader identifies a non-canonical header expected to be
// included with the request. The key (k) is expected to match the case as
// specified. An option value (v) may be specified; if no value is specified
// then the header only needs to be present; if a value is also specified
// then the header must be present with the specified value.
//
// If multiple values are specified only the first is significant; additional values
// are discarded.
//
// To configured a canonical header, ensuring that the header key is normalised
// using textproto.CanonicalMIMEHeaderKey, use WithHeader().
func (mock *MockRequest) WithNonCanonicalHeader(k string, v ...string) *MockRequest {
	if len(v) > 0 {
		kv := v[0]
		mock.headers[k] = &kv
	} else {
		mock.headers[k] = nil
	}
	return mock
}
