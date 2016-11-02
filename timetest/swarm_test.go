package timeouttest

import (
	"io"
	"net"
	"testing"
	"time"

	ps "github.com/jbenet/go-peerstream"

	yamux "gx/ipfs/QmSHTSkxXGQgaHWz91oZV3CDy3hmKmDgpjbYRT6niACG4E/go-smux-yamux"
)

func TestConnTimeout(t *testing.T) {
	ps.NoStreamCloseTimeout = time.Second / 2
	ps.GarbageCollectTimeout = time.Second / 10
	tpt := yamux.DefaultTransport
	echo := func(s *ps.Stream) {
		defer s.Close()
		io.Copy(s, s)
	}

	swarm1 := ps.NewSwarm(tpt)
	swarm1.SetStreamHandler(echo)

	swarm2 := ps.NewSwarm(tpt)
	swarm2.SetStreamHandler(echo)

	list1, err := net.Listen("tcp", "0.0.0.0:7334")
	if err != nil {
		t.Fatal(err)
	}

	_, err = swarm1.AddListener(list1)
	if err != nil {
		t.Fatal(err)
	}

	c, err := net.Dial("tcp", "127.0.0.1:7334")
	if err != nil {
		t.Fatal(err)
	}

	psc, err := swarm2.AddConn(c)
	if err != nil {
		t.Fatal(err)
	}

	s, err := psc.NewStream()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)

	if len(swarm1.Conns()) != 1 {
		t.Fatal("should still have one connection")
	}

	if len(swarm2.Conns()) != 1 {
		t.Fatal("should still have one connection")
	}

	s.Close()

	time.Sleep(time.Second)
	if len(swarm1.Conns()) != 0 {
		t.Fatal("should have closed connection with no streams")
	}

	if len(swarm2.Conns()) != 0 {
		t.Fatal("should have closed connection with no streams")
	}
}
