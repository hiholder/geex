package websocket

import (
	"bufio"
	"errors"
	"github.com/smartystreets/goconvey/convey"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccept(t *testing.T) {
	t.Run("badClientHandshake", func(t *testing.T) {
		convey.Convey("", t, func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			_, err := Accept(w, r, nil)
			if err != nil {
				convey.So(err.Error(), convey.ShouldContainSubstring, "protocol violation")
			}
		})

	})
	t.Run("requireHttpHijacker", func(t *testing.T) {
		convey.Convey("", t, func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Set("Connection", "Upgrade")
			r.Header.Set("Upgrade", "websocket")
			r.Header.Set("Sec-WebSocket-Version", "13")
			r.Header.Set("Sec-WebSocket-Key", "meow123")

			_, err := Accept(w, r, nil)
			if err != nil {
				convey.So(err.Error(), convey.ShouldContainSubstring, "http.ResponseWriter does not implement http.Hijacker")
			}
		})
	})
	t.Run("badHijack", func(t *testing.T) {
		convey.Convey("badHijack", t, func() {
			w := mockHijacker{
				ResponseWriter: httptest.NewRecorder(),
				hijack: func() (net.Conn, *bufio.ReadWriter, error) {
					return nil, nil, errors.New("test")
				},
			}
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Set("Connection", "Upgrade")
			r.Header.Set("Upgrade", "websocket")
			r.Header.Set("Sec-WebSocket-Version", "13")
			r.Header.Set("Sec-WebSocket-Key", "meow123")
			_, err := Accept(w, r, nil)
			if err != nil {
				convey.So(err.Error(), convey.ShouldContainSubstring, "failed to hijack connection")
			}
		})
	})
}

type mockHijacker struct {
	http.ResponseWriter
	hijack func() (net.Conn, *bufio.ReadWriter, error)
}

var _ http.Hijacker = mockHijacker{}

func (mj mockHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return mj.hijack()
}
