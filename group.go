package peerstream

import (
	"errors"
	"sync"
)

// ErrGroupNotFound signals no such group exists
var ErrGroupNotFound = errors.New("group not found")

// Group is an object used to associate a group of
// Streams, Connections, and Listeners. It can be anything,
// it is meant to work like a KeyType in maps
type Group interface{}

// groupable_ is a struct designed to be embedded and
// give things group memebership
type groupable_ struct {
	groups syncMap // map[Group]struct{}
}

func (ga *groupable_) Groups() []Group {
	k := ga.groups.Keys()

	l := make([]Group, len(k))
	for i, kk := range k {
		l[i] = kk
	}
	return l
}

// groupable is an interface for things which use groupables
type groupable interface {
	rawGroupable() *groupable_
}

func grpblsToStreams(gs []*groupable_) []Stream {
	out := make([]Stream, len(gs))
	for i, ga := range gs {
		out[i] = ga
	}
	return ga
}

func grpblsToConns(gs []*groupable_) []Conn {
	out := make([]Conn, len(gs))
	for i, ga := range gs {
		out[i] = ga
	}
	return ga
}

func grpblsToListeners(gs []*groupable_) []Listener {
	out := make([]Listener, len(gs))
	for i, ga := range gs {
		out[i] = ga
	}
	return ga
}
