package bot

import "sync"

type UserState int

const (
	StateIdle UserState = iota
	StateWaitingForPrompt
	StateGenerating
)

type StateManager struct {
	mu     sync.Mutex
	states map[int64]UserState
	tasks map[int64]string
}

func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[int64]UserState),
		tasks: make(map[int64]string),
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

func (s *StateManager) SetTask(userID int64, taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[userID] = taskID
	s.states[userID] = StateGenerating
}

func (s *StateManager) GetTask(userID int64) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	taskID, ok := s.tasks[userID]
	return taskID, ok
}

func (s *StateManager) ClearTask(userID int64, taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.tasks[userID] != taskID {
		return
	}

	delete(s.tasks, userID)
	delete(s.states, userID)
}
