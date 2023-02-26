package main

import (
	"context"
	"github.com/hiholder/geex/framework/websocket"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_echoServer(t *testing.T) {
	t.Parallel()

	s := httptest.NewServer(echoServer{
		logf: t.Logf,
	})
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c, _, err := websocket.Dial(ctx, s.URL, &websocket.DialOptions{
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	for i := 0; i < 5; i++ {
		err = websocket.Write(ctx, c, map[string]int{
			"i": i,
		})
		if err != nil {
			t.Fatal(err)
		}

		v := map[string]int{}
		err = websocket.Read(ctx, c, &v)
		if err != nil {
			t.Fatal(err)
		}

		if v["i"] != i {
			t.Fatalf("expected %v but got %v", i, v)
		}
	}

	c.Close(websocket.StatusNormalClosure, "")
}