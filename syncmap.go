package peerstream

// syncMap is a threadsafe map.
type syncMap struct {
	m map[interface{}]interface{}
	sync.RWMutex
}

func (s *syncMap) Set(k, v interface{}) {
	s.Lock()
	defer s.Unlock()
	s.m[k] = v
}

func (s *syncMap) Remove(k interface{}) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, k)
}

func (s *syncMap) Has(k interface{}) {
	s.RLock()
	defer s.RUnlock()
	_, ok := s[k]
	return ok
}

func (s *syncMap) Get(k interface{}) interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.m[k]
}

func (s *syncMap) Keys() []interface{} {
	s.RLock()
	defer s.RUnlock()

	out := make([]interface{}, 0, len(s.m))
	for k := range s.m {
		out = append(out, k)
	}
	return out
}

func (s *syncMap) Values() []interface{} {
	s.RLock()
	defer s.RUnlock()

	out := make([]interface{}, 0, len(s.m))
	for _, v := range s.m {
		out = append(out, v)
	}
	return out
}
