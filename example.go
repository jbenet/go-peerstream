// +build ignore

package peerstream

// Here's the logical model of peerstream.
// - *many* net.Conn
// - *many* net.Listener - generates net.Conns
// - *many* SPDY streams per net.Conn
// - *many* logical Stream groups
// - 1 multiplexor (Swarm)

// create a new Swarm
swarm := peerstream.NewSwarm()
defer swarm.Close()


// tell swarm what to do with a new incoming streams.
// EchoHandler just echos back anything they write.
swarm.SetStreamHandler(peerstream.EchoHandler)

// EchoHandler looks like this:
EchoHandler := func(s Stream) {
  io.Copy(s, s)
}

// Okay, let's try listening on some transports
l1, err := net.Listen("tcp", "localhost:8001")
if err != nil {
  panic(err)
}

l2, err := net.Listen("tcp", "localhost:8002")
if err != nil {
  panic(err)
}

// tell swarm to accept incoming connections on these
// listeners. Swarm will start accepting new connections.
if err := swarm.AddListener(l1); err != nil {
  panic(err)
}
if err := swarm.AddListener(l2); err != nil {
  panic(err)
}

// ok, let's try some outgoing connections
c1, err := net.Dial("tcp", "localhost:8001")
if err != nil {
  panic(err)
}

c2, err := net.Dial("tcp", "localhost:8002")
if err != nil {
  panic(err)
}

// add them to the swarm
if err := swarm.AddConn(c1); err != nil {
  panic(err)
}
if err := swarm.AddConn(c2); err != nil {
  panic(err)
}

// Swarm treats listeners as sources of new connections and does
// not distinguish between outgoing or incoming connections.
// It provides the net.Conn to the StreamHandler so you can
// distinguish between them however you wish.

// now let's try opening some streams!
// You can specify what connection you want to use
s1, err := swarm.NewStreamWithConn(c1)
if err != nil {
  panic(err)
}

// Or, you can specify a SelectConn function that picks between all
// (it calls NewStreamWithConn underneath the hood)
s2, err := swarm.NewStreamSelectConn(func(conns []net.Conn) net.Conn {
  if len(conns) > 0 {
    return conns[0]
  }
  return nil
})
if err != nil {
  panic(err)
}

// Or, you can bind connections to ConnGroup ids. You can bind a conn to
// multiple groups. And, if conn wasn't in swarm, it calls swarm.AddConn.
// You can use any Go `KeyType` as a group A `KeyType` as in maps...)
if err := swarm.AddConnToGroup(c2, 1); err != nil {
  panic(err)
}

// And then use that group to select a connection. Swarm will use any
// connection it finds in that group, using a SelectConn you can rebind:
//   swarm.SetGroupSelectConn(1, SelectConn)
//   swarm.SetDegaultGroupSelectConn(SelectConn)
s3, err := swarm.NewStreamWithGroup(1); err != nil {
  panic(err)
}

// Why groups? It's because with many connections, and many transports,
// and many Servers (or Protocols), we can use the Swarm to associate
// a different StreamHandlers per group, and to let us create NewStreams
// on a given group.


// Ok, we have streams. now what. Use them! Our Streams are basically
// streams from github.com/docker/spdystream, so they work the same
// way:

for i, stream := range []peerstream.Stream{s1, s2, s3} {
  stream.Wait()
  fmt.FprintF(stream, "Hello World %d\n", i)

  buf := make([]byte, 13)
  stream.Read(buf)
  fmt.Println(string(buf))

  stream.Close()
}



r := peerstream.ProtoRouter()
r.AddRoute("bitswap", BitswapHandler)
r.AddRoute("dht", DHTHandler)
r.AddRoute("id", IDHandler)


// The router's StreamHandler does this
swarm.SetStreamHandler(router.StreamHandler())

func (r *router) StreamHandler(s Stream) {

}
