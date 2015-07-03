package peerstream_multiplex

import (
	"net"

	mp "github.com/jbenet/go-peerstream/Godeps/_workspace/src/github.com/whyrusleeping/go-multiplex" // Conn is a connection to a remote peer.
	pst "github.com/jbenet/go-peerstream/transport"
)

type conn struct {
	*mp.Multiplex
}

func (c *conn) Close() error {
	return c.Multiplex.Close()
}

func (c *conn) IsClosed() bool {
	return c.Multiplex.IsClosed()
}

// OpenStream creates a new stream.
func (c *conn) OpenStream() (pst.Stream, error) {
	return c.Multiplex.NewStream(), nil
}

// Serve starts listening for incoming requests and handles them
// using given StreamHandler
func (c *conn) Serve(handler pst.StreamHandler) {
	c.Multiplex.Serve(func(s *mp.Stream) {
		handler(s)
	})
}

// Transport is a go-peerstream transport that constructs
// multiplex-backed connections.
type Transport struct{}

// DefaultTransport has default settings for multiplex
var DefaultTransport = &Transport{}

func (t *Transport) NewConn(nc net.Conn, isServer bool) (pst.Conn, error) {
	return &conn{mp.NewMultiplex(nc, isServer)}, nil
}
