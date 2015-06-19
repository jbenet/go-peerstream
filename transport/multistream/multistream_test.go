package multistream

import (
	"testing"

	psttest "github.com/jbenet/go-peerstream/transport/test"
)

func TestMultiStreamTransport(t *testing.T) {
	psttest.SubtestAll(t, NewTransport())
}
