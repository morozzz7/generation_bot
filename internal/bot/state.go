package bot

import "sync"

type UserState int

const (
	StateIdle UserState = iota
	StateWaitingForModel
	StateWaitingForPrompt
	StateGenerating
)

type StateManager struct {
	mu     sync.Mutex
	states map[int64]UserState
	tasks  map[int64]string
	models map[int64]string
}

func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[int64]UserState),
		tasks:  make(map[int64]string),
		models: make(map[int64]string),
	}
}

func (s *StateManager) Set(userID int64, state UserState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if state == StateIdle {
		delete(s.states, userID)
		delete(s.models, userID)
		return
	}

	s.states[userID] = state
}

func (s *StateManager) Get(userID int64) UserState {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.states[userID]
}

func (s *StateManager) BeginModelSelection(userID int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, active := s.tasks[userID]; active {
		return false
	}

	s.states[userID] = StateWaitingForModel
	delete(s.models, userID)

	return true
}

func (s *StateManager) SelectModel(userID int64, modelID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.states[userID] != StateWaitingForModel {
		return false
	}

	s.models[userID] = modelID
	s.states[userID] = StateWaitingForPrompt

	return true
}

func (s *StateManager) StartTask(
	userID int64,
	taskID string,
) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.states[userID] != StateWaitingForPrompt {
		return "", false
	}

	if _, active := s.tasks[userID]; active {
		return "", false
	}

	modelID, ok := s.models[userID]
	if !ok {
		return "", false
	}

	s.tasks[userID] = taskID
	s.states[userID] = StateGenerating

	return modelID, true
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
	delete(s.models, userID)
}
