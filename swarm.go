package peerstream

import (
	"errors"
	"net"
	"sync"
)

// fd is a (file) descriptor, unix style
type fd uint32

type Swarm struct {
	// active streams.
	streams    map[*stream]struct{}
	streamLock sync.RWMutex

	// active connections. generate new Streams
	conns    map[*Conn]struct{}
	connLock sync.RWMutex

	// active listeners. generate new Listeners
	listeners    map[*Listener]struct{}
	listenerLock sync.RWMutex

	// selectConn is the default SelectConn function
	selectConn   SelectConn
	selectConnLk sync.RWMutex

	// streamHandler receives Streams initiated remotely
	// should be accessed with SetStreamHandler / StreamHandler
	// as this pointer may be changed at any time.
	streamHandler   StreamHandler
	streamHandlerLk sync.RWMutex
}

func NewSwarm() *Swarm {
	return &Swarm{
		streams:       make(map[*stream]struct{}),
		conns:         make(map[*Conn]struct{}),
		listeners:     make(map[*Listener]struct{}),
		selectConn:    SelectRandomConn,
		streamHandler: CloseHandler,
	}
}

// SetStreamHandler assigns the stream handler in the swarm.
// The handler assumes responsibility for closing the stream.
// This need not happen at the end of the handler, leaving the
// stream open (to be used and closed later) is fine.
// It is also fine to keep a pointer to the Stream.
// If handler is nil, will use CloseHandler
// This is a threadsafe (atomic) operation
func (s *Swarm) SetStreamHandler(sh StreamHandler) {
	if sh == nil {
		sh = CloseHandler
	}
	s.streamHandlerLk.Lock()
	defer s.streamHandlerLk.Unlock()
	s.streamHandler = sh
	// atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&s.streamHandler)), unsafe.Pointer(&sh))
}

// StreamHandler returns the Swarm's current StreamHandler.
// This is a threadsafe (atomic) operation
func (s *Swarm) StreamHandler() StreamHandler {
	s.streamHandlerLk.RLock()
	defer s.streamHandlerLk.RUnlock()
	return s.streamHandler
	// p := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&s.streamHandler)))
	// return StreamHandler(*(*StreamHandler)(p))
}

// SetConnSelect assigns the connection selector in the swarm.
// If cs is nil, will use SelectRandomConn
// This is a threadsafe (atomic) operation
func (s *Swarm) SetSelectConn(cs SelectConn) {
	if cs == nil {
		cs = SelectRandomConn
	}
	s.selectConnLk.Lock()
	defer s.selectConnLk.Unlock()
	s.selectConn = cs
	// atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&s.selectConn)), unsafe.Pointer(&cs))
}

// ConnSelect returns the Swarm's current connection selector.
// ConnSelect is used in order to select the best of a set of
// possible connections. The default chooses one at random.
// This is a threadsafe (atomic) operation
func (s *Swarm) SelectConn() SelectConn {
	s.selectConnLk.RLock()
	defer s.selectConnLk.RUnlock()
	return s.selectConn
	// p := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&s.selectConn)))
	// return SelectConn(*(*SelectConn)(p))
}

// Conns returns all the connections associated with this Swarm.
func (s *Swarm) Conns() []*Conn {
	conns := make([]*Conn, 0, len(s.conns))
	for c := range s.conns {
		conns = append(conns, c)
	}
	return conns
}

// Listeners returns all the listeners associated with this Swarm.
func (s *Swarm) Listeners() []*Listener {
	out := make([]*Listener, 0, len(s.listeners))
	for c := range s.listeners {
		out = append(out, c)
	}
	return out
}

// Streams returns all the streams associated with this Swarm.
func (s *Swarm) Streams() []Stream {
	out := make([]Stream, 0, len(s.streams))
	for c := range s.streams {
		out = append(out, c)
	}
	return out
}

// AddListener adds net.Listener to the Swarm, and immediately begins
// accepting incoming connections.
func (s *Swarm) AddListener(l net.Listener) error {
	return s.addListener(l)
}

// AddListenerWithRateLimit adds Listener to the Swarm, and immediately
// begins accepting incoming connections. The rate of connection acceptance
// depends on the RateLimit option
// func (s *Swarm) AddListenerWithRateLimit(net.Listner, RateLimit) // TODO

// AddConn gives the Swarm ownership of net.Conn. The Swarm will open a
// SPDY session and begin listening for Streams.
// Returns the resulting Swarm-associated peerstream.Conn.
// Idempotent: if the Connection has already been added, this is a no-op.
func (s *Swarm) AddConn(netConn net.Conn) (*Conn, error) {
	return s.addConn(netConn, false)
}

// NewStream opens a new Stream on the best available connection,
// as selected by current swarm.SelectConn.
func (s *Swarm) NewStream() (Stream, error) {
	return s.NewStreamSelectConn(s.SelectConn())
}

func (s *Swarm) newStreamSelectConn(selConn SelectConn, conns []*Conn) (Stream, error) {
	if selConn == nil {
		return nil, errors.New("nil SelectConn")
	}

	best := selConn(conns)
	if best == nil || !ConnInConns(best, conns) {
		return nil, ErrInvalidConnSelected
	}
	return s.NewStreamWithConn(best)
}

// NewStreamWithSelectConn opens a new Stream on a connection selected
// by selConn.
func (s *Swarm) NewStreamSelectConn(selConn SelectConn) (Stream, error) {
	if selConn == nil {
		return nil, errors.New("nil SelectConn")
	}

	conns := s.Conns()
	if len(conns) == 0 {
		return nil, ErrNoConnections
	}
	return s.newStreamSelectConn(selConn, conns)
}

// NewStreamWithGroup opens a new Stream on an available connection in
// the given group. Uses the current swarm.SelectConn to pick between
// multiple connections.
func (s *Swarm) NewStreamWithGroup(group Group) (Stream, error) {
	conns := s.ConnsWithGroup(group)
	return s.newStreamSelectConn(s.SelectConn(), conns)
}

// NewStreamWithNetConn opens a new Stream on given net.Conn.
// Calls s.AddConn(netConn).
func (s *Swarm) NewStreamWithNetConn(netConn net.Conn) (Stream, error) {
	c, err := s.AddConn(netConn)
	if err != nil {
		return nil, err
	}
	return s.NewStreamWithConn(c)
}

// NewStreamWithConnection opens a new Stream on given connection.
func (s *Swarm) NewStreamWithConn(conn *Conn) (Stream, error) {
	if conn == nil {
		return nil, errors.New("nil Conn")
	}
	if conn.Swarm() != s {
		return nil, errors.New("connection not associated with swarm")
	}

	s.connLock.RLock()
	if _, found := s.conns[conn]; !found {
		s.connLock.RUnlock()
		return nil, errors.New("connection not associated with swarm")
	}
	s.connLock.RUnlock()
	return s.createStream(conn)
}

// AddConnToGroup assigns given Group to conn
func (s *Swarm) AddConnToGroup(conn *Conn, g Group) {
	conn.groups.Add(g)
}

// ConnsWithGroup returns all the connections with a given Group
func (s *Swarm) ConnsWithGroup(g Group) []*Conn {
	return ConnsWithGroup(g, s.Conns())
}

// Close shuts down the Swarm, and it's listeners.
func (s *Swarm) Close() error {
	// shut down TODO
	return nil
}
