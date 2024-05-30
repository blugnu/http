package http

import (
	"fmt"
	"net/http"
	"net/url"
)

// MaxRetries sets the maximum number of retries for requests made using the client.
// Individual requests may be configured to override this value on a case-by-case basis.
func MaxRetries(n uint) ClientOption {
	return func(c *client) error {
		c.maxRetries = n
		return nil
	}
}

// URL sets the base URL for requests made using the client.  The URL may be specified
// as a string or a *url.URL.
//
// If a string is provided, it will be parsed to ensure it is a valid, absolute URL.
//
// If a URL is provided is must be absolute.
func URL(u any) ClientOption {
	return func(c *client) error {
		switch u := u.(type) {
		case string:
			url, err := url.Parse(u)
			if err != nil {
				return fmt.Errorf("http: URL option: %w: %w", ErrInvalidURL, err)
			}
			return URL(url)(c)

		case *url.URL:
			if !u.IsAbs() {
				return fmt.Errorf("http: URL option: %w: URL must be absolute", ErrInvalidURL)
			}
			c.url = u.String()

		default:
			return fmt.Errorf("http: URL option: %w: must be a string or *url.URL", ErrInvalidURL)
		}
		return nil
	}
}

// Using sets the HTTP client to use for requests made using the client.  Any value
// that implements the `Do(*http.Request) (*http.Response, error)` method may be used.
func Using(httpClient interface {
	Do(*http.Request) (*http.Response, error)
}) ClientOption {
	return func(c *client) error {
		c.wrapped = httpClient
		return nil
	}
}
