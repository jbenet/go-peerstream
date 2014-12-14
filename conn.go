package peerstream

import (
	"net"
)

// SelectConn selects a connection out of list. It allows
// delegation of decision making to clients. Clients can
// make SelectConn functons that check things connection
// qualities -- like latency andbandwidth -- or pick from
// a logical set of connections.
type SelectConn func([]Conn) Conn

// ErrInvalidConnSelected signals that a connection selected
// with a SelectConn function is invalid. This may be due to
// the Conn not being part of the original set given to the
// function, or the value being nil.
var ErrInvalidConnSelected = errors.New("invalid selected connection")

// ErrNoConnections signals that no connections are available
var ErrNoConnections = errors.New("no connections")

// Conn is a Swarm-associated connection.
type Conn interface {
	net.Conn

	// NetConn returns the underlying net.Conn
	NetConn() net.Conn

	// Swarm returns the Swarm associated with this Conn
	Swarm() Swarm

	// Stream returns a stream associated with this Conn
	NewStream() Stream

	groupable
}

type conn struct {
	grp   groupable_ // to give it group membership
	swarm *swarm

	netConn net.Conn // underlying connection
	streams map[fd]*stream

	// spdystream implementation details
	ssConn *ss.Connection
}

func (c *conn) Swarm() Swarm {
	return c.swarm
}

func (c *conn) NetConn() net.Conn {
	return c.netConn
}

func (c *conn) rawGroupable() groupable_ {
	return &c.grp
}

// SPDYConn returns the spdystream.Connection we use
// Warning: modifying this object is undefined.
func (c *conn) SPDYConn() *ss.Connection {
	return c.ssConn
}

// ConnsWithGroup narrows down a set of connections to
// those in a given group.
func ConnsWithGroup(g GroupID, conns []Conn) []Conn {
	var out []Conn
	for _, c := range conns {
		if g.Has(c.grp) {
			out = append(out, c)
		}
	}
	return out
}

func ConnInConns(c Conn, cs []Conn) bool {
	for _, c2 := range conns {
		if c2 == c1 {
			return true
		}
	}
	return false
}
