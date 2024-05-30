package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blugnu/http/request"
	"github.com/blugnu/test"
)

func TestNewClient(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "no errors",
			exec: func(t *testing.T) {
				// ACT
				result, err := NewClient("name", func(c *client) error { return nil })

				// ASSERT
				test.That(t, err).IsNil()
				test.That(t, result).Equals(client{
					name:    "name",
					wrapped: http.DefaultClient,
				})
			},
		},
		{scenario: "option error",
			exec: func(t *testing.T) {
				opterr := errors.New("option error")
				opts := []ClientOption{func(c *client) error { return opterr }}

				// ACT
				result, err := NewClient("name", opts...)

				// ASSERT
				test.Error(t, err).Is(ErrInitialisingClient)
				test.Error(t, err).Is(opterr)
				test.That(t, result).IsNil()
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}

func TestNewRequest(t *testing.T) {
	// ARRANGE
	ctx := context.Background()

	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "invalid client url",
			// NOTE: there is no test for an invalid request url.  Only the client url must be valid.
			//  anything "invalid" in the request url will be escaped, so cannot be invalid
			exec: func(t *testing.T) {
				// ARRANGE
				c := client{url: "\n"}

				// ACT
				rq, err := c.NewRequest(ctx, http.MethodGet, "some/url")

				// ASSERT
				test.Error(t, err).Is(ErrInvalidURL)
				test.That(t, rq).IsNil()
			},
		},
		{scenario: "invalid method",
			exec: func(t *testing.T) {
				// ARRANGE
				c := client{url: "http://hostname:80"}

				// ACT
				rq, err := c.NewRequest(ctx, " ", "some/url")

				// ASSERT
				test.Error(t, err).Is(ErrInitialisingRequest)
				test.That(t, rq).IsNil()
			},
		},
		{scenario: "option error",
			exec: func(t *testing.T) {
				// ARRANGE
				c := client{url: "http://hostname:80"}
				opterr := errors.New("option error")

				// ACT
				rq, err := c.NewRequest(ctx, http.MethodGet, "some/url", func(*http.Request) error { return opterr })

				// ASSERT
				test.Error(t, err).Is(opterr)
				test.That(t, rq).IsNil()
			},
		},
		{scenario: "valid request",
			exec: func(t *testing.T) {
				// ARRANGE
				c := client{url: "http://hostname:80"}
				want, _ := http.NewRequest(http.MethodPut, "http://hostname:80/some/url", nil)

				// ACT
				rq, err := c.NewRequest(ctx, http.MethodPut, "some/url")

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, rq).Equals(want)
			},
		},
		{scenario: "QueryP execution order",
			exec: func(t *testing.T) {
				// ARRANGE
				c := client{url: "http://hostname:80"}

				// ACT
				rq, err := c.NewRequest(ctx, "GET", "som/url",
					request.QueryP("fizz", nil),
					request.QueryP("buzz", nil),
					request.QueryP("whizz", nil),
				)

				// ASSERT
				test.That(t, err).IsNil()
				test.That(t, rq.URL.RawQuery).Equals("fizz&buzz&whizz")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}

type fakeClient struct {
	body       []byte
	statusCode int
	error
	requests []http.Request
}

func (fake *fakeClient) Do(rq *http.Request) (_ *http.Response, err error) {
	fake.requests = append(fake.requests, *rq)
	if fake.error != nil {
		return nil, fake.error
	}

	rec := httptest.NewRecorder()
	func(rw http.ResponseWriter, _ *http.Request) {
		if fake.statusCode != 0 {
			rw.WriteHeader(fake.statusCode)
		}
		if fake.body != nil {
			if _, err = writeBody(rw, fake.body); err != nil {
				return
			}
		}
	}(rec, nil)

	return rec.Result(), nil
}

func TestDo(t *testing.T) {
	// ARRANGE
	ctx := context.Background()

	testcases := []struct {
		scenario string
		exec     func(*testing.T)
	}{
		{scenario: "wrapped client error",
			exec: func(t *testing.T) {
				// ARRANGE
				wcerr := errors.New("wrapped client error")
				c := client{
					wrapped: &fakeClient{error: wcerr},
				}
				rq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "", nil)

				// ACT
				r, err := c.Do(rq)

				// ASSERT
				test.Error(t, err).Is(wcerr)
				test.That(t, r).IsNil()
			},
		},
		{scenario: "error reading response body",
			exec: func(t *testing.T) {
				// ARRANGE
				readerr := errors.New("read error")
				c := client{
					wrapped: &fakeClient{},
				}
				rq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "", nil)

				og := ioReadAll
				defer func() { ioReadAll = og }()
				ioReadAll = func(io.Reader) ([]byte, error) { return nil, readerr }

				// ACT
				r, err := c.Do(rq)

				// ASSERT
				test.Error(t, err).Is(readerr)
				test.That(t, r).IsNotNil()
				test.That(t, r.ContentLength).Equals(0)
				test.IsTrue(t, r.Body == http.NoBody)
			},
		},
		{scenario: "empty response body",
			exec: func(t *testing.T) {
				// ARRANGE
				c := client{
					wrapped: &fakeClient{body: []byte{}},
				}
				rq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "", nil)

				// ACT
				r, err := c.Do(rq)

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, r).IsNotNil()
				test.That(t, r.ContentLength).Equals(0)
				test.IsTrue(t, r.Body == http.NoBody)
			},
		},
		{scenario: "non-empty response body",
			exec: func(t *testing.T) {
				// ARRANGE
				c := client{
					wrapped: &fakeClient{body: []byte("body")},
				}
				rq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "", nil)

				// ACT
				r, err := c.Do(rq)

				// ASSERT
				body, _ := io.ReadAll(r.Body)
				defer r.Body.Close()

				test.Error(t, err).IsNil()
				test.That(t, r).IsNotNil()
				test.That(t, r.ContentLength).Equals(4)
				test.Bytes(t, body).Equals([]byte("body"))
			},
		},
		{scenario: "retries/configured on client",
			exec: func(t *testing.T) {
				// ARRANGE
				permerr := errors.New("permanent failure")
				fake := &fakeClient{error: permerr}
				c := client{
					wrapped:    fake,
					maxRetries: 2,
				}
				rq, _ := http.NewRequest("", "", nil)

				// ACT
				r, err := c.Do(rq)

				// ASSERT
				// maxRetries is 2, so there should be 3 requests made, including the initial failed request
				test.Error(t, err).Is(permerr)
				test.That(t, r).IsNil()
				test.That(t, len(fake.requests)).Equals(3)
			},
		},
		{scenario: "retries/request overrides client",
			exec: func(t *testing.T) {
				// ARRANGE
				permerr := errors.New("permanent failure")
				fake := &fakeClient{error: permerr}
				c := client{
					wrapped:    fake,
					maxRetries: 2,
				}
				rq, _ := http.NewRequest("", "", nil)
				rq.Header[request.MaxRetriesHeader] = []string{"1"}

				// ACT
				r, err := c.Do(rq)

				// ASSERT
				// although 2 retries are specified on the client, maxRetries is 1 on the request,
				// so there should be only 2 requests made, including the initial failed request
				test.Error(t, err).Is(permerr)
				test.That(t, r).IsNil()
				test.That(t, len(fake.requests)).Equals(2)
			},
		},
		{scenario: "retries/invalid request header",
			exec: func(t *testing.T) {
				// ARRANGE
				fake := &fakeClient{}
				c := client{wrapped: fake}
				rq, _ := http.NewRequest("", "", nil)
				rq.Header[request.MaxRetriesHeader] = []string{"invalid"}

				// ACT
				r, err := c.Do(rq)

				// ASSERT
				test.Error(t, err).Is(ErrInvalidRequestHeader)
				test.That(t, r).IsNil()
				test.That(t, len(fake.requests)).Equals(0)
			},
		},
		{
			scenario: "acceptable status",
			exec: func(t *testing.T) {
				// ARRANGE
				fake := &fakeClient{statusCode: http.StatusNotFound}
				c := client{wrapped: fake}
				rq, _ := http.NewRequest("", "", nil)
				rq.Header[request.AcceptStatusHeader] = []string{"[200,404]"}

				// ACT
				r, err := c.Do(rq)

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, r.StatusCode).Equals(http.StatusNotFound)

				sent := fake.requests[0]
				test.That(t, sent.Header[request.AcceptStatusHeader]).IsNil()
			},
		},
		{
			scenario: "acceptable status/unacceptable",
			exec: func(t *testing.T) {
				// ARRANGE
				fake := &fakeClient{statusCode: http.StatusUnauthorized}
				c := client{wrapped: fake}
				rq, _ := http.NewRequest("", "", nil)
				rq.Header[request.AcceptStatusHeader] = []string{"[200,404]"}

				// ACT
				r, err := c.Do(rq)

				// ASSERT
				test.Error(t, err).Is(ErrUnexpectedStatusCode)
				test.That(t, r.StatusCode).Equals(http.StatusUnauthorized)

				sent := fake.requests[0]
				test.That(t, sent.Header[request.AcceptStatusHeader]).IsNil()
			},
		},
		{
			scenario: "acceptable status/malformed header",
			exec: func(t *testing.T) {
				// ARRANGE
				fake := &fakeClient{statusCode: http.StatusUnauthorized}
				c := client{wrapped: fake}
				rq, _ := http.NewRequest("", "", nil)
				rq.Header[request.AcceptStatusHeader] = []string{"this is not json"}

				// ACT
				r, err := c.Do(rq)

				// ASSERT
				test.Error(t, err).Is(ErrInvalidJSON)
				test.That(t, r).IsNil()
			},
		},
		{scenario: "response body required/present",
			exec: func(t *testing.T) {
				// ARRANGE
				fake := &fakeClient{body: []byte("body")}
				c := client{wrapped: fake}
				rq, _ := http.NewRequest("", "", nil)
				rq.Header[request.ResponseBodyRequiredHeader] = []string{""}

				// ACT
				r, err := c.Do(rq)

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, r.ContentLength).Equals(4)
			},
		},
		{scenario: "response body required/empty",
			exec: func(t *testing.T) {
				// ARRANGE
				fake := &fakeClient{body: []byte{}}
				c := client{wrapped: fake}
				rq, _ := http.NewRequest("", "", nil)
				rq.Header[request.ResponseBodyRequiredHeader] = []string{"true"}

				// ACT
				r, err := c.Do(rq)

				// ASSERT
				test.Error(t, err).Is(ErrNoResponseBody)
				test.That(t, r.ContentLength).Equals(0)
				test.IsTrue(t, r.Body == http.NoBody)
			},
		},
		{scenario: "stream response",
			exec: func(t *testing.T) {
				// ARRANGE
				fake := &fakeClient{body: []byte("non-empty")}
				c := client{wrapped: fake}
				rq, _ := http.NewRequest("", "", nil)
				rq.Header[request.StreamResponseHeader] = []string{"true"}

				// ACT
				r, err := c.Do(rq)

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, r.ContentLength).Equals(-1)
				test.IsTrue(t, r.Body != http.NoBody)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}

func TestConvenienceMethods(t *testing.T) {
	// ARRANGE
	ctx := context.Background()

	testcases := []struct {
		scenario string
		exec     func(*testing.T)
	}{
		// do tests
		{scenario: "do/invalid client url",
			exec: func(t *testing.T) {
				// ARRANGE
				c := client{url: "\n"}

				// ACT
				r, err := c.execute(ctx, http.MethodTrace, "")

				// ASSERT
				test.Error(t, err).Is(ErrInvalidURL)
				test.That(t, r).IsNil()
			},
		},
		{scenario: "do/invalid method",
			exec: func(t *testing.T) {
				// ARRANGE
				c := client{url: "http://host"}

				// ACT
				r, err := c.execute(ctx, "\n", "")

				// ASSERT
				test.Error(t, err).Is(ErrInitialisingRequest)
				test.That(t, r).IsNil()
			},
		},
		{scenario: "do/successful request",
			exec: func(t *testing.T) {
				// ARRANGE
				fake := &fakeClient{body: []byte{}}
				c := client{url: "http://host", wrapped: fake}

				// ACT
				r, err := c.execute(ctx, http.MethodTrace, "url")

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, r.StatusCode).Equals(http.StatusOK)
				test.That(t, fake.requests[0].URL.String()).Equals("http://host/url")
				test.That(t, fake.requests[0].Method).Equals(http.MethodTrace)
			},
		},

		// Delete/Get/Patch/Post/Put tests
		{scenario: "Delete",
			exec: func(t *testing.T) {
				// ARRANGE
				fake := &fakeClient{body: []byte{}}
				c := client{url: "http://host", wrapped: fake}

				// ACT
				r, err := c.Delete(ctx, "url")

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, r.StatusCode).Equals(http.StatusOK)
				test.That(t, fake.requests[0].URL.String()).Equals("http://host/url")
				test.That(t, fake.requests[0].Method).Equals(http.MethodDelete)
			},
		},
		{scenario: "Get",
			exec: func(t *testing.T) {
				// ARRANGE
				fake := &fakeClient{body: []byte{}}
				c := client{url: "http://host", wrapped: fake}

				// ACT
				r, err := c.Get(ctx, "url")

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, r.StatusCode).Equals(http.StatusOK)
				test.That(t, fake.requests[0].URL.String()).Equals("http://host/url")
				test.That(t, fake.requests[0].Method).Equals(http.MethodGet)
			},
		},
		{scenario: "Patch",
			exec: func(t *testing.T) {
				// ARRANGE
				fake := &fakeClient{body: []byte{}}
				c := client{url: "http://host", wrapped: fake}

				// ACT
				r, err := c.Patch(ctx, "url")

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, r.StatusCode).Equals(http.StatusOK)
				test.That(t, fake.requests[0].URL.String()).Equals("http://host/url")
				test.That(t, fake.requests[0].Method).Equals(http.MethodPatch)
			},
		},
		{scenario: "Post",
			exec: func(t *testing.T) {
				// ARRANGE
				fake := &fakeClient{body: []byte{}}
				c := client{url: "http://host", wrapped: fake}

				// ACT
				r, err := c.Post(ctx, "url")

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, r.StatusCode).Equals(http.StatusOK)
				test.That(t, fake.requests[0].URL.String()).Equals("http://host/url")
				test.That(t, fake.requests[0].Method).Equals(http.MethodPost)
			},
		},
		{scenario: "Put",
			exec: func(t *testing.T) {
				// ARRANGE
				fake := &fakeClient{body: []byte{}}
				c := client{url: "http://host", wrapped: fake}

				// ACT
				r, err := c.Put(ctx, "url")

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, r.StatusCode).Equals(http.StatusOK)
				test.That(t, fake.requests[0].URL.String()).Equals("http://host/url")
				test.That(t, fake.requests[0].Method).Equals(http.MethodPut)
			},
		},

		// MapFromMultipartFormData tests
		{scenario: "MapFromMultipartFormData/parse media error",
			exec: func(t *testing.T) {
				// ARRANGE
				parseerr := errors.New("parse error")
				r := &http.Response{}
				og := parseMediaType
				defer func() { parseMediaType = og }()
				parseMediaType = func(v string) (string, map[string]string, error) { return "", nil, parseerr }

				// ACT
				result, err := MapFromMultipartFormData[string, string](ctx, r, nil)

				// ASSERT
				test.Error(t, err).Is(parseerr)
				test.That(t, result).IsNil()
			},
		},
		{scenario: "MapFromMultipartFormData/part error",
			exec: func(t *testing.T) {
				// ARRANGE
				parterr := errors.New("part error")
				r := &http.Response{
					Header: map[string][]string{
						"Content-Type": {"multipart-formdata; boundary=boundary"},
					},
				}
				og := nextPart
				defer func() { nextPart = og }()
				nextPart = func(*multipart.Reader) (*multipart.Part, error) { return nil, parterr }

				// ACT
				result, err := MapFromMultipartFormData[string, string](ctx, r, nil)

				// ASSERT
				test.Error(t, err).Is(parterr)
				test.That(t, result).IsNil()
			},
		},
		{scenario: "MapFromMultipartFormData/part read error",
			exec: func(t *testing.T) {
				// ARRANGE
				readerr := errors.New("part read error")
				r := &http.Response{
					Header: map[string][]string{
						"Content-Type": {"multipart/form-data; boundary=boundary"},
					},
					Body: io.NopCloser(bytes.NewReader([]byte("--boundary\r\n" +
						"Content-Disposition: form-data; name=\"1\"; filename=\"file1.txt\"\r\n" +
						"Content-Type: application/text\r\n" +
						"\r\n" +
						"content\r\n" +
						"--boundary--",
					))),
				}
				og := ioReadAll
				defer func() { ioReadAll = og }()
				ioReadAll = func(r io.Reader) ([]byte, error) { return nil, readerr }

				// ACT
				result, err := MapFromMultipartFormData[string, string](ctx, r, nil)

				// ASSERT
				test.Error(t, err).Is(readerr)
				test.That(t, result).IsNil()
			},
		},
		{scenario: "MapFromMultipartFormData/transform error",
			exec: func(t *testing.T) {
				// ARRANGE
				xformerr := errors.New("transform error")
				r := &http.Response{
					Header: map[string][]string{
						"Content-Type": {"multipart/form-data; boundary=boundary"},
					},
					Body: io.NopCloser(bytes.NewReader([]byte("--boundary\r\n" +
						"Content-Disposition: form-data; name=\"1\"; filename=\"file1.txt\"\r\n" +
						"Content-Type: application/text\r\n" +
						"\r\n" +
						"content\r\n" +
						"--boundary--",
					))),
				}

				// ACT
				result, err := MapFromMultipartFormData[string, string](ctx, r, func(field, filename string, data []byte) (string, string, error) {
					return "", "", xformerr
				})

				// ASSERT
				test.Error(t, err).Is(xformerr)
				test.That(t, result).IsNil()
			},
		},
		{scenario: "MapFromMultipartFormData/success",
			exec: func(t *testing.T) {
				// ARRANGE
				r := &http.Response{
					Header: map[string][]string{
						"Content-Type": {"multipart/form-data; boundary=boundary"},
					},
					Body: io.NopCloser(bytes.NewReader([]byte("--boundary\r\n" +
						"Content-Disposition: form-data; name=\"1\"; filename=\"file1.txt\"\r\n" +
						"Content-Type: application/text\r\n" +
						"\r\n" +
						"content\r\n" +
						"--boundary--",
					))),
				}

				// ACT
				result, err := MapFromMultipartFormData[string, string](ctx, r, func(name, filename string, data []byte) (string, string, error) {
					return fmt.Sprintf("%s:%s", name, filename), string(data), nil
				})

				// ASSERT
				test.Error(t, err).IsNil()
				if result, ok := test.IsType[map[string]string](t, result); ok {
					test.Map(t, result).Equals(map[string]string{"1:file1.txt": "content"})
				}
			},
		},

		// UnmarshalJSON tests
		{scenario: "UnmarshalJSON/error reading body",
			exec: func(t *testing.T) {
				// ARRANGE
				readerr := errors.New("read error")
				response := &http.Response{Body: http.NoBody}

				og := ioReadAll
				defer func() { ioReadAll = og }()
				ioReadAll = func(r io.Reader) ([]byte, error) { return nil, readerr }

				// ACT
				result, err := UnmarshalJSON[map[string]string](ctx, response)

				// ASSERT
				test.Error(t, err).Is(readerr)
				test.That(t, result).IsNil()
			},
		},
		{scenario: "UnmarshalJSON/error unmarshalling",
			exec: func(t *testing.T) {
				// ARRANGE
				response := &http.Response{Body: io.NopCloser(bytes.NewReader([]byte("not valid JSON")))}

				// ACT
				result, err := UnmarshalJSON[map[string]string](ctx, response)

				// ASSERT
				test.Error(t, err).Is(ErrInvalidJSON)
				test.That(t, result).IsNil()
			},
		},
		{scenario: "UnmarshalJSON/incorrect type",
			exec: func(t *testing.T) {
				// ARRANGE
				response := &http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"key":"value"}`)))}

				// ACT
				result, err := UnmarshalJSON[int](ctx, response)

				// ASSERT
				test.Error(t, err).Is(ErrInvalidJSON)
				test.That(t, result).Equals(0)
			},
		},
		{scenario: "UnmarshalJSON/ok",
			exec: func(t *testing.T) {
				// ARRANGE
				response := &http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"key":"value"}`)))}

				// ACT
				result, err := UnmarshalJSON[map[string]string](ctx, response)

				// ASSERT
				test.Error(t, err).Is(nil)
				test.That(t, result).Equals(map[string]string{"key": "value"})
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
