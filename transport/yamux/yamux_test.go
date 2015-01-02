package peerstream_yamux

import (
	"testing"

	psttest "github.com/jbenet/go-peerstream/transport/test"
)

func TestYamuxTransport(t *testing.T) {
	psttest.SubtestAll(t, DefaultTransport)
}
