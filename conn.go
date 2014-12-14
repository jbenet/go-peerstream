package peerstream

import (
	"errors"
	"net"
	"net/http"
	"sync"

	ss "github.com/docker/spdystream"
)

// SelectConn selects a connection out of list. It allows
// delegation of decision making to clients. Clients can
// make SelectConn functons that check things connection
// qualities -- like latency andbandwidth -- or pick from
// a logical set of connections.
type SelectConn func([]*Conn) *Conn

// ErrInvalidConnSelected signals that a connection selected
// with a SelectConn function is invalid. This may be due to
// the Conn not being part of the original set given to the
// function, or the value being nil.
var ErrInvalidConnSelected = errors.New("invalid selected connection")

// ErrNoConnections signals that no connections are available
var ErrNoConnections = errors.New("no connections")

// Conn is a Swarm-associated connection.
type Conn struct {
	ssConn  *ss.Connection
	netConn net.Conn // underlying connection

	swarm  *Swarm
	groups groupSet

	streams    map[*stream]struct{}
	streamLock sync.RWMutex
}

func newConn(nconn net.Conn, sconn *ss.Connection, s *Swarm) *Conn {
	return &Conn{
		netConn: nconn,
		ssConn:  sconn,
		swarm:   s,
		streams: map[*stream]struct{}{},
	}
}

// Swarm returns the Swarm associated with this Conn
func (c *Conn) Swarm() *Swarm {
	return c.swarm
}

// NetConn returns the underlying net.Conn
func (c *Conn) NetConn() net.Conn {
	return c.netConn
}

// SPDYConn returns the spdystream.Connection we use
// Warning: modifying this object is undefined.
func (c *Conn) SPDYConn() *ss.Connection {
	return c.ssConn
}

// Stream returns a stream associated with this Conn
func (c *Conn) NewStream() (Stream, error) {
	return c.swarm.NewStreamWithConn(c)
}

// ConnsWithGroup narrows down a set of connections to
// those in a given group.
func ConnsWithGroup(g Group, conns []*Conn) []*Conn {
	var out []*Conn
	for _, c := range conns {
		if c.groups.Has(g) {
			out = append(out, c)
		}
	}
	return out
}

func ConnInConns(c1 *Conn, conns []*Conn) bool {
	for _, c2 := range conns {
		if c2 == c1 {
			return true
		}
	}
	return false
}

// ------------------------------------------------------------------
// All the connection setup logic here, in one place.
// these are mostly *Swarm methods, but i wanted a less-crowded place
// for them.
// ------------------------------------------------------------------

// addConn is the internal version of AddConn. we need the server bool
// as spdystream requires it.
func (s *Swarm) addConn(netConn net.Conn, server bool) (*Conn, error) {
	if netConn == nil {
		return nil, errors.New("nil conn")
	}

	s.connLock.Lock()
	defer s.connLock.Unlock()

	// first, check if we already have it...
	for c := range s.conns {
		if c.netConn == netConn {
			return c, nil
		}
	}

	// create a new spdystream connection
	ssConn, err := ss.NewConnection(netConn, server)
	if err != nil {
		return nil, err
	}

	// add the connection
	c := newConn(netConn, ssConn, s)
	s.conns[c] = struct{}{}

	// go listen for incoming streams on this connection
	go c.ssConn.Serve(func(ssS *ss.Stream) {
		stream := s.setupSSStream(ssS, c)
		s.StreamHandler()(stream) // call our handler
	})

	return c, nil
}

// createStream is the internal function that creates a new stream. assumes
// all validation has happened.
func (s *Swarm) createStream(c *Conn) (*stream, error) {

	// Create a new ss.Stream
	ssStream, err := c.ssConn.CreateStream(http.Header{}, nil, false)
	if err != nil {
		return nil, err
	}

	// create a new stream
	return s.setupSSStream(ssStream, c), nil
}

// newStream is the internal function that creates a new stream. assumes
// all validation has happened.
func (s *Swarm) setupSSStream(ssS *ss.Stream, c *Conn) *stream {
	// create a new *stream
	stream := newStream(ssS, c)

	// add it to our streams maps

	// add it to our map
	s.streamLock.Lock()
	c.streamLock.Lock()
	s.streams[stream] = struct{}{}
	c.streams[stream] = struct{}{}
	s.streamLock.Unlock()
	c.streamLock.Unlock()
	return stream
}

func (s *Swarm) removeStream(stream *stream) error {

	// remove from our maps
	s.streamLock.Lock()
	stream.conn.streamLock.Lock()
	delete(s.streams, stream)
	delete(stream.conn.streams, stream)
	s.streamLock.Unlock()
	stream.conn.streamLock.Unlock()

	return stream.ssStream.Close()
}
