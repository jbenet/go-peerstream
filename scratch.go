// +build ignore

// this file is just scratch space.


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
