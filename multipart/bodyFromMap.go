package multipart

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
)

// function variables to facilitate testing
var (
	ioCopy = io.Copy

	mpwSetBoundary = func(writer *multipart.Writer, s string) error {
		return writer.SetBoundary(s)
	}
	mpwCreateFormFile = func(
		writer *multipart.Writer,
		fieldname string,
		filename string,
	) (io.Writer, error) {
		return writer.CreateFormFile(fieldname, filename)
	}
	mpwClose = func(writer *multipart.Writer) error {
		return writer.Close()
	}
)

// options holds the options configured for the BodyFromMap function.  This is
// a generic type, with type parameters K and V for the key and value types
// of any configured transform function.
type options[K comparable, V any] struct {
	boundary string
	xform    func(K, V) (string, string, []byte, error)
}

type Options interface {
	setBoundary(string)
}

// setBoundary is an options method to set the string to be used for the
// multipart boundary.  This is not part of the public API; it is used
// internally by the Boundary configuration function.  This avoids the need
// to export the options type.
func (cfg *options[K, V]) setBoundary(s string) {
	cfg.boundary = s
}

// Boundary is a configuration function that sets the boundary string for
// the multipart body.
//
// If no boundary is set then "boundary" is used.
func Boundary(s string) func(Options) {
	return func(cfg Options) {
		cfg.setBoundary(s)
	}
}

// TransformMap sets the transformation function for the BodyFromMap function.
//
// If no transformation function is set then the default transformation is
// applied.  This will create a part for each key:value pair in the map, with:
//
//   - the key as the fieldname
//   - an empty string as the filename
//   - an octet-stream ([]byte) containing the string representation of
//     the value as the content
//
// If the supplied transformation function returns an error for any item
// then this will be returned as the error from BodyFromMap; the returned
// body and content type will be empty and should be ignored.
func TransformMap[K comparable, V any](fn func(K, V) (string, string, []byte, error)) func(Options) {
	return func(cfg Options) {
		cfg.(*options[K, V]).xform = fn
	}
}

// BodyFromMap creates a multipart/formdata encoded body by applying a
// transform function to generate form parts for each item in a map.
// Configuration functions can be used to set the boundary string and
// the transformation function.
//
// # Returns
//
//	string  // the content type for the body
//	[]byte  // the body
//	error   // an error (if non-nil, content type and body should be ignored)
//
// # Configuration Functions
//
//	// to set the boundary string for the body
//	Boundary(string)
//
//	// to set the transformation function for the body
//	TransformMap(func(K, V) (string, string, []byte, error))
//
// If no boundary is configured, "boundary" is used.
//
// If no transformation function is configured a default transformation is
// applied (see: TransformMap for details).
//
// If the transformation function returns an error for any item then this
// will be returned as the error from BodyFromMap; the returned body and
// content type will be empty and should be ignored.
//
// # Example
//
// Demonstrates using the `BodyFromMap` function to create a multipart/formdata
// encoded body from a map, with a custom boundary string and transformation:
//
//	ct, body, err := BodyFromMap(
//		map[string]string{"part-id": "content data"},
//		Boundary("ABCDEF"),
//		MapTransform(func(k, v string) (string, string, []byte, error) {
//			return "field-" + k, "filename-" + k, []byte(v), nil
//		}),
//	)
func BodyFromMap[K comparable, V any](
	m map[K]V,
	opts ...func(Options),
) (string, []byte, error) {
	handle := func(err error) (string, []byte, error) {
		return "", nil, fmt.Errorf("multipart.BodyFromMap: %w", err)
	}

	cfg := &options[K, V]{
		boundary: "boundary",
		xform: func(k K, v V) (string, string, []byte, error) {
			return fmt.Sprintf("%v", k), "", []byte(fmt.Sprintf("%v", v)), nil
		},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	buf := &bytes.Buffer{}
	mpw := multipart.NewWriter(buf)
	if err := mpwSetBoundary(mpw, cfg.boundary); err != nil {
		return handle(fmt.Errorf("writer.SetBoundary: %w", err))
	}

	for k, v := range m {
		fld, filename, data, err := cfg.xform(k, v)
		if err != nil {
			return handle(err)
		}

		file, err := mpwCreateFormFile(mpw, fld, filename)
		if err != nil {
			return handle(fmt.Errorf("writer.CreateFormFile: %w", err))
		}

		_, err = ioCopy(file, bytes.NewReader(data))
		if err != nil {
			return handle(fmt.Errorf("io.Copy: %w", err))
		}
	}

	if err := mpwClose(mpw); err != nil {
		return handle(fmt.Errorf("writer.Close: %w", err))
	}

	ct := mpw.FormDataContentType()
	body := append([]byte{}, buf.Bytes()...)

	return ct, body, nil
}
