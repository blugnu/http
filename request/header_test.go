package request

import (
	"net/http"
	"testing"

	"github.com/blugnu/test"
)

func TestHeader(t *testing.T) {
	// ARRANGE
	rq, err := http.NewRequest(http.MethodTrace, "notused", nil)
	test.Error(t, err).IsNil()

	// ACT
	// header key is specified in lowercase but being canonical should be normalised
	err = Header("content-type", "application/json")(rq)

	// ASSERT
	test.Error(t, err).IsNil()

	v, exists := rq.Header["Content-Type"]
	test.Value(t, v[0]).Equals("application/json")
	test.IsTrue(t, exists, "is canonicalised")
}

func TestNonCanonicalHeader(t *testing.T) {
	// ARRANGE
	// to verify that canonicalisation is not enforced we specify a
	// canonical header in non-canonical form
	//
	// we use a variable to hold the header key to avoid the staticcheck
	// linter flagging the use of a non-canonical header form (SA1008)
	rq, _ := http.NewRequest(http.MethodTrace, "notused", nil)
	h := "content-type"

	// ACT
	err := NonCanonicalHeader(h, "application/json")(rq)

	// ASSERT
	test.Error(t, err).IsNil()

	v, exists := rq.Header[h]
	test.Value(t, v[0]).Equals("application/json")
	test.IsTrue(t, exists, "not normalised")
}
