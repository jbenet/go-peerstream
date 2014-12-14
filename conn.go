package peerstream

import (
	"errors"
	"net"

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

	swarm   *Swarm
	streams map[fd]*stream
	groups  groupSet
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
