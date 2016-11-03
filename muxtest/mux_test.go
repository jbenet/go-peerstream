package muxtest

import (
	"testing"

	yamux "gx/ipfs/QmSHTSkxXGQgaHWz91oZV3CDy3hmKmDgpjbYRT6niACG4E/go-smux-yamux"
	multistream "gx/ipfs/Qme8hbiTP4VNr1s7FxsfnnqrxbxPz3KPWtuGYeGbtFnhGC/go-smux-multistream"
	spdy "gx/ipfs/QmfXgTygwsTPyUWPWTAeBK6cFtTdMqmeeqhyhcNMhRpT1g/go-smux-spdystream"
)

func TestYamuxTransport(t *testing.T) {
	SubtestAll(t, yamux.DefaultTransport)
}

func TestSpdyStreamTransport(t *testing.T) {
	SubtestAll(t, spdy.Transport)
}

/*
func TestMultiplexTransport(t *testing.T) {
	SubtestAll(t, multiplex.DefaultTransport)
}

func TestMuxadoTransport(t *testing.T) {
	SubtestAll(t, muxado.Transport)
}
*/

func TestMultistreamTransport(t *testing.T) {
	tpt := multistream.NewBlankTransport()
	tpt.AddTransport("/yamux", yamux.DefaultTransport)
	tpt.AddTransport("/spdy", spdy.Transport)
	SubtestAll(t, tpt)
}
