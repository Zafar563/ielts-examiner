package bot

import (
	"sync"
)

type UserStep int

const (
	StepIdle UserStep = iota
	StepWaitingLang
	StepWaitingTopic
	StepWaitingEssay
)

type UserSession struct {
	Step  UserStep
	Lang  string
	Topic string
}

type StateManager struct {
	mu       sync.RWMutex
	sessions map[int64]*UserSession
}

func NewStateManager() *StateManager {
	return &StateManager{
		sessions: make(map[int64]*UserSession),
	}
}

func (s *StateManager) GetSession(userID int64) *UserSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[userID]
	if !exists {
		session = &UserSession{Step: StepIdle}
		s.sessions[userID] = session
	}
	return session
}

func (s *StateManager) ResetSession(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, userID)
}
