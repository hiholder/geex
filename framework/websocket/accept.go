package websocket

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strings"
)

type AcceptOptions struct {

}

func Accept(w http.ResponseWriter, r *http.Request, opts *AcceptOptions) (*Conn, error) {
	return accept(w, r, opts)
}

func accept(w http.ResponseWriter, r *http.Request, opts *AcceptOptions) (_ *Conn, err error) {
	if opts == nil {
		opts = &AcceptOptions{}
	}
	errCode, err := verifyClientRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), errCode)
		return nil, err
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		err = errors.New("http.ResponseWriter does not implement http.Hijacker")
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return nil, err
	}
	netConn, brw, err := hj.Hijack()
	if err != nil {
		err = fmt.Errorf("failed to hijack connection: %w", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return nil, err
	}
	// 设置响应参数
	upgradeResponse(w, r)
	// https://github.com/golang/go/issues/32314
	b, _ := brw.Reader.Peek(brw.Reader.Buffered())
	brw.Reader.Reset(io.MultiReader(bytes.NewReader(b), netConn))
	return newConn(connConfig{
		subProtocol: w.Header().Get("Sec-WebSocket-Protocol"),
		rwc: netConn,
		client: false,
		br: brw.Reader,
		bw: brw.Writer,
	}), nil
}

func verifyClientRequest(w http.ResponseWriter, r *http.Request) (errCode int, _ error) {
	if !r.ProtoAtLeast(1, 1) {
		return http.StatusUpgradeRequired, fmt.Errorf("WebSocket protocol violation: handshake request must be at least HTTP/1.1: %q", r.Proto)
	}

	if !headerContainsTokenIgnoreCase(r.Header, "Connection", "Upgrade") {
		w.Header().Set("Connection", "Upgrade")
		w.Header().Set("Upgrade", "websocket")
		return http.StatusUpgradeRequired, fmt.Errorf("WebSocket protocol violation: Connection header %q does not contain Upgrade", r.Header.Get("Connection"))
	}

	if !headerContainsTokenIgnoreCase(r.Header, "Upgrade", "websocket") {
		w.Header().Set("Connection", "Upgrade")
		w.Header().Set("Upgrade", "websocket")
		return http.StatusUpgradeRequired, fmt.Errorf("WebSocket protocol violation: Upgrade header %q does not contain websocket", r.Header.Get("Upgrade"))
	}

	if r.Method != "GET" {
		return http.StatusMethodNotAllowed, fmt.Errorf("WebSocket protocol violation: handshake request method is not GET but %q", r.Method)
	}

	if r.Header.Get("Sec-WebSocket-Version") != "13" {
		w.Header().Set("Sec-WebSocket-Version", "13")
		return http.StatusBadRequest, fmt.Errorf("unsupported WebSocket protocol version (only 13 is supported): %q", r.Header.Get("Sec-WebSocket-Version"))
	}

	if r.Header.Get("Sec-WebSocket-Key") == "" {
		return http.StatusBadRequest, errors.New("WebSocket protocol violation: missing Sec-WebSocket-Key")
	}

	return 0, nil
}

func upgradeResponse(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Upgrade", "websocket")
	w.Header().Set("Connection", "Upgrade")

	key := r.Header.Get("Sec-WebSocket-Key")
	w.Header().Set("Sec-WebSocket-Accept", secWebSocketAccept(key))
}

func headerContainsTokenIgnoreCase(h http.Header, key, token string) bool {
	for _, t := range headerTokens(h, key) {
		if strings.EqualFold(t, token) {
			return true
		}
	}
	return false
}

func headerTokens(h http.Header, key string) []string {
	key = textproto.CanonicalMIMEHeaderKey(key)
	var tokens []string
	for _, v := range h[key] {
		v = strings.TrimSpace(v)
		for _, t := range strings.Split(v, ",") {
			t = strings.TrimSpace(t)
			tokens = append(tokens, t)
		}
	}
	return tokens
}

var keyGUID = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
func secWebSocketAccept(secWebSocketKey string) string {
	h := sha1.New()
	h.Write([]byte(secWebSocketKey))
	h.Write(keyGUID)

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}