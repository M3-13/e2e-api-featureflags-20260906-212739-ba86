package store

import (
	"hash/fnv"
)

// Evaluate deterministically decides whether the feature flag identified by key
// is enabled for the given user. It returns ErrNotFound when the key is unknown.
// The decision is stable: the same key and user always produce the same result.
func (s *Store) Evaluate(key, user string) (bool, error) {
	s.mu.RLock()
	flag, ok := s.flags[key]
	s.mu.RUnlock()
	if !ok {
		return false, ErrNotFound
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(key + "\x00" + user))
	bucket := h.Sum32() % 100

	return flag.Enabled && int(bucket) < flag.RolloutPercent, nil
}
