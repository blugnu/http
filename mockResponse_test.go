package http

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/blugnu/http/multipart"
	"github.com/blugnu/test"
)

type unmarshallable struct{}

func (unmarshallable) MarshalJSON() ([]byte, error) {
	return nil, errors.New("unmarshallable")
}

func TestMockResponse(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(*testing.T)
	}{
		{scenario: "WithBody",
			exec: func(t *testing.T) {
				// ARRANGE
				response := &mockResponse{}

				// ACT
				result := response.WithBody([]byte("foo"))

				// ASSERT
				test.That(t, response.body).Equals([]byte("foo"))
				test.IsTrue(t, result == response)
			},
		},
		{scenario: "WithHeader",
			exec: func(t *testing.T) {
				// ARRANGE
				response := &mockResponse{}

				// ACT
				result := response.WithHeader("content-type", "application/json")

				// ASSERT
				test.That(t, response.headers).Equals(map[string]string{"Content-Type": "application/json"})
				test.IsTrue(t, result == response)
			},
		},
		{scenario: "WithJSON/int",
			exec: func(t *testing.T) {
				// ARRANGE
				response := &mockResponse{}

				// ACT
				result := response.WithJSON(42)

				// ASSERT
				test.That(t, response.body).Equals([]byte("42"))
				test.IsTrue(t, result == response)
			},
		},
		{scenario: "WithJSON/string",
			exec: func(t *testing.T) {
				// ARRANGE
				response := &mockResponse{}

				// ACT
				result := response.WithJSON("string value")

				// ASSERT
				test.That(t, response.body).Equals([]byte(`"string value"`))
				test.IsTrue(t, result == response)
			},
		},
		{scenario: "WithJSON/map",
			exec: func(t *testing.T) {
				// ARRANGE
				response := &mockResponse{}

				// ACT
				result := response.WithJSON(map[string]string{"key": "value"})

				// ASSERT
				test.That(t, response.body).Equals([]byte(`{"key":"value"}`))
				test.IsTrue(t, result == response)
			},
		},
		{scenario: "WithJSON/slice",
			exec: func(t *testing.T) {
				// ARRANGE
				response := &mockResponse{}

				// ACT
				result := response.WithJSON([]int{1, 2, 3})

				// ASSERT
				test.That(t, response.body).Equals([]byte(`[1,2,3]`))
				test.IsTrue(t, result == response)
			},
		},
		{scenario: "WithJSON/unmarshallable",
			exec: func(t *testing.T) {
				// ARRANGE
				response := &mockResponse{}

				// ACT
				result := response.WithJSON(unmarshallable{})

				// ASSERT
				test.That(t, string(response.body)).Equals("WithJSON: json: error calling MarshalJSON for type http.unmarshallable: unmarshallable")
				test.IsTrue(t, result == response)
			},
		},
		{scenario: "WithNonCanonicalHeader",
			exec: func(t *testing.T) {
				// ARRANGE
				response := &mockResponse{}

				// ACT
				result := response.WithNonCanonicalHeader("sessionid", "5f6903df-c6e9-4cf6-a95b-fafd76fee730")

				// ASSERT
				test.That(t, response.headers).Equals(map[string]string{"sessionid": "5f6903df-c6e9-4cf6-a95b-fafd76fee730"})
				test.IsTrue(t, result == response)
			},
		},
		{scenario: "WithStatusCode",
			exec: func(t *testing.T) {
				// ARRANGE
				response := &mockResponse{}
				sc := http.StatusInternalServerError

				// ACT
				result := response.WithStatusCode(sc)

				// ASSERT
				test.That(t, *response.statusCode).Equals(sc)
				test.IsTrue(t, result == response)
			},
		},

		// WithMultipartFormDataFromMap tests
		{scenario: "WithMultipartFormDataFromMap/ok",
			exec: func(t *testing.T) {
				// ARRANGE
				response := &mockResponse{}

				// ACT
				result := response.WithMultipartFormDataFromMap(
					map[any]any{"part": "data"},
					multipart.TransformMap(
						func(k, v any) (field string, filename string, data []byte, _ error) {
							field = fmt.Sprintf("field-%s", k.(string))
							filename = fmt.Sprintf("filename-%s", k.(string))
							data = []byte(v.(string))
							return
						}),
				)

				// ASSERT
				test.That(t, string(response.body)).Equals("--boundary\r\n" +
					"Content-Disposition: form-data; name=\"field-part\"; filename=\"filename-part\"\r\n" +
					"Content-Type: application/octet-stream\r\n" +
					"\r\n" +
					"data\r\n" +
					"--boundary--\r\n")
				test.IsTrue(t, result == response)
			},
		},
		{scenario: "WithMultipartFormDataFromMap/body error",
			exec: func(t *testing.T) {
				// ARRANGE
				response := &mockResponse{}
				bodyerr := errors.New("body error")

				// ACT
				result := response.WithMultipartFormDataFromMap(
					map[any]any{"part": "data"},
					multipart.TransformMap(
						func(k, v any) (string, string, []byte, error) {
							return "", "", nil, bodyerr
						}),
				)

				// ASSERT
				test.That(t, string(response.body)).Equals("MockResponse: WithMultipartFormDataFromMap: multipart.BodyFromMap: body error")
				test.IsTrue(t, result == response)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
