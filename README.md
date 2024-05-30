<div align="center" style="margin-bottom:20px">
  <img src=".assets/banner.png" alt="http" />
  <div align="center">
    <a href="https://github.com/blugnu/http/actions/workflows/release.yml">
      <img alt="build-status" src="https://github.com/blugnu/http/actions/workflows/release.yml/badge.svg"/>
    </a>
    <a href="https://goreportcard.com/report/github.com/blugnu/http" >
      <img alt="go report" src="https://goreportcard.com/badge/github.com/blugnu/http"/>
    </a>
    <a>
      <img alt="go version >= 1.14" src="https://img.shields.io/github/go-mod/go-version/blugnu/http?style=flat-square"/>
    </a>
    <a href="https://github.com/blugnu/http/blob/master/LICENSE">
      <img alt="MIT License" src="https://img.shields.io/github/license/blugnu/http?color=%234275f5&style=flat-square"/>
    </a>
    <a href="https://coveralls.io/github/blugnu/http?branch=master">
      <img alt="coverage" src="https://img.shields.io/coveralls/github/blugnu/http?style=flat-square"/>
    </a>
    <a href="https://pkg.go.dev/github.com/blugnu/http">
      <img alt="docs" src="https://pkg.go.dev/badge/github.com/blugnu/http"/>
    </a>
  </div>
</div>

# blugnu/http

A `net/http.Client` wrapper with quality of life improvements:

- [x] Configurable request retries
- [x] Simplified response handling
- [x] Multipart form data transformation to/from maps
- [x] JSON marshalling helpers for request and response bodies
- [x] A mock client for request and response mocking

# Installation

`go get github.com/blugnu/http`

# Using the Client

The `NewClient()` function in the `github.com/blugnu/http` package is used to create a new `http.Client`:

| param | type            | description |
| ----- | --------------- | ----------- |
| name  | string          | a name for the client, used in error messages and test failure reports |
| url   | string          | the base url for the client |
| opts  | ...ClientOption | optional client configuration |

The function returns an `HttpClient` interface providing the following methods:

<!-- markdownlint-disable MD013 -->
| method | description |
| ------ | ----------- |
| `NewRequest(ctx context.Context, method string, path string, opts ...RequestOption) (*http.Request, error)` | creates a new request with the specified method and path, and additional request options as specified |
| `Delete(ctx context.Context, url string, opts ...RequestOption) (*http.Response, error)` | performs a DELETE request using a specified path and request options as specified |
| `Get(ctx context.Context, url string, opts ...RequestOption) (*http.Response, error)` | performs a GET request using a specified path and request options as specified |
| `Patch(ctx context.Context, url string, opts ...RequestOption) (*http.Response, error)` | performs a PATCH request using a specified path and request options as specified |
| `Post(ctx context.Context, url string, opts ...RequestOption) (*http.Response, error)` | performs a POST request using a specified path and request options as specified |
| `Put(ctx context.Context, url string, opts ...RequestOption) (*http.Response, error)` | performs a PUT request using a specified path and request options as specified |
| `Do(rq *http.Request) (*http.Response, error)` | performs a request using the specified `http.Request`, initialised separately |
<!-- markdownlint-restore -->

## Response Handling

The client in this module provides extended handling of responses, to simplify error handling in
the code using the client. In addition to any error that might result from attempting to perform
the request, the following additional errors may also be returned with or without a response:

<!-- markdownlint-disable MD013 -->
| error                          | response included | description |
| ------------------------------ | ----------------- | ----------- |
| `http.ErrNoResponseBody`       | yes               | returned if the response body is empty and the `request.ResponseBodyRequired()` request option was specified; NOTE: _will never be returned if `request.StreamResponse()` is also specified_ |
| `http.ErrUnexpectedStatusCode` | yes               | returned if the response has a status code other than `http.StatusOK` and which is not identified as acceptable using the `request.AcceptStatus()` request option |
| `http.ErrMaxRetriesExceeded`   | no                | returned if the request was retried the maximum number of times specified for the request |
<!-- markdownlint-restore -->

> Maximum retries for a request are determined by the `request.MaxRetries()` request option or
> a `http.MaxRetries` client option configured on the client used to make the request.  When a
> `http.ErrMaxRetriesExceeded` error is returned it is wrapped with the error that occurred returned
> when making the final, failed request

### Acceptable Status Codes

By default, the only acceptable status code for a response is `http.StatusOK`.  A response with any
other status code will result in an `http.ErrUnexpectedStatusCode` error. This may be overridden
using the `request.AcceptStatus()` request option, which configures the request to treat the
specified status code as acceptable.

### Examples

#### : response body is expected

```golang
r, err := client.Get(ctx, "v1/customer",
    request.ResponseBodyRequired(),
)
if err != nil {
    return err
}

// ... proceed with processing the response body
```

#### : return a specific error when receiving 404 Not Found

```golang
r, err := client.Get(ctx, "v1/customer",
    request.AcceptStatus(http.StatusNotFound),
)
if err != nil {
    return err
}
switch {
    case r.StatusCode == http.StatusNotFound:
        return ErrCustomerNotFound

    default:
        // can only be an OK response; client.Get() would otherwise 
        // have returned ErrUnexpectedStatusCode
}
```

# Request Options

Request options are used to configure the properties of a request. The following request options
are provided:

<!-- markdownlint-disable MD013 -->
| option | description |
| ------ | ----------- |
| `request.Accept()`                   | adds an `Accept` header to the request |
| `request.AcceptStatus()`             | configures the request to accept a specific status code |
| `request.BearerToken()`              | adds an `Authorization` header with a value of `Bearer` |
| `request.Body()`                     | adds a body to the request |
| `request.ContentType()`              | adds a `Content-Type` header to the request |
| `request.Header()`                   | adds a canonical header to the request |
| `request.JSONBody()`                 | adds a JSON body to the request, marshalling a supplied `any` |
| `request.MaxRetries()`               | configures the request to be retried; overrides any retries configured on the client |
| `request.MultipartFormDataFromMap()` | adds a multipart form data body to the request |
| `request.NonCanonicalHeader()`       | adds a non-canonical header to the request |
| `request.Query()`                    | adds a map of query parameters to the request |
| `request.QueryP()`                   | adds an individual `key:value` parameter to the request query |
| `request.RawQuery()`                 | specifies an appropriately url encoded query string for the request |
| `request.StreamResponse()`           | configures the response to be streamed |
<!-- markdownlint-restore -->

Some of these options can affect the behaviour of the client when processing a response:

<!-- markdownlint-disable MD013 -->
| option                           | affect on client |
| -------------------------------- | ---------------- |
| `request.AcceptStatus()`         | prevents the client from returning an error if the response status code is configured as acceptable |
| `request.MaxRetries()`           | causes the client to retry the request if the response status code is not acceptable; overrides any `http.MaxRetries()` option if specified on the client used to perform the request |
| `request.ResponseBodyRequired()` | causes the client to return an error if the response body is empty; has no effect if `request.StreamResponse()` is also specified |
| `request.StreamResponse()`       | causes the response body to be streamed |
<!-- markdownlint-restore -->

## Multipart Form Data

### Requests

To submit a multipart form data body with a request, the `request.MultipartFormDataFromMap()` request
option may be used.

This is a generic function with type parameters for key and value types in a supplied `map`.  These
types will be inferred from a function that must also be provided to be called for each `key:value`
in the map to encode that `key:value` as an individual part in the form data.

The supplied function must accepts a key and value parameter of the keys and values in the map; the
function must return a field name `string`, filename `string` and data `[]byte` for each part, or
an `error`.

<!-- markdownlint-disable MD013 -->
```golang
resp, err := client.Post(ctx, "v1/documents",
        request.MultipartFormDataFromMap(docs, func(id string, doc Document) (string, string, []byte, error) {
            return doc.id, doc.filename, doc.Content, nil
        }),
    )
```
<!-- markdownlint-restore -->

### Responses

When handling responses containing multipart form data, a corresponding function is
provided that will parse a response containing a multipart form data body and transform
it into a map: `MapFromMultipartFormData()`.

This is again a generic function also accepting a function which in this case performs the
transformation in reverse. The function is called with the field name, filename and data
for each part in the multipart form and must return a `key:value` pair to be stored in
the map, or an error.

```golang
    docs, err := http.MapFromMultipartFormData[string, []byte](ctx, r,
        func(field, filename string, data []byte) (string, []byte, error) {
            return filename, data, nil
        })
    if err != nil {
        return err
    }
```

<hr>

# Mocking

This module provides two facilities for mocking http Client behaviors:

1. testing that code under test issues the expected requests
2. providing mock responses to http requests issues by code under test

Both use cases start with creating a mock client using the `NewMockClient()` function:

```go
   client, mock := http.NewMockClient("client")
```

The name argument to the function is used in error messages and test failure reports to
identify the client involved.

The `client` returned from this function should be injected into code under test, to
replace the production `Client`.

The `mock` returned by the function is used to set and test expected request properties
and to establish mock responses for those requests.

## Using a Mock to Verify Expected Requests

```golang
    mock.ExpectGet("v1/customer")
```

This configures the mock to expect a `GET` request to the specified url.  With no other
configuration specified, any `GET` request will satisfy this expectation.  Normally,
specific properties of the expected request will be configured using the fluent api for
configuring expected request properties.

For example, if the url involved required an authorization header then it would be typical
to specify that the request is expected to include the appropriate header:

```golang
    mock.ExpectGet("v1/customer").
        WithHeader("Authorisation")
```

After the code under test has been executed, the mock may then be used to verify that the
expected requests were made with the correct properties using the `ExpectationsWereMet()`
method of the mock. This returns an error describing any expectations that were not
satisfied or `nil` if all expectations were met:

```golang
    // ARRANGE
    mock.ExpectGet("v1/customer").
        WithHeader("Authorisation")

    // ACT
    ...

    // ASSERT
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
```

## Mocking Responses

If no response details are configured for an expected request, the mock client will provide
a `200 OK` response with no body or headers.

This is configurable using the fluent api returned by a mocked request to configure the
response to be returned.

For example, to mock a `403 Forbidden` response:

```golang
    mock.ExpectGet("v1/customer").
        WithHeader("Authorisation").
        WillRespond().WithStatusCode(http.StatusForbidden)
```

To provide more detailed configuration of a response, identifying one or more headers, body
and status code details, the `WillRespond()` method provides a response configuration fluent api:

```golang
    mock.ExpectGet("v1/customer").
        WithHeader("Authorisation").
        WillRespond().
            WithHeader("Content-Type", "application/json").
            WithBody([]byte(`{"id":1,"name":"Jane Smith"}`))
```
