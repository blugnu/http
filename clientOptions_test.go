package http

import (
	"net/url"
	"testing"

	"github.com/blugnu/test"
)

func TestMaxRetries(t *testing.T) {
	// ARRANGE
	client := &client{}

	// ACT
	err := MaxRetries(3)(client)

	// ASSERT
	test.That(t, err).IsNil()
	test.That(t, client.maxRetries).Equals(3)
}

func TestClientOptions(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "URL/int",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &client{}

				// ACT
				err := URL(42)(client)

				// ASSERT
				test.Error(t, err).Is(ErrInvalidURL)
			},
		},
		{scenario: "URL/string/invalid url",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &client{}

				// ACT
				err := URL("http://example.com:foo")(client)

				// ASSERT
				test.Error(t, err).Is(ErrInvalidURL)
			},
		},
		{scenario: "URL/string/relative url",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &client{}

				// ACT
				err := URL("example.com")(client)

				// ASSERT
				test.Error(t, err).Is(ErrInvalidURL)
			},
		},
		{scenario: "URL/URL/relative",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &client{}
				url, _ := url.Parse("example.com")

				// ACT
				err := URL(url)(client)

				// ASSERT
				test.Error(t, err).Is(ErrInvalidURL)
			},
		},
		{scenario: "URL/string/successful",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &client{}

				// ACT
				err := URL("http://example.com")(client)

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, client.url).Equals("http://example.com")
			},
		},
		{scenario: "URL/URL/successful",
			exec: func(t *testing.T) {
				// ARRANGE
				client := &client{}
				url, _ := url.Parse("http://example.com")

				// ACT
				err := URL(url)(client)

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, client.url).Equals("http://example.com")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
