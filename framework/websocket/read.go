package websocket

import (
	"context"
	"errors"
	"fmt"
	gerrors "github.com/pkg/errors"
	"io"
	"time"
)

// MessageType represents the type of a WebSocket message.
// See https://tools.ietf.org/html/rfc6455#section-5.6
type MessageType int

// MessageType constants.
const (
	// MessageText is for UTF-8 encoded text messages like JSON.
	MessageText MessageType = iota + 1
	// MessageBinary is for binary messages like protobufs.
	MessageBinary
)


func newMsgReader(c *Conn) *msgReader {
	mr := &msgReader{
		c:   c,
		fin: true,

	}
	mr.readerFunc = mr.read
	mr.r = mr.readerFunc
	return mr
}

// Reader 返回io.Reader结构
func (c *Conn) Reader(ctx context.Context) (MessageType, io.Reader, error) {
	return c.reader(ctx)
}

func (c *Conn) reader(ctx context.Context) (_ MessageType, _ io.Reader, err error) {
	if err = c.readMu.lock(ctx); err != nil {
		return 0, nil, err
	}
	if !c.msgReader.fin {
		err = errors.New("previous message not read to completion")
		c.close(fmt.Errorf("fail to get reader: %w", err))
		return 0, nil, err
	}
	defer c.readMu.unlock()
	// 读取协议头
	h, err := c.readLoop(ctx)
	if err != nil {
		return 0, nil, err
	}
	// 状态码-0
	if h.opcode == opContinuation {
		err := gerrors.New("received continuation frame without text or binary frame")
		c.writeError(StatusProtocolError, err)
		return 0, nil, err
	}
	c.msgReader.reset(ctx, h)
	return MessageType(h.opcode), c.msgReader, err
}

func (c *Conn) readLoop(ctx context.Context) (header, error) {
	for {
		h, err := c.readFrameHeader(ctx)
		if err != nil {
			return header{}, err
		}
		if !c.client && !h.masked {
			return header{}, errors.New("received unmasked frame from client")
		}
		switch h.opcode {
		case opClose, opPing, opPong:
			err = c.readControl(ctx, h)
			if err != nil {
				if h.opcode == opClose && CloseStatus(err) != -1 {
					return header{}, err
				}
				return header{}, gerrors.Errorf("failed to handle control frame %v: %v", h.opcode, err)
			}
		case opContinuation, opText, opBinary:
			return h, nil
		default:
			err := fmt.Errorf("received unknown opcode %v", h.opcode)
			c.close(err)
			return header{}, err
		}
	}
}

func (c *Conn) readFrameHeader(ctx context.Context) (header, error) {
	select {
	case <- c.closed:
		return header{}, c.closeErr
	case c.readTimeout <- ctx:

	}
	// frame方法
	h, err := readFrameHeader(c.br, c.readHeaderBuf[:])
	if err != nil {
		select {
		case <-c.closed:
			return header{}, c.closeErr
		case <- ctx.Done():
			return header{}, ctx.Err()
		default:
			c.close(err)
			return header{}, err
		}
	}
	select {
	case <- c.closed:
		return header{}, c.closeErr
	case c.readTimeout <- context.Background():
	}
	return h, nil
}

func (c *Conn) readFramePayload(ctx context.Context, p []byte) (int, error) {
	select {
	case <-c.closed:
		return 0, c.closeErr
	case c.readTimeout <- ctx:
	}
	n, err := io.ReadFull(c.br, p)
	if err != nil {
		select {
		case <- c.closed:
			return n, c.closeErr
		case <-ctx.Done():
			return n, ctx.Err()
		default:
			err = gerrors.Errorf("failed to read frame payload: %v", err)
			c.close(err)
			return n, err
		}
	}
	select {
	case <-c.closed:
		return n, c.closeErr
	case c.readTimeout <- context.Background():
	}
	return n, err
}

// 控制帧数据处理
func (c *Conn) readControl(ctx context.Context, h header) (err error) {
	if h.payloadLength < 0 || h.payloadLength > maxControlPayload {
		err := gerrors.Errorf("received control frame payload with invalid length: %v", h.payloadLength)
		return err
	}
	if !h.fin {
		err := gerrors.New("received fragmented control frame")
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	b := c.readControlBuf[:h.payloadLength]
	_, err = c.readFramePayload(ctx, b)
	if err != nil {
		return err
	}
	switch h.opcode {
	case opPing:
		return c.writeControl(ctx, opPong, b)
	case opPong:
		pong, ok := c.activePings[string(b)]
		if ok {
			select {
			case pong <- struct{}{}:
			default:
			}
		}
		return nil
	}
	// 处理关闭帧
	defer func() {
		c.readCloseFrameErr = err
	}()
	ce, err := parseClosePayload(b)
	if err != nil {
		err = gerrors.Wrapf(err, "received invalid close payload")
		c.writeError(StatusProtocolError, err)
		return err
	}
	err = gerrors.Errorf("received close frame: %v", ce)
	c.setCloseErr(err)
	c.writeClose(ce.Code, ce.Reason)
	c.close(err)
	return err
}

type msgReader struct {
	c *Conn

	ctx           context.Context
	fin           bool
	payloadLength int64
	maskKey       uint32

	// 读取数据
	r io.Reader
	n int64

	readerFunc readerFunc
}

func (mr *msgReader) reset(ctx context.Context, h header)  {
	mr.ctx = ctx
	mr.r = mr.readerFunc
	mr.setFrame(h)
}

func (mr *msgReader) setFrame(h header)  {
	mr.fin = h.fin
	mr.payloadLength = h.payloadLength
	mr.maskKey = h.maskKey
}

func (mr *msgReader) Read(p []byte) (int, error) {
	if err := mr.c.readMu.lock(mr.ctx); err != nil {
		return 0, gerrors.Errorf("failed to read: %v", err)
	}
	defer mr.c.readMu.unlock()
	// 实际读取数据
	if mr.n <= 0 {
		err := gerrors.Errorf("read limited at %v bytes")
		return 0, err
	}
	if int64(len(p)) > mr.n {
		p = p[:mr.n]
	}
	n, err := mr.r.Read(p)
	mr.n -= int64(n)
	if err != nil {
		err = gerrors.Errorf("fail to read: %v", err)
		mr.c.close(err)
	}
	return n, err
}

func (mr *msgReader) read(p []byte) (int, error) {
	for {
		if mr.payloadLength == 0 {
			h, err := mr.c.readLoop(mr.ctx)
			if err != nil {
				return 0, err
			}
			if h.opcode != opContinuation {
				err := gerrors.New("received new data message without finishing the previous message")
				mr.c.writeError(StatusProtocolError, err)
				return 0, err
			}
			mr.setFrame(h)
			continue
		}
		if int64(len(p)) > mr.payloadLength {
			p = p[:mr.payloadLength]
		}
		n, err := mr.c.readFramePayload(mr.ctx, p)
		if err != nil {
			return n, err
		}
		mr.payloadLength -= int64(n)
		if !mr.c.client {
			mr.maskKey = mask(mr.maskKey, p)
		}
		return n, nil
	}
}

type readerFunc func(p []byte) (int, error)

func (f readerFunc) Read(p []byte) (int, error) {
	return f(p)
}