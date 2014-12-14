package peerstream

import (
	"io"
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
	Conn() Conn

	// Swarm returns the Swarm asociated with this Stream
	Swarm() Swarm

	groupable
}

type stream struct {
	groups groupable_
	conn   *conn

	// spdystream implementation details
	ssStream *ss.Stream
}

func newStream(c *conn, ssS *ss.Stream) *stream {
	s := &stream{conn: c, ssStream: ssS}
	s.groups.AddSet(c.groups) // inherit groups
	return s
}

func (s *stream) rawGroupable() groupable_ {
	return &s.grp
}
