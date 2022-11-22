package bpool

import (
	"bytes"
	"sync"
)

var bpool sync.Pool

func Get() *bytes.Buffer {
	buf := bpool.Get()
	if buf == nil {
		return &bytes.Buffer{}
	}
	return buf.(*bytes.Buffer)
}

func Put(buf *bytes.Buffer) {
	buf.Reset()
	bpool.Put(buf)
}
