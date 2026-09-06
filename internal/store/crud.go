package store

import "sort"

// maxFlags is the maximum number of flags the store will hold.
const maxFlags = 1000

// Create stores a new flag. It returns ErrDuplicate if a flag with the same
// key already exists, and ErrTooManyFlags if the store already holds maxFlags
// flags.
func (s *Store) Create(f Flag) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.flags[f.Key]; ok {
		return ErrDuplicate
	}
	if len(s.flags) >= maxFlags {
		return ErrTooManyFlags
	}
	s.flags[f.Key] = f
	return nil
}

// List returns all stored flags sorted by key.
func (s *Store) List() []Flag {
	s.mu.RLock()
	defer s.mu.RUnlock()

	flags := make([]Flag, 0, len(s.flags))
	for _, f := range s.flags {
		flags = append(flags, f)
	}
	sort.Slice(flags, func(i, j int) bool {
		return flags[i].Key < flags[j].Key
	})
	return flags
}

// Get returns the flag for key and whether it exists.
func (s *Store) Get(key string) (Flag, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, ok := s.flags[key]
	return f, ok
}

// Update replaces the flag identified by key. It returns the stored flag and
// false if no flag with that key exists.
func (s *Store) Update(key string, f Flag) (Flag, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.flags[key]; !ok {
		return Flag{}, false
	}
	f.Key = key
	s.flags[key] = f
	return f, true
}

// Delete removes the flag identified by key and reports whether it existed.
func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.flags[key]; !ok {
		return false
	}
	delete(s.flags, key)
	return true
}
