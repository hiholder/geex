package websocket

type msgReader struct {
	c    *Conn
	fin  bool
}

func newMsgReader(c *Conn) *msgReader {
	mr := &msgReader{
		c : c,
		fin: true,
	}
	return mr
}
