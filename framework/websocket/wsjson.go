package websocket

import (
	"context"
	"github.com/hiholder/geex/framework/websocket/internal/bpool"
	"github.com/json-iterator/go"
	gerrors "github.com/pkg/errors"
)

func Read(ctx context.Context, c *Conn, v interface{}) error {
	return read(ctx, c, v)
}

func read(ctx context.Context, c *Conn, v interface{}) (err error) {
	_, reader, err := c.Reader(ctx)
	if err != nil {
		return err
	}
	bf := bpool.Get()
	defer bpool.Put(bf)
	_, err = bf.ReadFrom(reader)
	if err != nil {
		return err
	}
	if err = jsoniter.Unmarshal(bf.Bytes(), v); err != nil {
		return gerrors.Wrap(err, "unmarshal fail")
	}

	return nil
}