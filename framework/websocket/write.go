package websocket

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	gerrors "github.com/pkg/errors"
	"io"
	"time"
)

func (c *Conn) Writer(ctx context.Context, typ MessageType) (io.WriteCloser, error) {
	return c.writer(ctx, typ)
}

func (c *Conn) writer(ctx context.Context, typ MessageType) (io.WriteCloser, error) {
	err := c.msgWriterStats.reset(ctx, typ)
	if err != nil {
		return nil, err
	}
	return &msgWriter{
		mw: c.msgWriterStats,
		closed: false,
	}, nil
}

func (c *Conn) write(ctx context.Context, typ MessageType, p []byte) (int, error) {
	_, err := c.writer(ctx, typ)
	if err != nil {
		return 0, err
	}
	defer c.msgWriterStats.mu.unlock()
	return c.writeFrame(ctx, true, c.msgWriterStats.opcode, p)
}

type msgWriter struct {
	mw *msgWriterState
	closed bool
}

func (m *msgWriter) Write(p []byte) (n int, err error) {
	if m.closed {
		return 0, gerrors.New("cannot use closed writer")
	}
	return m.mw.Write(p)
}

func (m *msgWriter) Close() error {
	if m.closed {
		return gerrors.New("cannot use closed writer")
	}
	m.closed = true
	return m.mw.Close()
}

type msgWriterState struct {
	c    *Conn
	mu   *mu
	writeMu  *mu
	ctx  context.Context
	opcode opcode
	w    io.Writer
}

func newMsgWriterState(c *Conn) *msgWriterState {
	return &msgWriterState{
		c: c,
		mu: newMu(c),
		writeMu: newMu(c),
	}
}

func (mw *msgWriterState) Write(p []byte) (_ int, err error) {
	if err = mw.writeMu.lock(mw.ctx); err != nil {
		return 0, gerrors.Errorf("failed to write: %v", err)
	}
	defer mw.writeMu.unlock()
	defer func() {
		if err != nil {
			err = gerrors.Errorf("failed to write: %v", err)
			mw.c.close(err)
		}
	}()
	return mw.write(p)
}

func (mw *msgWriterState)write(p []byte) (int, error) {
	n, err := mw.c.writeFrame(mw.ctx, false, mw.opcode, p)
	if err != nil {
		return 0, gerrors.Errorf("failed to write data frame: %v", err)
	}
	mw.opcode = opContinuation
	return n, nil
}

func (mw *msgWriterState) Close() (err error) {
	if err = mw.writeMu.lock(mw.ctx); err != nil {
		return err
	}
	defer mw.writeMu.unlock()
	_, err = mw.c.writeFrame(mw.ctx, true, mw.opcode, nil)
	if err != nil {
		return gerrors.Errorf("failed to write fin frame: %v", err)
	}
	mw.mu.unlock()
	return nil
}

func (mw *msgWriterState) reset(ctx context.Context, typ MessageType) error {
	err := mw.mu.lock(ctx)
	if err != nil {
		return err
	}
	mw.ctx = ctx
	mw.opcode = opcode(typ)
	return nil
}


func (c *Conn) writeControl(ctx context.Context, opcode opcode, p []byte) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	_, err := c.writeFrame(ctx, true, opcode, p)
	if err != nil {
		return gerrors.Errorf("failed to write control frame %v: %v", opcode, err)
	}
	return nil
}

func (c *Conn) writeFrame(ctx context.Context, fin bool, opcode opcode, p []byte) (int, error) {
	if err := c.writeFrameMu.lock(ctx); err != nil {
		return 0, err
	}
	defer c.writeFrameMu.unlock()
	// 判断写关闭的状态
	c.closeMu.Lock()
	wroteClose := c.wroteClose
	c.closeMu.Unlock()
	// 写关闭状态且不是关闭操作，需要等待写完成
	if wroteClose && opcode != opClose {
		select {
		case <- ctx.Done():
			return 0, ctx.Err()
		case <-c.closed:
			return 0, c.closeErr
		}
	}
	select {
	case <- c.closed:
		return 0, c.closeErr
	case c.writeTimeout <- ctx:
	}
	c.writeHeader.fin = fin
	c.writeHeader.opcode = opcode
	c.writeHeader.payloadLength = int64(len(p))

	if c.client {
		c.writeHeader.masked = true
		_, err := io.ReadFull(rand.Reader, c.writeHeaderBuf[:4])
		if err != nil {
			return 0, gerrors.Wrapf(err, "failed to generate masking key")
		}
		c.writeHeader.maskKey = binary.LittleEndian.Uint32(c.writeHeaderBuf[:])
	}
	c.writeHeader.rsv1 = false
	err := writeFrameHeader(c.writeHeader, c.bw, c.writeHeaderBuf[:])
	if err != nil {
		return 0, err
	}
	n, err := c.writeFramePayload(p)
	if err != nil {
		return n, err
	}
	if c.writeHeader.fin {
		if err = c.bw.Flush(); err != nil {
			return 0, gerrors.Wrap(err, "failed to flush")
		}
	}
	select {
	case <-c.closed:
		return n, c.closeErr
	case c.writeTimeout <- context.Background():
	}
	return n, nil
}

func (c *Conn) writeFramePayload(p []byte) (n int, err error) {
	if !c.writeHeader.masked {
		return c.bw.Write(p)
	}
	maskKey := c.writeHeader.maskKey
	for len(p) > 0 {
		// 如果buffer已满，需要flush
		if c.bw.Available() == 0 {
			if err = c.bw.Flush(); err != nil {
				return n, err
			}
		}
		// 下一次写
		i := c.bw.Buffered()
		j := len(p)
		if j > c.bw.Available() {
			j = c.bw.Available()
		}
		if _, err := c.bw.Write(p[:j]); err != nil {
			return n, err
		}
		mask(maskKey, c.writeBuf[i:c.bw.Buffered()])
		p = p[j:]
		n += j
	}
	return n, nil
}

func (c *Conn) writeError(code StatusCode, err error) {
	c.setCloseErr(err)
	c.writeClose(code, err.Error())
	c.close(nil)
}