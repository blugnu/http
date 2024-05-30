package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"

	"github.com/blugnu/errorcontext"
	"github.com/blugnu/http/request"
)

var (
	ioReadAll      = io.ReadAll
	parseMediaType = mime.ParseMediaType
	nextPart       = func(mpr *multipart.Reader) (*multipart.Part, error) { return mpr.NextPart() }
)

// RequestOption is a function that applies an option to a request
type RequestOption = func(*http.Request) error

// HttpClient is an interface that describes the methods of an http client.
//
// The interface is intended to be used as a wrapper around an http.Client
// or other http client implementation, allowing for the addition of
// additional functionality or configuration.
type HttpClient interface {
	Delete(context.Context, string, ...RequestOption) (*http.Response, error)
	Do(*http.Request) (*http.Response, error)
	Get(context.Context, string, ...RequestOption) (*http.Response, error)
	Patch(context.Context, string, ...RequestOption) (*http.Response, error)
	Post(context.Context, string, ...RequestOption) (*http.Response, error)
	Put(context.Context, string, ...RequestOption) (*http.Response, error)
	NewRequest(context.Context, string, string, ...RequestOption) (*http.Request, error)
}

// ClientInterface is an interface that describes a wrappable http client
type ClientInterface interface {
	Do(*http.Request) (*http.Response, error)
}

// ClientOption is a function that applies an option to a client
type ClientOption func(*client) error

// client is a wrapper around an http.Client that provides additional functionality
// and configuration options.
//
// This type is not exported; functionality is accessed through the implmented
// HttpClient interface.
type client struct {
	// name is used to identify the client in error messages
	name string

	// url is prepended to the url of any request made with the client
	url string

	// wrapped is the underlying http client
	wrapped ClientInterface

	// maxRetries is the maximum number of times a request will be retried
	maxRetries uint
}

// NewClient returns a new HttpClient with the name and url specified, wrapping
// a supplied ClientInterface implementation.  Additional configuration options
// may be optionally specified.
//
// # params
//
//	name  // identifies the client, e.g. in errors
//	opts  // optional configuration
//
// The url typically includes the protocol, hostname and port for the client
// but may include any additional url components consistently required for
// requests performed using the client.
func NewClient(name string, opts ...ClientOption) (HttpClient, error) {
	w := client{
		name:    name,
		wrapped: http.DefaultClient,
	}
	errs := make([]error, 0, len(opts))
	for _, opt := range opts {
		if err := opt(&w); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("%w: %w", ErrInitialisingClient, errors.Join(errs...))
	}
	return w, nil
}

// NewRequest returns a new http.Request with the method and options specified.  The path
// is appended to the client url to form the complete request url.
//
// If a query string is required then it MUST be specified using the provided request
// options:
//
//	request.AddRawQuery("query=string")    // adds to any existing query string
//	request.RawQuery("query=string")          // replaces any existing query string
//	request.Query("key", "value")             // adds a key-value pair to the query string

// request option.  If a "?" is present in the path it will be url encoded when appended
// to the client url:
//
// # Example (Incorrect Usage)
//
//	// c is an http client with a base url of "http://example.com"
//	rq, err := c.NewRequest(ctx, http.MethodGet, "/path?query=string")
//
// The above code will result in a request being made to "http://example.com/path%3Fquery%3Dstring"
//
// # Example (Correct Usage)
//
//	// c is an http client with a base url of "http://example.com"
//	rq, err := c.NewRequest(ctx, http.MethodGet, "/path",
//		request.RawQuery("query=string"),
//	)
//
// The above code will result in a request being made to "http://example.com/path?query=string"
func (c client) NewRequest(
	ctx context.Context,
	method string,
	path string,
	opts ...RequestOption,
) (*http.Request, error) {
	url, err := url.JoinPath(c.url, path)
	if err != nil {
		return nil, errorcontext.Errorf(ctx, "NewRequest: %w: %w", ErrInvalidURL, err)
	}

	rq, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, errorcontext.Errorf(ctx, "NewRequest: %w: %w", ErrInitialisingRequest, err)
	}

	for _, opt := range opts {
		if err := opt(rq); err != nil {
			return nil, errorcontext.Errorf(ctx, "NewRequest: %w", err)
		}
	}

	return rq, nil
}

// do submits a supplied request using the wrapped client.
//
// If an error occurs while submitting the request then it will be resubmitted up
// to the number of retries specified on the request or the client.
//
// If a response is received with a status code that is not http.StatusOK or any
// additional acceptable statuses configured on the request using the request.AcceptStatus()
// option, then the response is returned with an http.ErrUnexpectedResponse error.
func (c client) do(
	ctx context.Context,
	rq *http.Request,
	retries uint,
	accept []uint,
) (*http.Response, error) {
	n := retries
	for {
		r, err := c.wrapped.Do(rq)
		if err != nil {
			switch {
			// no retries were configured
			case retries == 0:
				return r, err

			// retries were configured but have been exhausted
			case n == 0:
				return r, errorcontext.Wrap(ctx, ErrMaxRetriesExceeded, err)

			// at least one retry attempt remains
			default:
				n--
			}
			continue
		}

		// if the response has any of the acceptable status codes then it
		// is returned without error
		for _, sc := range accept {
			if uint(r.StatusCode) == sc {
				return r, nil
			}
		}

		// if we reach this point then we have received a response with a status
		// code that is not acceptable
		return r, errorcontext.Errorf(ctx, "%w: %s", ErrUnexpectedStatusCode, r.Status)
	}
}

// parseRequestHeaders parses the headers of a specified request to identify
// configuration relevant to the execution of the request and initial handling
// of any response.
//
// Any headers found and parsed are removed from the request.
func (c client) parseRequestHeaders(rq *http.Request) (
	maxRetries uint,
	acceptableStatusCodes []uint,
	responseBodyRequired bool,
	streamResponse bool,
	err error,
) {
	ctx := rq.Context()

	parse := func(hdr string, fn func(string) error) error {
		defer delete(rq.Header, hdr)

		if s, ok := rq.Header[hdr]; ok {
			if err := fn(s[0]); err != nil {
				return errorcontext.Errorf(ctx, "%w: %s: %w", ErrInvalidRequestHeader, hdr, err)
			}
		}
		return nil
	}

	// default values if option headers are not present
	maxRetries = c.maxRetries
	acceptableStatusCodes = []uint{http.StatusOK}
	responseBodyRequired = false
	streamResponse = false
	errs := []error{}

	// extract max retries
	errs = append(errs, parse(request.MaxRetriesHeader, func(s string) error {
		i, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		maxRetries = uint(i)
		return nil
	}))

	// extract acceptable statuses
	errs = append(errs, parse(request.AcceptStatusHeader, func(s string) error {
		if err := json.Unmarshal([]byte(s), &acceptableStatusCodes); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidJSON, err)
		}
		return nil
	}))

	// extract response body required flag
	errs = append(errs, parse(request.ResponseBodyRequiredHeader, func(s string) error {
		responseBodyRequired = s == "true"
		return nil
	}))

	// extract stream response flag
	errs = append(errs, parse(request.StreamResponseHeader, func(s string) error {
		streamResponse = s == "true"
		return nil
	}))

	err = errors.Join(errs...)
	return
}

// execute is used by the exported convenience methods to execute a specific method
func (c client) execute(
	ctx context.Context,
	method string,
	url string,
	opts ...RequestOption,
) (*http.Response, error) {
	rq, err := c.NewRequest(ctx, method, url, opts...)
	if err != nil {
		return nil, errorcontext.Errorf(ctx, "%s: %s: %w", c.name, method, err)
	}
	return c.Do(rq)
}

// Do submits a request using the wrapped client, handling the response and
// returning the response or an error.
func (c client) Do(rq *http.Request) (*http.Response, error) {
	ctx := rq.Context()
	handle := func(r *http.Response, err error) (*http.Response, error) {
		return r, errorcontext.Errorf(ctx, "%s: %s %s: %w", c.name, rq.Method, rq.URL, err)
	}

	retries, statusCodes, bodyRequired, stream, err := c.parseRequestHeaders(rq)
	if err != nil {
		return handle(nil, err)
	}

	r, err := c.do(ctx, rq, retries, statusCodes)
	if err != nil {
		return handle(r, err)
	}
	if stream {
		return r, nil
	}

	body, err := ioReadAll(r.Body)
	defer r.Body.Close()

	r.ContentLength = 0
	r.Body = http.NoBody

	switch {
	case err != nil:
		return handle(r, errorcontext.Errorf(ctx, "response.Body: %w", err))

	case len(body) == 0 && bodyRequired:
		return handle(r, ErrNoResponseBody)

	case len(body) == 0:
		return r, nil

	default:
		r.ContentLength = int64(len(body))
		r.Body = io.NopCloser(bytes.NewReader(body))
		return r, nil
	}
}

// Delete is a convenience method for constructing and performing a Delete request,
// appending the specified path to the client url and applying any RequestOptions
func (c client) Delete(
	ctx context.Context,
	path string,
	opts ...RequestOption,
) (*http.Response, error) {
	return c.execute(ctx, http.MethodDelete, path, opts...)
}

// Get is a convenience method for constructing and performing a Get request,
// appending the specified path to the client url and applying any RequestOptions
func (c client) Get(
	ctx context.Context,
	path string,
	opts ...RequestOption,
) (*http.Response, error) {
	return c.execute(ctx, http.MethodGet, path, opts...)
}

// Patch is a convenience method for constructing and performing a Patch request,
// appending the specified path to the client url and applying any RequestOptions
func (c client) Patch(
	ctx context.Context,
	path string,
	opts ...RequestOption,
) (*http.Response, error) {
	return c.execute(ctx, http.MethodPatch, path, opts...)
}

// Post is a convenience method for constructing and performing a Post request,
// appending the specified path to the client url and applying any RequestOptions
func (c client) Post(
	ctx context.Context,
	path string,
	opts ...RequestOption,
) (*http.Response, error) {
	return c.execute(ctx, http.MethodPost, path, opts...)
}

// Put is a convenience method for constructing and performing a Put request,
// appending the specified path to the client url and applying any RequestOptions
func (c client) Put(
	ctx context.Context,
	path string,
	opts ...RequestOption,
) (*http.Response, error) {
	return c.execute(ctx, http.MethodPut, path, opts...)
}

// MapFromMultipartFormData is a generic function that parses an http.Response body expected
// to contain multipart form data, transforming each part into a key-value pair using
// a supplied function.
func MapFromMultipartFormData[K comparable, V any](
	ctx context.Context,
	r *http.Response,
	fn func(string, string, []byte) (K, V, error),
) (map[K]V, error) {
	_, params, err := parseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return nil, errorcontext.Errorf(ctx, "MapFromMultipartFormData: ParseMediaType: %w", err)
	}

	mpr := multipart.NewReader(r.Body, params["boundary"])
	results := make(map[K]V)

	var p *multipart.Part
	for {
		if p, err = nextPart(mpr); err != nil {
			break
		}
		fieldname := p.FormName()
		filename := p.FileName()
		b, err := ioReadAll(p)
		if err != nil {
			return nil, errorcontext.Errorf(ctx, "MapFromMultipartFormData: ReadAll (part): %w", err)
		}
		k, v, err := fn(fieldname, filename, b)
		if err != nil {
			return nil, errorcontext.Errorf(ctx, "MapFromMultipartFormData: transform func: %w", err)
		}
		results[k] = v
	}
	if err != io.EOF {
		return nil, errorcontext.Errorf(ctx, "MapFromMultipartFormData: NextPart: %w", err)
	}

	return results, nil
}

// UnmarshalJSON is a generic function that unmarshals the body of an http.Response
// into a value of a specified type.
//
// The function returns an error if the body cannot be read or if the body does not
// contain valid JSON and the result will be the zero value of the generic type.
func UnmarshalJSON[T any](ctx context.Context, r *http.Response) (T, error) {
	result := *new(T)

	handle := func(sen, err error) (T, error) {
		return result, errorcontext.Errorf(ctx, "http.UnmarshalJSON: %w: %w", sen, err)
	}

	body, err := ioReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return handle(ErrReadingResponseBody, err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return handle(ErrInvalidJSON, err)
	}

	return result, nil
}
