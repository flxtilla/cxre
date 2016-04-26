package state_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flxtilla/app"
	"github.com/flxtilla/cxre/engine"
	"github.com/flxtilla/cxre/extension"
	"github.com/flxtilla/cxre/flash"
	"github.com/flxtilla/cxre/log"
	"github.com/flxtilla/cxre/session"
	"github.com/flxtilla/cxre/state"
	"github.com/flxtilla/cxre/xrr"
	"github.com/flxtilla/txst"
)

func AppForTest(t *testing.T, name string, conf ...app.ConfigurationFn) *app.App {
	conf = append(conf, app.Mode("Testing", true))
	a := app.New(name, conf...)
	err := a.Configure()
	if err != nil {
		t.Errorf("Error in app configuration: %s", err.Error())
	}
	return a
}

func testingState(method string, t *testing.T) {
	var resetPassed, rerunPassed, bouncePassed bool = false, false, false

	exp1, _ := txst.NewExpectation(
		200,
		method,
		"/test_stateReset",
		func(t *testing.T) state.Manage { return func(s state.State) { resetPassed = true } },
	)
	exp1.SetPost(
		func(*testing.T, *httptest.ResponseRecorder) {
			if resetPassed != true {
				t.Errorf("TestState & associated route handler %s were not Reset.", method)
			}
		},
	)

	exp2, _ := txst.NewExpectation(
		200,
		method,
		"/test_stateRerun",
		func(t *testing.T) state.Manage {
			return func(s state.State) { s.Rerun(func(s state.State) { rerunPassed = true }) }
		},
	)
	exp2.SetPost(
		func(*testing.T, *httptest.ResponseRecorder) {
			if rerunPassed != true {
				t.Errorf("TestState & associated route handler %s were not Rerun.", method)
			}
		},
	)

	exp3, _ := txst.NewExpectation(
		200,
		method,
		"/test_stateBounce",
		func(t *testing.T) state.Manage {
			return func(s state.State) { s.Bounce(func(s state.State) { bouncePassed = true }) }
		},
	)
	exp3.SetPost(
		func(*testing.T, *httptest.ResponseRecorder) {
			if bouncePassed != true {
				t.Errorf("TestState & associated route handler %s were not Bounced.", method)
			}
		},
	)

	var original, replicated state.State

	exp4, _ := txst.NewExpectation(
		200,
		method,
		"/test_stateReplicate",
		func(t *testing.T) state.Manage {
			return func(s state.State) { original, replicated = s, s.Replicate() }
		},
	)
	exp4.SetPost(
		func(*testing.T, *httptest.ResponseRecorder) {
			if original == replicated {
				t.Error("replicated State should not equal original state")
			}
		},
	)

	a := AppForTest(t, "testState")

	a.Configure()

	txst.MultiPerformer(t, a, exp1, exp2, exp3, exp4).Perform()
}

func TestDefaultState(t *testing.T) {
	for _, m := range txst.METHODS {
		testingState(m, t)
	}
}

type testState struct {
	engine.Resulter
	flash.Flasher
	extension.Extension
	xrr.Xrroror
	session.SessionStore
	log.Logger
	state.Handlers
	h state.Manage
}

func (s *testState) Request() *http.Request { return nil }

func (s *testState) RWriter() state.ResponseWriter { return nil }

func (s *testState) Reset(*http.Request, http.ResponseWriter, []state.Manage) {}

func (s *testState) Replicate() state.State { return s }

func (s *testState) Run() { s.Next() }

func (s *testState) Rerun(m ...state.Manage) {}

func (s *testState) Next() { s.h(s) }

func (s *testState) Cancel() {}

func MakeTestState(rw http.ResponseWriter, rq *http.Request, rs *engine.Result, m []state.Manage) state.State {
	c := &testState{h: m[0]}
	return c
}

func customState(method string, t *testing.T) {
	//var passed bool = false

	//r := route.New(route.DefaultRouteConf(method, "/test_state", []state.Manage{func(s state.State) { passed = true }}))

	//a := Base("test_ctx")

	//b := NewBlueprint("/")

	//b.MakeCtx = MakeTestCtx

	//a.Blueprint = b

	//a.Config = newConfig()

	//a.Manage(r)

	//a.Configure()

	//ZeroExpectationPerformer(t, a, 200, method, "/test_ctx").Perform()

	//if passed == false {
	//	t.Errorf("TestState & associated route handler %s were not invoked.", method)
	//}
}

func TestCustomState(t *testing.T) {
	for _, m := range txst.METHODS {
		customState(m, t)
	}
}
