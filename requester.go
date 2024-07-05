package http

import (
	"context"

	"github.com/blugnu/errorcontext"
	"github.com/blugnu/http/request"
)

type Returning[T any] struct {
	HttpClient
}

func (r Returning[T]) Get(ctx context.Context, path string, opts ...func(*Request) error) (*T, error) {
	opts = append([]func(*Request) error{request.AcceptJSON()}, opts...)
	resp, err := r.HttpClient.Get(ctx, path, opts...)
	if err != nil {
		return nil, errorcontext.Wrap(ctx, err)
	}

	result, err := UnmarshalJSON[T](resp)
	if err != nil {
		return nil, errorcontext.Wrap(ctx, err)
	}
	return result, nil
}

// func Post[RQ any, R any](ctx context.Context, h HttpClient, path string, body *RQ, opts ...func(*Request) error) (*R, error) {
// 	opts = append([]func(*Request) error{request.AcceptJSON()}, opts...)
// 	resp, err := h.Get(ctx, path, opts...)
// 	if err != nil {
// 		return nil, errorcontext.Wrap(ctx, err)
// 	}

// 	result, err := UnmarshalJSON[R](resp)
// 	if err != nil {
// 		return nil, errorcontext.Wrap(ctx, err)
// 	}
// 	return result, nil
// }
