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

func Write(ctx context.Context, c *Conn, v interface{}) error {
	return write(ctx, c, v)
}

func write(ctx context.Context, c *Conn, v interface{}) error {
	w, err := c.Writer(ctx, MessageText)
	if err != nil {
		return err
	}
	if err := jsoniter.NewEncoder(w).Encode(v); err != nil {
		return gerrors.Wrap(err, "marshal fail")
	}
	return w.Close()
}