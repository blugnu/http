package request

import (
	"net/http"
	"testing"

	"github.com/blugnu/test"
)

func TestContentType(t *testing.T) {
	// ARRANGE
	rq, err := http.NewRequest(http.MethodTrace, "notused", nil)
	test.Error(t, err).IsNil()

	// ACT
	err = ContentType("application/json")(rq)

	// ASSERT
	test.Error(t, err).IsNil()
	test.Value(t, rq.Header.Get("Content-Type")).Equals("application/json")
}
