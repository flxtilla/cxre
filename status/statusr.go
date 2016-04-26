package status

import (
	"net/http"

	"github.com/flxtilla/cxre/engine"
	"github.com/flxtilla/cxre/extension"
	"github.com/flxtilla/cxre/state"
)

type Statusr interface {
	GetStatus(int) Status
	SetRawStatus(int, ...state.Manage)
	SetStatus(Status)
	StatusRule() engine.Rule
	SwapStatusRule(engine.Rule)
	StateStatusExtension() extension.Extension
}

type Makes interface {
	Making() state.Make
}

type statusr struct {
	s     map[int]Status
	rule  engine.Rule
	makes Makes
}

func New(m Makes) Statusr {
	s := &statusr{
		s:     make(map[int]Status),
		makes: m,
	}
	s.SwapStatusRule(s.defaultRule)
	return s
}

func (s *statusr) GetStatus(code int) Status {
	if st, ok := s.s[code]; ok {
		return st
	}
	return newStatus(code)
}

func (s *statusr) SetRawStatus(code int, m ...state.Manage) {
	s.s[code] = newStatus(code, m...)
}

func (s *statusr) SetStatus(st Status) {
	s.s[st.Code()] = st
}

func (s *statusr) defaultRule(rw http.ResponseWriter, rq *http.Request, rs *engine.Result) {
	st := s.GetStatus(rs.Code())
	stateFn := s.makes.Making()
	state := stateFn(rw, rq, rs, st.Managers())
	state.Run()
	state.Cancel()
}

func (s *statusr) StatusRule() engine.Rule {
	return s.rule
}

func (s *statusr) SwapStatusRule(er engine.Rule) {
	s.rule = er
}

func (s *statusr) StateStatusExtension() extension.Extension {
	return extension.New("Status_Extension", extension.NewFunction("status", s.StateStatus))
}

func (s *statusr) StateStatus(st state.State, code int) error {
	sts := s.GetStatus(code)
	st.Rerun(sts.Managers()...)
	return nil
}
