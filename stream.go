package peerstream

import (
	"io"

	ss "github.com/docker/spdystream"
)

// StreamHandler is a function which receives a Stream. It
// allows clients to set a function to receive newly created
// streams, and decide whether to continue adding them.
// It works sort of like a http.HandleFunc.
// Note: the StreamHandler is called sequentially, so spawn
// goroutines or pass the Stream. See EchoHandler.
type StreamHandler func(s Stream)

type Stream interface {
	io.Reader
	io.Writer
	io.Closer

	// Conn returns the Conn associated with this Stream
	Conn() *Conn

	// Swarm returns the Swarm asociated with this Stream
	Swarm() *Swarm

	// Write writes bytes to a stream, calling write data for each call.
	Wait() error
}

type stream struct {
	ssStream *ss.Stream

	conn   *Conn
	groups groupSet
}

func newStream(ssS *ss.Stream, c *Conn) *stream {
	s := &stream{
		conn:     c,
		ssStream: ssS,
		groups:   groupSet{m: make(map[Group]struct{})},
	}
	s.groups.AddSet(&c.groups) // inherit groups
	return s
}

// SPDYStream returns the underlying *spdystream.Stream
func (s *stream) SPDYStream() *ss.Stream {
	return s.ssStream
}

// Conn returns the Conn associated with this Stream
func (s *stream) Conn() *Conn {
	return s.conn
}

// Swarm returns the Swarm asociated with this Stream
func (s *stream) Swarm() *Swarm {
	return s.conn.swarm
}

// Write writes bytes to a stream, calling write data for each call.
func (s *stream) Wait() error {
	return s.ssStream.Wait()
}

func (s *stream) Read(p []byte) (n int, err error) {
	return s.ssStream.Read(p)
}

func (s *stream) Write(p []byte) (n int, err error) {
	return s.ssStream.Write(p)
}

func (s *stream) Close() error {
	return s.conn.swarm.removeStream(s)
}
