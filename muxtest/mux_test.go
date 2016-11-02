package muxtest

import (
	"testing"

	multistream "gx/ipfs/QmVcmcQE9eX4HQ8QwhVXpoHt3ennG7d299NDYFq9D1Uqa1/go-smux-multistream"
	spdy "gx/ipfs/QmWMKNLGkYJTZ4Tq3DQ8E9j86QaKvGjKgFzvLzGYXvW69Z/go-smux-spdystream"
	yamux "gx/ipfs/QmYaeRqthWTco7oQF4dztuqA94P8JF36gVjd2z2eEqKfrh/go-smux-yamux"
	muxado "gx/ipfs/QmZJM54H2j26QwmWAYaW24L8Hwo3ojrQmiGj3gBF9Quj8d/go-smux-muxado"
	multiplex "gx/ipfs/Qmao31fmDJxp9gY7YNqkN1i3kGaknL6XSzKtdC1VCU7Qj8/go-smux-multiplex"
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
