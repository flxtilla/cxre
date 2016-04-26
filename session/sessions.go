package session

import (
	"fmt"
	"net/http"

	"github.com/flxtilla/cxre/store"
)

type Sessions interface {
	Manager() *Manager
	SwapManager(*Manager)
	Init()
	Start(http.ResponseWriter, *http.Request) (SessionStore, error)
}

type sessions struct {
	defaultConfig string
	manager       *Manager
}

func NewSessions(s store.Store) Sessions {
	return &sessions{
		defaultConfig: defaultSessionConfig(s),
	}
}

func (s *sessions) Manager() *Manager {
	return s.manager
}

func (s *sessions) SwapManager(m *Manager) {
	s.manager = m
	s.Init()
}

func defaultSessionConfig(s store.Store) string {
	cookieName := s.String("session_cookiename")
	secretKey := s.String("secret_key")
	sessionLifetime := s.Int64("session_lifetime")
	prvdrcfg := fmt.Sprintf(`"ProviderConfig":"{\"maxage\": %d,\"cookieName\":\"%s\",\"securityKey\":\"%s\"}"`, sessionLifetime, cookieName, secretKey)
	return fmt.Sprintf(`{"cookieName":"%s","enableSetCookie":false,"gclifetime":3600, %s}`, cookieName, prvdrcfg)
}

func (s *sessions) defaultSessionManager() *Manager {
	d, err := NewManager("cookie", s.defaultConfig)
	if err != nil {
		panic(fmt.Sprintf("sessions default session manager error: %s", err))
	}
	return d
}

// SessionInit intializes the SessionManager stored with the Env.
func (s *sessions) Init() {
	if s.manager == nil {
		s.manager = s.defaultSessionManager()
	}
	go s.manager.GC()
}

func (s *sessions) Start(w http.ResponseWriter, r *http.Request) (SessionStore, error) {
	st, err := s.manager.SessionStart(w, r)
	if err != nil {
		return nil, err
	}
	return st, nil
}
