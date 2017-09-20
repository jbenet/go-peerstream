package muxtest

import (
	"bytes"
	crand "crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"
	"os"
	"reflect"
	"runtime"
	"sync"
	"testing"

	ps "github.com/libp2p/go-peerstream"

	smux "github.com/jbenet/go-stream-muxer"
	tcpc "github.com/libp2p/go-tcp-transport"
	ma "github.com/multiformats/go-multiaddr"
)

var zeroaddr = ma.StringCast("/ip4/0.0.0.0/tcp/0")

var randomness []byte
var nextPort = 20000
var verbose = false

func init() {
	// read 1MB of randomness
	randomness = make([]byte, 1<<20)
	if _, err := crand.Read(randomness); err != nil {
		panic(err)
	}
}

func randBuf(size int) []byte {
	n := len(randomness) - size
	if size < 1 {
		panic(fmt.Errorf("requested too large buffer (%d). max is %d", size, len(randomness)))
	}

	start := mrand.Intn(n)
	return randomness[start : start+size]
}

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func log(s string, v ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stderr, "> "+s+"\n", v...)
	}
}

type echoSetup struct {
	swarm *ps.Swarm
	conns []*ps.Conn
}

func singleConn(t *testing.T, tr smux.Transport) echoSetup {
	tcp := tcpc.NewTCPTransport()
	swarm := ps.NewSwarm(tr)
	swarm.SetStreamHandler(func(s *ps.Stream) {
		defer s.Close()
		log("accepted stream")
		io.Copy(s, s) // echo everything
		log("closing stream")
	})

	log("listening at %s", "localhost:0")
	d, err := tcp.Dialer(zeroaddr)
	checkErr(t, err)
	l, err := tcp.Listen(ma.StringCast("/ip4/0.0.0.0/tcp/0"))
	checkErr(t, err)

	_, err = swarm.AddListener(l)
	checkErr(t, err)

	log("dialing to %s", l.Multiaddr())
	nc1, err := d.Dial(l.Multiaddr())
	checkErr(t, err)

	c1, err := swarm.AddConn(nc1)
	checkErr(t, err)

	return echoSetup{
		swarm: swarm,
		conns: []*ps.Conn{c1},
	}
}

func makeSwarm(t *testing.T, tr smux.Transport, nListeners int) *ps.Swarm {
	tcp := tcpc.NewTCPTransport()
	swarm := ps.NewSwarm(tr)
	swarm.SetStreamHandler(func(s *ps.Stream) {
		defer s.Close()
		log("accepted stream")
		io.Copy(s, s) // echo everything
		log("closing stream")
	})

	for i := 0; i < nListeners; i++ {
		log("%p listening at %s", swarm, "localhost:0")
		l, err := tcp.Listen(ma.StringCast("/ip4/0.0.0.0/tcp/0"))
		checkErr(t, err)
		_, err = swarm.AddListener(l)
		checkErr(t, err)
	}

	return swarm
}

func makeSwarms(t *testing.T, tr smux.Transport, nSwarms, nListeners int) []*ps.Swarm {
	swarms := make([]*ps.Swarm, nSwarms)
	for i := 0; i < nSwarms; i++ {
		swarms[i] = makeSwarm(t, tr, nListeners)
	}
	return swarms
}

func SubtestConstructSwarm(t *testing.T, tr smux.Transport) {
	ps.NewSwarm(tr)
}

func SubtestSimpleWrite(t *testing.T, tr smux.Transport) {
	tcp := tcpc.NewTCPTransport()
	swarm := ps.NewSwarm(tr)
	defer swarm.Close()

	piper, pipew := io.Pipe()
	swarm.SetStreamHandler(func(s *ps.Stream) {
		defer s.Close()
		log("accepted stream")
		w := io.MultiWriter(s, pipew)
		io.Copy(w, s) // echo everything and write it to pipew
		log("closing stream")
	})

	log("listening at %s", "localhost:0")
	l, err := tcp.Listen(ma.StringCast("/ip4/0.0.0.0/tcp/0"))
	checkErr(t, err)

	_, err = swarm.AddListener(l)
	checkErr(t, err)

	d, err := tcp.Dialer(zeroaddr)
	checkErr(t, err)

	log("dialing to %s", l.Addr().String())
	nc1, err := d.Dial(l.Multiaddr())
	checkErr(t, err)

	c1, err := swarm.AddConn(nc1)
	checkErr(t, err)
	defer c1.Close()

	log("creating stream")
	s1, err := c1.NewStream()
	checkErr(t, err)
	defer s1.Close()

	buf1 := randBuf(4096)
	log("writing %d bytes to stream", len(buf1))
	_, err = s1.Write(buf1)
	checkErr(t, err)

	buf2 := make([]byte, len(buf1))
	log("reading %d bytes from stream (echoed)", len(buf2))
	_, err = s1.Read(buf2)
	checkErr(t, err)
	if string(buf2) != string(buf1) {
		t.Error("buf1 and buf2 not equal: %s != %s", string(buf1), string(buf2))
	}

	buf3 := make([]byte, len(buf1))
	log("reading %d bytes from pipe (tee)", len(buf3))
	_, err = piper.Read(buf3)
	checkErr(t, err)
	if string(buf3) != string(buf1) {
		t.Error("buf1 and buf3 not equal: %s != %s", string(buf1), string(buf3))
	}
}

func SubtestSimpleWrite100msgs(t *testing.T, tr smux.Transport) {

	msgs := 100
	msgsize := 1 << 19
	es := singleConn(t, tr)
	defer es.swarm.Close()

	log("creating stream")
	stream, err := es.conns[0].NewStream()
	checkErr(t, err)

	bufs := make(chan []byte, msgs)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for i := 0; i < msgs; i++ {
			buf := randBuf(msgsize)
			bufs <- buf
			log("writing %d bytes (message %d/%d #%x)", len(buf), i, msgs, buf[:3])
			if _, err := stream.Write(buf); err != nil {
				t.Error(fmt.Errorf("stream.Write(buf): %s", err))
				continue
			}
		}
		close(bufs)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		buf2 := make([]byte, msgsize)
		i := 0
		for buf1 := range bufs {
			log("reading %d bytes (message %d/%d #%x)", len(buf1), i, msgs, buf1[:3])
			i++

			if _, err := io.ReadFull(stream, buf2); err != nil {
				t.Error(fmt.Errorf("readFull(stream, buf2): %s", err))
				continue
			}
			if !bytes.Equal(buf1, buf2) {
				t.Error(fmt.Errorf("buffers not equal (%x != %x)", buf1[:3], buf2[:3]))
			}
		}
	}()

	wg.Wait()
}

func SubtestStressNSwarmNConnNStreamNMsg(t *testing.T, tr smux.Transport, nSwarm, nConn, nStream, nMsg int) {

	msgsize := 1 << 11

	rateLimitN := 5000
	rateLimitChan := make(chan struct{}, rateLimitN) // max of 5k funcs.
	for i := 0; i < rateLimitN; i++ {
		rateLimitChan <- struct{}{}
	}

	rateLimit := func(f func()) {
		<-rateLimitChan
		f()
		rateLimitChan <- struct{}{}
	}

	writeStream := func(s *ps.Stream, bufs chan<- []byte) {
		log("writeStream %p, %d nMsg", s, nMsg)

		for i := 0; i < nMsg; i++ {
			buf := randBuf(msgsize)
			bufs <- buf
			log("%p writing %d bytes (message %d/%d #%x)", s, len(buf), i, nMsg, buf[:3])
			if _, err := s.Write(buf); err != nil {
				t.Error(fmt.Errorf("s.Write(buf): %s", err))
				continue
			}
		}
	}

	readStream := func(s *ps.Stream, bufs <-chan []byte) {
		log("readStream %p, %d nMsg", s, nMsg)

		buf2 := make([]byte, msgsize)
		i := 0
		for buf1 := range bufs {
			i++
			log("%p reading %d bytes (message %d/%d #%x)", s, len(buf1), i-1, nMsg, buf1[:3])

			if _, err := io.ReadFull(s, buf2); err != nil {
				log("%p failed to read %d bytes (message %d/%d #%x)", s, len(buf1), i-1, nMsg, buf1[:3])
				t.Error(fmt.Errorf("io.ReadFull(s, buf2): %s", err))
				continue
			}
			if !bytes.Equal(buf1, buf2) {
				t.Error(fmt.Errorf("buffers not equal (%x != %x)", buf1[:3], buf2[:3]))
			}
		}
	}

	openStreamAndRW := func(c *ps.Conn) {
		log("openStreamAndRW %p, %d nMsg", c, nMsg)

		s, err := c.NewStream()
		if err != nil {
			t.Error(fmt.Errorf("Failed to create NewStream: %s", err))
			return
		}

		bufs := make(chan []byte, nMsg)
		go func() {
			writeStream(s, bufs)
			close(bufs)
		}()

		readStream(s, bufs)
		s.Close()
	}

	openConnAndRW := func(a, b *ps.Swarm) {
		log("openConnAndRW %p -> %p, %d nStream", a, b, nConn)

		ls := b.Listeners()
		l := ls[mrand.Intn(len(ls))]
		nl := l.NetListener()
		nla := nl.Addr()

		tcp := tcpc.NewTCPTransport()
		d, err := tcp.Dialer(zeroaddr)
		checkErr(t, err)

		nc, err := d.Dial(nl.Multiaddr())
		if err != nil {
			t.Fatal(fmt.Errorf("net.Dial(%s, %s): %s", nla.Network(), nla.String(), err))
			return
		}

		c, err := a.AddConn(nc)
		if err != nil {
			t.Fatal(fmt.Errorf("a.AddConn(%s <--> %s): %s", nc.LocalAddr(), nc.RemoteAddr(), err))
			return
		}

		var wg sync.WaitGroup
		for i := 0; i < nStream; i++ {
			wg.Add(1)
			go rateLimit(func() {
				defer wg.Done()
				openStreamAndRW(c)
			})
		}
		wg.Wait()
		c.Close()
	}

	openConnsAndRW := func(a, b *ps.Swarm) {
		log("openConnsAndRW %p -> %p, %d conns", a, b, nConn)

		var wg sync.WaitGroup
		for i := 0; i < nConn; i++ {
			wg.Add(1)
			go rateLimit(func() {
				defer wg.Done()
				openConnAndRW(a, b)
			})
		}
		wg.Wait()
	}

	connectSwarmsAndRW := func(swarms []*ps.Swarm) {
		log("connectSwarmsAndRW %d swarms", len(swarms))

		var wg sync.WaitGroup
		for _, a := range swarms {
			for _, b := range swarms {
				wg.Add(1)
				a := a // race
				b := b // race
				go rateLimit(func() {
					defer wg.Done()
					openConnsAndRW(a, b)
				})
			}
		}
		wg.Wait()
	}

	swarms := makeSwarms(t, tr, nSwarm, 3) // 3 listeners per swarm.
	connectSwarmsAndRW(swarms)
	for _, s := range swarms {
		s.Close()
	}

}

func SubtestStress1Swarm1Conn1Stream1Msg(t *testing.T, tr smux.Transport) {
	SubtestStressNSwarmNConnNStreamNMsg(t, tr, 1, 1, 1, 1)
}

func SubtestStress1Swarm1Conn1Stream100Msg(t *testing.T, tr smux.Transport) {
	SubtestStressNSwarmNConnNStreamNMsg(t, tr, 1, 1, 1, 100)
}

func SubtestStress1Swarm1Conn100Stream100Msg(t *testing.T, tr smux.Transport) {
	SubtestStressNSwarmNConnNStreamNMsg(t, tr, 1, 1, 100, 100)
}

func SubtestStress1Swarm10Conn50Stream50Msg(t *testing.T, tr smux.Transport) {
	SubtestStressNSwarmNConnNStreamNMsg(t, tr, 1, 10, 50, 50)
}

func SubtestStress5Swarm2Conn20Stream20Msg(t *testing.T, tr smux.Transport) {
	SubtestStressNSwarmNConnNStreamNMsg(t, tr, 5, 2, 20, 20)
}

func SubtestStress10Swarm2Conn100Stream100Msg(t *testing.T, tr smux.Transport) {
	SubtestStressNSwarmNConnNStreamNMsg(t, tr, 10, 2, 100, 100)
}

func SubtestAll(t *testing.T, tr smux.Transport) {

	tests := []TransportTest{
		SubtestConstructSwarm,
		SubtestSimpleWrite,
		SubtestSimpleWrite100msgs,
		SubtestStress1Swarm1Conn1Stream1Msg,
		SubtestStress1Swarm1Conn1Stream100Msg,
		SubtestStress1Swarm1Conn100Stream100Msg,
		SubtestStress1Swarm10Conn50Stream50Msg,
		SubtestStress5Swarm2Conn20Stream20Msg,
		// SubtestStress10Swarm2Conn100Stream100Msg, <-- this hoses the osx network stack...
	}

	for _, f := range tests {
		if testing.Verbose() {
			fmt.Fprintf(os.Stderr, "==== RUN %s\n", GetFunctionName(f))
		}
		f(t, tr)
	}
}

type TransportTest func(t *testing.T, tr smux.Transport)

func TestNoOp(t *testing.T) {}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
