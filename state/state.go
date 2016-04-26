package state

import (
	"fmt"
	"net/http"

	"github.com/flxtilla/cxre/engine"
	"github.com/flxtilla/cxre/extension"
	"github.com/flxtilla/cxre/flash"
	"github.com/flxtilla/cxre/log"
	"github.com/flxtilla/cxre/session"
	"github.com/flxtilla/cxre/xrr"
)

type Make func(
	http.ResponseWriter,
	*http.Request,
	*engine.Result,
	[]Manage,
) State

type Manage func(State)

type State interface {
	engine.Resulter
	flash.Flasher
	extension.Extension
	xrr.Xrroror
	session.SessionStore
	log.Logger
	Handlers
	Request() *http.Request
	RWriter() ResponseWriter
	Reset(*http.Request, http.ResponseWriter, []Manage)
	Replicate() State
	Run()
	Rerun(...Manage)
	Next()
	Cancel()
}

type Handlers interface {
	Push(Manage)
	Bounce(Manage)
}

type handlers struct {
	index    int8
	managers []Manage
	deferred []Manage
}

func defaultHandlers() *handlers {
	return &handlers{index: -1}
}

func (h *handlers) Push(fn Manage) {
	h.deferred = append(h.deferred, fn)
}

func (h *handlers) Bounce(fn Manage) {
	h.deferred = []Manage{fn}
}

type state struct {
	*engine.Result
	*context
	*handlers
	xrr.Xrroror
	extension.Extension
	session.SessionStore
	rw      responseWriter
	RW      ResponseWriter
	request *http.Request
	Data    map[string]interface{}
	log.Logger
	flash.Flasher
}

func empty() *state {
	return &state{
		handlers: defaultHandlers(),
	}
}

func New(ext extension.Extension, rs *engine.Result, lg log.Logger) *state {
	s := empty()
	s.Logger = lg
	s.Result = rs
	s.Xrroror = rs.Xrroror
	s.Extension = ext
	s.RW = &s.rw
	s.Flasher = flash.New()
	return s
}

func (s *state) Request() *http.Request {
	return s.request
}

func (s *state) RWriter() ResponseWriter {
	return s.RW
}

func release(s State) {
	w := s.RWriter()
	if !w.Written() {
		s.Out(s)
		s.SessionRelease(w)
	}
}

func LogFmt(s *state) string {
	r := s.Record()
	return fmt.Sprintf("| %3d | %12v | %s |%-7s %s",
		r.Status,
		r.Latency,
		r.Requester,
		r.Method,
		r.Path,
	)
}

func (s *state) Run() {
	s.Push(release)
	s.Next()
	for _, fn := range s.deferred {
		fn(s)
	}
	s.PostProcess(s.request, s.RW.Status())
	s.Logger.Printf(LogFmt(s))
}

func (s *state) Rerun(managers ...Manage) {
	if s.index != -1 {
		s.index = -1
	}
	s.managers = managers
	s.Next()
}

func (s *state) Next() {
	s.index++
	lm := int8(len(s.managers))
	for ; s.index < lm; s.index++ {
		s.managers[s.index](s)
	}
}

func (s *state) Cancel() {
	s.PostProcess(s.request, s.RW.Status())
	s.context.cancel(true, Canceled)
}

func (s *state) Reset(rq *http.Request, rw http.ResponseWriter, m []Manage) {
	s.request = rq
	s.rw.reset(rw)
	s.context = &context{done: make(chan struct{}), value: s}
	s.handlers = defaultHandlers()
	s.managers = m
	s.Insert(s)
}

func (s *state) Replicate() State {
	child := &context{parent: s.context, done: make(chan struct{}), value: s}
	propagateCancel(s.context, child)
	var rcopy state = *s
	rcopy.context = child
	return &rcopy
}
