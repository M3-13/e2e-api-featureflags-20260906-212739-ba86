package store

import (
	"errors"
	"sync"
)

type Flag struct {
	Key            string `json:"key"`
	Enabled        bool   `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent int    `json:"rollout_percent"`
}

type Store struct {
	mu    sync.RWMutex
	flags map[string]Flag
}

var (
	ErrNotFound  = errors.New("flag not found")
	ErrDuplicate = errors.New("flag already exists")
)

func New() *Store {
	return &Store{
		flags: make(map[string]Flag),
	}
}
