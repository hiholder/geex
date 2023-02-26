package websocket

import (
	"context"
	"io"
	"net/http"
)

type DialOptions struct {

}

func Dial(ctx context.Context, urls string, opts *DialOptions) (*Conn, *http.Response, error) {
	return dial(ctx, urls, opts, nil)
}

func dial(ctx context.Context, urls string, opts *DialOptions, rand io.Reader) (*Conn, *http.Response, error) {
	return nil, nil, nil
}