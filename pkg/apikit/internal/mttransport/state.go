package mttransport

import "sync"

type Identity struct {
	Token     string
	DID       string
	VisitorID string
}

type State struct {
	mu    sync.RWMutex
	state Identity
}

func NewState(state Identity) *State {
	return &State{state: state}
}

func (s *State) Snapshot() Identity {
	if s == nil {
		return Identity{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *State) Set(next Identity) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = next
}

func (s *State) UpdateDID(did string) {
	if s == nil || did == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.DID = did
}
