package http

import (
	"errors"
	"testing"

	"github.com/blugnu/test"
)

func TestMockExpectationsError(t *testing.T) {
	// ARRANGE
	sut := MockExpectationsError{
		name: "foo",
		errors: []error{
			errors.New("first error"),
			errors.New("second error"),
		},
	}

	// ACT
	got := sut.Error()

	// ASSERT
	wanted := "foo: expectations not met: [\n" +
		"   first error\n" +
		"   second error\n" +
		"]"
	test.That(t, got).Equals(wanted)
}
