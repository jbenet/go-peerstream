package peerstream_muxado

import (
	"testing"

	test "github.com/jbenet/go-peerstream/Godeps/_workspace/src/github.com/jbenet/go-stream-muxer/test"
)

func TestMuxadoTransport(t *testing.T) {
	test.SubtestAll(t, Transport)
}
