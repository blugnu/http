package request

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/blugnu/http/multipart"
	"github.com/blugnu/test"
)

func TestMultipartFormData(t *testing.T) {
	// ARRANGE
	bodyerr := errors.New("body error")

	testcases := []struct {
		scenario string
		exec     func(*testing.T, *http.Request)
	}{
		{scenario: "MultipartFormDataFromMap/successful",
			exec: func(t *testing.T, rq *http.Request) {
				// NOTE: we encode a map with only one k:v pair to avoid a fragile
				// test case which may break due to changes in the ordering when
				// ranging over the map.  Once go1.22 is adopted this test case
				// could be extended to cover multiple k:v pairs, with a cmp.Ordered
				// constraint added on the generic map key type parameter to enable
				// ranging over the map in explicitly sorted key order

				// ACT
				err := MultipartFormDataFromMap(
					map[string]string{
						"part-id": "content data",
					},
					multipart.TransformMap(func(k, v string) (string, string, []byte, error) {
						return "field-" + k, "filename-" + k, []byte(v), nil
					}),
				)(rq)

				// ASSERT
				body, _ := io.ReadAll(rq.Body)
				defer rq.Body.Close()

				wantBody := []byte("--boundary\r\n" +
					"Content-Disposition: form-data; name=\"field-part-id\"; filename=\"filename-part-id\"\r\n" +
					"Content-Type: application/octet-stream\r\n" +
					"\r\n" +
					"content data\r\n" +
					"--boundary--\r\n")

				test.Error(t, err).IsNil()
				test.That(t, rq.Header.Get("Content-Type")).Equals("multipart/form-data; boundary=boundary")
				test.Bytes(t, body, "request body", func(v []byte) string { return fmt.Sprintf("[\n%s\n]", string(v)) }).Equals(wantBody)
				test.Bytes(t, body, "request body", 300, test.BytesDecimal).Equals(wantBody)
			},
		},
		{scenario: "MultipartFormDataFromMap/BodyFromMap returns error",
			exec: func(t *testing.T, rq *http.Request) {
				// NOTE: we encode a map with only one k:v pair to avoid a fragile
				// test case which may break due to changes in the ordering when
				// ranging over the map.  Once go1.22 is adopted this test case
				// could be extended to cover multiple k:v pairs, with a cmp.Ordered
				// constraint added on the generic map key type parameter to enable
				// ranging over the map in explicitly sorted key order

				// ACT
				err := MultipartFormDataFromMap(
					map[string]string{
						"part-id": "content data",
					},
					multipart.TransformMap(func(k, v string) (string, string, []byte, error) {
						return "", "", nil, bodyerr
					}),
				)(rq)

				// ASSERT
				test.Error(t, err).Is(bodyerr)
				test.That(t, rq.Header.Get("Content-Type")).Equals("")
				test.That(t, rq.Body).IsNil()
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			rq, _ := http.NewRequest(http.MethodTrace, "notused", nil)
			tc.exec(t, rq)
		})
	}
}
