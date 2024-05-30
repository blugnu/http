package request

import (
	"net/http"
	"testing"

	"github.com/blugnu/test"
)

func TestAccept(t *testing.T) {
	// ARRANGE
	rq, err := http.NewRequest(http.MethodTrace, "notused", nil)
	test.Error(t, err).IsNil()

	// ACT
	err = Accept("application/json")(rq)

	// ASSERT
	test.Error(t, err).IsNil()
	test.Value(t, rq.Header.Get("accept")).Equals("application/json")
}

func TestAcceptJSON(t *testing.T) {
	// ARRANGE
	rq, err := http.NewRequest(http.MethodTrace, "notused", nil)
	test.Error(t, err).IsNil()

	// ACT
	err = AcceptJSON()(rq)

	// ASSERT
	test.Error(t, err).IsNil()
	test.Value(t, rq.Header.Get("accept")).Equals("application/json")
}
