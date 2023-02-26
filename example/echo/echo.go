package main

import (
	"context"
	"fmt"
	"github.com/hiholder/geex/framework"
	"github.com/hiholder/geex/framework/websocket"
	"github.com/sanity-io/litter"
	"log"
)

func EchoMessage(ctx *framework.Context)  {
	conn, err := websocket.Accept(ctx.Writer, ctx.Req, nil)
	if err != nil {
		log.Printf("")
		return
	}
	for {
		var v interface{}
		if err = websocket.Read(context.TODO(), conn, &v); err != nil {
			log.Printf("read err=%v", err)
			panic(err)
		}
		fmt.Printf("sent: %s\n", litter.Sdump(v))
	}
}