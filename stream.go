package peerstream

import (
	"io"

	ss "github.com/docker/spdystream"
)

// StreamHandler is a function which receives a Stream. It
// allows clients to set a function to receive newly created
// streams. It works sort of like a http.HandleFunc.
type StreamHandler func(s Stream)

type Stream interface {
	io.Reader
	io.Writer
	io.Closer

	// Conn returns the Conn associated with this Stream
	Conn() *Conn

	// Swarm returns the Swarm asociated with this Stream
	Swarm() *Swarm
}

type stream struct {
	ssStream *ss.Stream

	conn   *Conn
	groups groupSet
}

func newStream(ssS *ss.Stream, c *Conn) *stream {
	s := &stream{conn: c, ssStream: ssS}
	s.groups.AddSet(&c.groups) // inherit groups
	return s
}

// Conn returns the Conn associated with this Stream
func (s *stream) Conn() *Conn {
	return s.conn
}

// Swarm returns the Swarm asociated with this Stream
func (s *stream) Swarm() *Swarm {
	return s.conn.swarm
}

func (s *stream) Read(p []byte) (n int, err error) {
	panic("nyi")
}

func (s *stream) Write(p []byte) (n int, err error) {
	panic("nyi")
}

func (s *stream) Close() error {
	panic("nyi")
}
