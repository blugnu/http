package multipart

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"testing"

	"github.com/blugnu/test"
)

func TestBodyFromMap(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(*testing.T)
		assert   func(*testing.T, string, []byte, error)
	}{
		// BodyFromMap tests
		{scenario: "BodyFromMap/successful",
			exec: func(*testing.T) {
				// NOTE: we encode a map with only one k:v pair to avoid a fragile
				// test case which may break due to changes in the ordering when
				// ranging over the map.  Once go1.22 is adopted this test case
				// could be extended to cover multiple k:v pairs, with a cmp.Ordered
				// constraint added on the generic map key type parameter to enable
				// ranging over the map in explicitly sorted key order

				// ACT
				ct, body, err := BodyFromMap(
					map[string]string{"part-id": "content data"},
					Boundary("boundary"),
					TransformMap(func(k, v string) (string, string, []byte, error) {
						return "field-" + k, "filename-" + k, []byte(v), nil
					}),
				)

				// ASSERT
				wantBody := []byte("--boundary\r\n" +
					"Content-Disposition: form-data; name=\"field-part-id\"; filename=\"filename-part-id\"\r\n" +
					"Content-Type: application/octet-stream\r\n" +
					"\r\n" +
					"content data\r\n" +
					"--boundary--\r\n")

				test.Error(t, err).IsNil()
				test.That(t, ct).Equals("multipart/form-data; boundary=boundary")
				test.Bytes(t, body, "request body", func(v []byte) string { return fmt.Sprintf("[\n%s\n]", string(v)) }).Equals(wantBody)
				test.Bytes(t, body, "request body", 300, test.BytesDecimal).Equals(wantBody)
			},
		},
		{scenario: "BodyFromMap/set boundary error",
			exec: func(*testing.T) {
				// ARRANGE
				berr := errors.New("set boundary error")

				og := mpwSetBoundary
				defer func() { mpwSetBoundary = og }()
				mpwSetBoundary = func(writer *multipart.Writer, s string) error { return berr }

				// ACT
				ct, body, err := BodyFromMap(map[string]string{})

				// ASSERT
				test.Error(t, err).Is(berr)
				test.That(t, ct, "content-type").Equals("")
				test.That(t, body, "body").IsNil()
			},
		},
		{scenario: "BodyFromMap/transformation function error",
			exec: func(*testing.T) {
				// ARRANGE
				maperr := errors.New("map error")

				// ACT
				ct, body, err := BodyFromMap(
					map[string]string{"part": "data"},
					TransformMap(
						func(k, v string) (string, string, []byte, error) {
							return "", "", nil, maperr
						}),
				)

				// ASSERT
				test.Error(t, err).Is(maperr)
				test.That(t, ct).Equals("")
				test.IsTrue(t, body == nil, "body is nil")
			},
		},
		{scenario: "BodyFromMap/create form file error",
			exec: func(*testing.T) {
				// ARRANGE
				formerr := errors.New("form file error")

				og := mpwCreateFormFile
				defer func() { mpwCreateFormFile = og }()
				mpwCreateFormFile = func(writer *multipart.Writer, fieldname, filename string) (io.Writer, error) {
					return nil, formerr
				}

				// ACT
				ct, body, err := BodyFromMap(map[string]string{"part": "data"})

				// ASSERT
				test.Error(t, err).Is(formerr)
				test.That(t, ct).Equals("")
				test.IsTrue(t, body == nil, "body is nil")
			},
		},
		{scenario: "BodyFromMap/file copy error",
			exec: func(*testing.T) {
				// ARRANGE
				copyerr := errors.New("copy file error")

				og := ioCopy
				defer func() { ioCopy = og }()
				ioCopy = func(io.Writer, io.Reader) (int64, error) {
					return 0, copyerr
				}

				// ACT
				ct, body, err := BodyFromMap(map[string]string{"part": "data"})

				// ASSERT
				test.Error(t, err).Is(copyerr)
				test.That(t, ct).Equals("")
				test.IsTrue(t, body == nil, "body is nil")
			},
		},
		{scenario: "BodyFromMap/close error",
			exec: func(*testing.T) {
				// ARRANGE
				closeerr := errors.New("close error")

				og := mpwClose
				defer func() { mpwClose = og }()
				mpwClose = func(*multipart.Writer) error {
					return closeerr
				}

				// ACT
				ct, body, err := BodyFromMap(map[string]string{"part": "data"})

				// ASSERT
				test.Error(t, err).Is(closeerr)
				test.That(t, ct).Equals("", "content type is empty")
				test.IsTrue(t, body == nil, "body is nil")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
