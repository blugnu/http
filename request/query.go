package request

import (
	"fmt"
	"net/http"
	"net/url"
)

// Query adds all key-value pairs in a supplied map to the query of a request.
//
// Keys are url encoded before being added to the query. If any value is nil
// the corresponding key is added to the query with no value:
//
//	request.Query(map[string]any{"foo", nil}) -> ?foo
//
// If the value is not nil it also is url encoded before being added to
// the query:
//
//	request.Query(map[string]any{"foo", true}) -> ?foo=true
//	request.Query(map[string]any{"foo", "subkey:value"}) -> ?foo=subkey%3Avalue
//
// The order of the keys in the resulting query is not guaranteed as a result
// of map iteration order being undefined in Go. If key order is important
// then QueryP should be called to add each key-value pair individually in the
// desired order.
func Query(params map[string]any) func(*http.Request) error {
	return func(r *http.Request) error {
		for k, v := range params {
			_ = QueryP(k, v)(r)
		}
		return nil
	}
}

// QueryP adds a single key-value pair (or Param) to the query of a request.
//
// The key is url encoded before being added to the query.  If the value is nil
// the key is added to the query with no value:
//
//	request.QueryP("foo", nil) -> ?foo
//
// If the value is not nil it is url encoded before being added to the query:
//
//	request.QueryP("foo", true) -> ?foo=true
//	request.QueryP("'a map'", "key=value") -> ?%27a+map%27=key%3Dvalue
func QueryP(k string, v any) func(*http.Request) error {
	return func(rq *http.Request) error {
		append := func(s string) {
			switch len(rq.URL.RawQuery) {
			case 0:
				rq.URL.RawQuery = s
			default:
				rq.URL.RawQuery += "&" + s
			}
		}

		k = url.QueryEscape(k)
		switch {
		case v == nil:
			append(k)
		default:
			s := k + "=" + url.QueryEscape(fmt.Sprintf("%v", v))
			append(s)
		}
		return nil
	}
}

// RawQuery sets the query string of a request.  Any existing
// query string will be overwritten.
//
// The string is expected to be a valid, url encoded string.
func RawQuery(s string) func(*http.Request) error {
	return func(rq *http.Request) error {
		rq.URL.RawQuery = s
		return nil
	}
}
