package bot

import "sync"

type UserState int

const (
	StateIdle UserState = iota
	StateWaitingForPrompt
)

type StateManager struct {
	mu     sync.Mutex
	states map[int64]UserState
}

func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[int64]UserState),
	}
}

func (s *StateManager) Set(userID int64, state UserState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if state == StateIdle {
		delete(s.states, userID)
		return
	}

	s.states[userID] = state
}

func (s *StateManager) Get(userID int64) UserState {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.states[userID]
}
