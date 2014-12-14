package peerstream

import (
	"io"
	"math/rand"
)

var SelectRandomConn = func(conns []*Conn) *Conn {
	return conns[rand.Intn(len(conns))]
}

func EchoHandler(s Stream) {
	go func() {
		io.Copy(s, s)
		s.Close()
	}()
}

func CloseHandler(s Stream) {
	s.Close()
}
