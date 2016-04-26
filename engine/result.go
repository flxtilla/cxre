package engine

import (
	"net/http"
	"time"

	"github.com/flxtilla/cxre/xrr"
)

// The Resulter interface provides integer code information, Params, and a
// Recorder where needed.
type Resulter interface {
	Code() int
	Params() Params
	Recorder
}

// Result is a struct for holding information on found routes when handled by
// an Engine.
type Result struct {
	code   int
	Rule   Rule
	params Params
	TSR    bool
	xrr.Xrroror
	Recorder
}

// NewResult provides a Result instance, provided an integer code, a Rule,
// Params, and a boolean indicating trailing slash redirect.
func NewResult(code int, rule Rule, params Params, tsr bool) *Result {
	return &Result{
		code:     code,
		Rule:     rule,
		params:   params,
		TSR:      tsr,
		Xrroror:  xrr.NewXrroror(),
		Recorder: newRecorder(),
	}
}

// The Result structure Code function returning an integer.
func (r *Result) Code() int {
	return r.code
}

// the Ressult structure Params function, returning a Params set.
func (r *Result) Params() Params {
	return r.params
}

// Recorder is an interface for recording request & handling data.
type Recorder interface {
	PostProcess(*http.Request, int)
	Record() *Recorded
}

// The Recorded struct holds request & handling data.
type Recorded struct {
	Start, Stop             time.Time
	Latency                 time.Duration
	Status                  int
	Method, Path, Requester string
}

type recorder struct {
	*Recorded
}

func newRecorder() Recorder {
	return &recorder{&Recorded{Start: time.Now()}}
}

func (r *recorder) stopRecorder() {
	r.Stop = time.Now()
}

func (r *recorder) latency() time.Duration {
	return r.Stop.Sub(r.Start)
}

// The default Recorder Record function, returns a Recorded instance.
func (r *recorder) Record() *Recorded {
	return r.Recorded
}

// The default Recorder PostProcess function takes a http.Request instance and
// a designated integer status to record latency, requester, method, and path
// data about a request.
func (r *recorder) PostProcess(req *http.Request, withstatus int) {
	r.stopRecorder()
	r.Latency = r.latency()
	r.Requester = req.RemoteAddr
	r.Method = req.Method
	r.Path = req.URL.Path
	r.Status = withstatus
}
