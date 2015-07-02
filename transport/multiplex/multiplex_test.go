package peerstream_multiplex

import (
	"testing"

	psttest "github.com/jbenet/go-peerstream/transport/test"
)

func TestMultiplexTransport(t *testing.T) {
	psttest.SubtestAll(t, DefaultTransport)
}
