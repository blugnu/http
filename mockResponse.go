package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/textproto"

	"github.com/blugnu/http/multipart"
)

// mockResponse captures the details of the response to be returned when
// responding to an expected request
type mockResponse struct {
	// the body to be returned in the response; may not be used if Value is also set
	body []byte

	// headers to be returned in the response
	headers map[string]string

	// the status code of the response (optional; if not set, 200 (OK) will be used)
	statusCode *int

	// an error to return
	Err error
}

// WithBody sets a body to be returned with the response.
func (resp *mockResponse) WithBody(b []byte) *mockResponse {
	resp.body = b
	return resp
}

// WithJSON sets a body to be returned with the response by marshalling
// a specified value as JSON.
func (resp *mockResponse) WithJSON(v any) *mockResponse {
	var err error
	if resp.body, err = json.Marshal(v); err != nil {
		resp.body = []byte(fmt.Sprintf("WithJSON: %s", err))
	}
	return resp
}

// WithMultipartFormdataFromMap sets a body to be returned with the response
// by mapping the key:value pairs from a supplied map.  A function must also
// be provided to map each k:v pair to the corresponding field, filename and
// data for each part in the multipart form.
//
// NOTE: the response Body value may not be used if a response Value is set
// and a custom request handler is configured.  Refer to the documentation
// for any such handler for details.
func (resp *mockResponse) WithMultipartFormDataFromMap(
	m map[any]any,
	opts ...func(multipart.Options),
) *mockResponse {
	handle := func(err error) *mockResponse {
		sc := http.StatusInternalServerError
		resp.statusCode = &sc
		resp.body = []byte(fmt.Sprintf("MockResponse: WithMultipartFormDataFromMap: %s", err))
		return resp
	}
	ct, body, err := multipart.BodyFromMap(m, opts...)
	if err != nil {
		return handle(err)
	}

	resp.WithHeader("Content-Type", ct)
	resp.body = body

	return resp
}

// WithHeader sets a canonical header to be returned with the response. The key (k)
// is normalised using textproto.CanonicalMIMEHeaderKey.
//
// To configured a non-canonical header, use WithNonCanonicalHeader().
func (resp *mockResponse) WithHeader(k, v string) *mockResponse {
	k = textproto.CanonicalMIMEHeaderKey(k)
	return resp.WithNonCanonicalHeader(k, v)
}

// WithNonCanonicalHeader sets a non-canonical header to be returned with the response.
// The key (k) is set exactly as specified.
//
// To configured a normalised, canonical header, use WithHeader().
func (resp *mockResponse) WithNonCanonicalHeader(k, v string) *mockResponse {
	if resp.headers == nil {
		resp.headers = map[string]string{}
	}
	resp.headers[k] = v
	return resp
}

// WithStatusCode sets the status code to be returned with the response.
func (resp *mockResponse) WithStatusCode(sc int) *mockResponse {
	resp.statusCode = &sc
	return resp
}
