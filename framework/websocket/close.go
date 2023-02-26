package websocket

import (
	"context"
	"encoding/binary"
	"errors"
	gerrors "github.com/pkg/errors"
	"log"
)

type StatusCode int

// https://www.rfc-editor.org/rfc/rfc6455.html#section-7.4.1
const (
	StatusNormalClosure   StatusCode = 1000
	StatusGoingAway       StatusCode = 1001
	StatusProtocolError   StatusCode = 1002
	StatusUnsupportedData StatusCode = 1003

	// 1004 is reserved and so unexported.
	statusReserved StatusCode = 1004

	// StatusNoStatusRcvd cannot be sent in a close message.
	// It is reserved for when a close message is received without
	// a status code.
	StatusNoStatusRcvd StatusCode = 1005

	// StatusAbnormalClosure is exported for use only with Wasm.
	// In non Wasm Go, the returned error will indicate whether the
	// connection was closed abnormally.
	StatusAbnormalClosure StatusCode = 1006

	StatusInvalidFramePayloadData StatusCode = 1007
	StatusPolicyViolation         StatusCode = 1008
	StatusMessageTooBig           StatusCode = 1009
	StatusMandatoryExtension      StatusCode = 1010
	StatusInternalError           StatusCode = 1011
	StatusServiceRestart          StatusCode = 1012
	StatusTryAgainLater           StatusCode = 1013
	StatusBadGateway              StatusCode = 1014

	// StatusTLSHandshake is only exported for use with Wasm.
	// In non Wasm Go, the returned error will indicate whether there was
	// a TLS handshake failure.
	StatusTLSHandshake StatusCode = 1015
)

type CloseError struct {
	Code   StatusCode // 错误码
	Reason string // 连接关闭原因
}

func (c *Conn) Close(code StatusCode, reason string) error {
	return c.closeHandshake(code, reason)
}

func (c *Conn) closeHandshake(code StatusCode, reason string) (err error) {
	writeErr := c.writeClose(code, reason)
	closeHandshakeErr := c.waitCloseHandshake()
	if writeErr != nil {
		return writeErr
	}
	if CloseStatus(closeHandshakeErr) != -1 {
		return closeHandshakeErr
	}
	return nil
}

func (c *Conn) isClosed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

func (c *Conn) setCloseErrLocked(err error) {
	if c.closeErr == nil {
		c.closeErr = gerrors.Errorf("WebSocket closed: %v", err)
	}
}

func (c *Conn) setCloseErr(err error)  {
	c.closeMu.Lock()
	c.setCloseErrLocked(err)
	c.closeMu.Unlock()
}

var errAlreadyWroteClose = gerrors.New("already wrote close")

// 将关闭信息写入返回结构中
func (c *Conn) writeClose(code StatusCode, reason string) error {
	c.closeMu.Lock()
	wroteClose := c.wroteClose
	c.wroteClose = true
	c.closeMu.Unlock()
	if wroteClose {
		return errAlreadyWroteClose
	}
	ce := &CloseError{
		Code: code,
		Reason: reason,
	}
	var p []byte
	var marshalErr error
	if ce.Code != StatusNoStatusRcvd {
		if p, marshalErr = ce.bytesErr(); marshalErr != nil {
			log.Printf("websocket: %v", marshalErr)
		}
	}
	writeErr := c.writeControl(context.Background(), opClose, p)
	if CloseStatus(writeErr) != -1 {
		// 不是一个真正的错误，只是接收到了关闭帧
		writeErr = nil
	}
	c.setCloseErr(gerrors.Errorf("sent close frame: %v", ce))
	if marshalErr != nil {
		return marshalErr
	}
	return nil
}

func (c *Conn) waitCloseHandshake() error {
	// todo: 待实现
	return nil
}

func parseClosePayload(p []byte) (CloseError, error) {
	if len(p) == 0 {
		return CloseError{
			Code: StatusNoStatusRcvd,
		}, nil
	}
	ce := CloseError{
		Code: StatusCode(binary.BigEndian.Uint32(p)),
		Reason: string(p[:2]),
	}
	return ce, nil
}

const maxCloseReason = maxControlPayload - 2

func (ce CloseError) bytesErr() ([]byte, error) {
	if len(ce.Reason) > maxCloseReason {
		return nil, gerrors.Errorf("reason length=%v beyond, reason=%v", len(ce.Reason), ce.Reason)
	}
	buf := make([]byte, 2+len(ce.Reason))
	binary.BigEndian.PutUint16(buf, uint16(ce.Code))
	copy(buf[:2], ce.Reason)
	return buf, nil
}

func CloseStatus(err error) StatusCode {
	var ce CloseError
	if errors.As(err, &ce) {
		return ce.Code
	}
	return -1
}