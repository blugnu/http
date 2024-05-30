package http

import "net/http"

type (
	Client         = http.Client
	Request        = http.Request
	Response       = http.Response
	ResponseWriter = http.ResponseWriter
	RoundTripper   = http.RoundTripper
	Transport      = http.Transport
)

var (
	NoBody         = http.NoBody
	ListenAndServe = http.ListenAndServe
)

const (
	MethodConnect = http.MethodConnect
	MethodDelete  = http.MethodDelete
	MethodGet     = http.MethodGet
	MethodHead    = http.MethodHead
	MethodOptions = http.MethodOptions
	MethodPatch   = http.MethodPatch
	MethodPost    = http.MethodPost
	MethodPut     = http.MethodPut
	MethodTrace   = http.MethodTrace
)

const (
	StatusBadRequest          = http.StatusBadRequest
	StatusForbidden           = http.StatusForbidden
	StatusInternalServerError = http.StatusInternalServerError
	StatusNotAcceptable       = http.StatusNotAcceptable
	StatusNotFound            = http.StatusNotFound
	StatusOK                  = http.StatusOK
	StatusUnauthorized        = http.StatusUnauthorized
)
