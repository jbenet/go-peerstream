package muxtest

import (
	"testing"

	multiplex "github.com/whyrusleeping/go-smux-multiplex"
	multistream "github.com/whyrusleeping/go-smux-multistream"
	muxado "github.com/whyrusleeping/go-smux-muxado"
	spdy "github.com/whyrusleeping/go-smux-spdystream"
	yamux "github.com/whyrusleeping/go-smux-yamux"
)

func TestYamuxTransport(t *testing.T) {
	SubtestAll(t, yamux.DefaultTransport)
}

func TestSpdyStreamTransport(t *testing.T) {
	SubtestAll(t, spdy.Transport)
}

func TestMultiplexTransport(t *testing.T) {
	SubtestAll(t, multiplex.DefaultTransport)
}

func TestMuxadoTransport(t *testing.T) {
	SubtestAll(t, muxado.Transport)
}

func TestMultistreamTransport(t *testing.T) {
	tpt := multistream.NewBlankTransport()
	tpt.AddTransport("/yamux", yamux.DefaultTransport)
	tpt.AddTransport("/spdy", spdy.Transport)
	SubtestAll(t, tpt)
}
