package syncmap

import (
	"bytes"
	"cmp"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"
)

// SortedKeys returns a sorted slice of the map's keys
func SortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

// ComparableAndOrdered defines the type constraints for SynchronisedMap
type ComparableAndOrdered interface {
	comparable
	cmp.Ordered
}

// New returns an instance of SynchronisedMap, containing the
// contents of the init map
func New[T ComparableAndOrdered, U any](init map[T]U) *SynchronisedMap[T, U] {
	m := &SynchronisedMap[T, U]{
		m: map[T]U{},
	}

	for k, v := range init {
		m.m[k] = v
	}

	return m
}

// ErrMissingKey is returned if the requested key is not in the map
var ErrMissingKey = errors.New("unknown key")

// ErrKeyExists is returned if Insert is called and key already exists
var ErrKeyExists = errors.New("key already exists")

// SynchronisedMap provides a concurrency safe map
type SynchronisedMap[T ComparableAndOrdered, U any] struct {
	lck sync.RWMutex
	m   map[T]U
}

// Insert adds the value at the specified key.
// If errIfExists is true and the key exists, then an error is raised.  Otherwise
// the value is inserted at the key, and any pre-existing value returned.
func (s *SynchronisedMap[T, U]) Insert(k T, v U, errIfExists bool) (U, error) {
	s.lck.Lock()
	defer s.lck.Unlock()

	var r U
	old, ok := s.m[k]
	if !ok {
		s.m[k] = v
		return r, nil
	}

	if errIfExists {
		return r, ErrKeyExists
	}

	s.m[k] = v
	return old, nil
}

// GetKeys returns the keys, sorted, within the map
func (s *SynchronisedMap[T, U]) GetKeys() []T {
	s.lck.RLock()
	defer s.lck.RUnlock()

	return SortedKeys(s.m)
}

// Contains returns true if the key is found
func (s *SynchronisedMap[T, U]) Contains(id T) bool {
	s.lck.RLock()
	defer s.lck.RUnlock()

	_, ok := s.m[id]
	return ok
}

// Get returns the value associated with the key,
// or a key missing error
func (s *SynchronisedMap[T, U]) Get(id T) (U, error) {
	s.lck.RLock()
	defer s.lck.RUnlock()

	if t, ok := s.m[id]; ok {
		return t, nil
	}

	var r U
	return r, ErrMissingKey
}

// Remove deletes the key from the map
func (s *SynchronisedMap[T, U]) Remove(id T) {
	s.lck.Lock()
	defer s.lck.Unlock()

	delete(s.m, id)
}

// Len returns the current length
func (s *SynchronisedMap[T, U]) Len() int {
	s.lck.RLock()
	defer s.lck.RUnlock()

	return len(s.m)
}

func (s *SynchronisedMap[T, U]) snap() map[T]U {
	s.lck.RLock()
	defer s.lck.RUnlock()

	m := map[T]U{}
	for k, v := range s.m {
		m[k] = v
	}

	return m
}

// String returns the contents of the map as a string,
// with entries ordered based on the key type T
func (s *SynchronisedMap[T, U]) String() string {

	m := s.snap()

	// Apply ordering so the output is deterministic
	keys := SortedKeys[T, U](m)
	buf := new(bytes.Buffer)
	io.WriteString(buf, "map[")
	for i, key := range keys {
		io.WriteString(buf, fmt.Sprint(key))
		io.WriteString(buf, ":")
		io.WriteString(buf, fmt.Sprint(m[key]))
		if i < len(keys)-1 {
			io.WriteString(buf, " ")
		}
	}
	io.WriteString(buf, "]")

	return buf.String()
}

// Bytes serialises the current contents of the map
func (s *SynchronisedMap[T, U]) Bytes() ([]byte, error) {

	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)

	if err := enc.Encode(s.snap()); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Merge attempts to decode the slice, assuming it is of the
// same type as returned by Bytes().  If successful, then
// adds any missing key/value pairs into this instance of the map.
func (s *SynchronisedMap[T, U]) Merge(b []byte) error {
	buf := new(bytes.Buffer)
	buf.Write(b)
	dec := gob.NewDecoder(buf)

	m := map[T]U{}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	s.lck.Lock()
	defer s.lck.Unlock()

	for k, v := range m {
		if _, ok := s.m[k]; !ok {
			s.m[k] = v
		}
	}

	return nil
}
