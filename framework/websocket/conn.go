package websocket

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync"
)

type Conn struct {
	rwc          io.ReadWriteCloser
	br           *bufio.Reader
	bw           *bufio.Writer
	client       bool
	readTimeout  chan context.Context
	writeTimeout chan context.Context
	// read
	readMu            *mu
	readCloseFrameErr error
	msgReader         *msgReader
	// write
	writeFrameMu *mu
	writeHeader  header

	closed     chan struct{}
	closeMu    sync.Mutex
	closeErr   error
	wroteClose bool
}

type connConfig struct {
	subProtocol    string
	rwc    io.ReadWriteCloser
	br     *bufio.Reader
	bw     *bufio.Writer
	client bool // 是否是客户端
}

type mu struct {
	c  *Conn
	ch chan struct{}
}

func newMu(c *Conn) *mu {
	return &mu{
		c:    c,
		ch : make(chan struct{}, 1),
	}
}

func newConn(cfg connConfig) *Conn {
	c := &Conn{
		rwc:    cfg.rwc,
		client: cfg.client,
		br:     cfg.br,
		bw:     cfg.bw,

		readTimeout:  make(chan context.Context),
		writeTimeout: make(chan context.Context),
		closed:       make(chan struct{}),
	}

	c.readMu = newMu(c)
	c.writeFrameMu = newMu(c)

	c.msgReader = newMsgReader(c)
	go c.timeoutLoop()
	return c
}

func (c *Conn) timeoutLoop()  {
	readCtx := context.Background()
	writeCtx := context.Background()
	for {
		select {
		case <- c.closed: // 连接关闭
			return
		case writeCtx = <- c.writeTimeout:
		case readCtx = <- c.readTimeout:
		case <-readCtx.Done():
			c.close(fmt.Errorf("read timed out: %w", readCtx.Err()))
		case <-writeCtx.Done():
			c.close(fmt.Errorf("write timed out: %w", writeCtx.Err()))
			return
		}
	}
}

func (c *Conn) close(err error)  {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()
	if c.isClosed() {
		return
	}
	c.setCloseErrLocked(err)
	close(c.closed)
	// 保证底层的连接关闭
	c.rwc.Close()
	// TODO：还需要关闭msgReader和msgWriter
}

func (c *Conn) setCloseErrLocked(err error)  {
	if c.closeErr == nil {
		c.closeErr = fmt.Errorf("WebSocket closed: %w", err)
	}
}

func (c *Conn) isClosed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}
