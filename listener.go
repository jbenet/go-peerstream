package peerstream

import (
	"net"
)

type Listener struct {
	netListener net.Listener
	groups      groupSet
}

// NetListener is the underlying net.Listener
func (l *Listener) NetListener() net.Listener {
	return l.netListener
}

// Groups returns the groups this Listener belongs to
func (l *Listener) Groups() []Group {
	return l.groups.Groups()
}
