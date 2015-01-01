package peerstream_spdystream

import (
	"testing"

	psttest "github.com/jbenet/go-peerstream/transport/test"
)

func TestSpdyStreamTransport(t *testing.T) {
	psttest.SubtestAll(t, Transport)
}
