package status_test

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flxtilla/app"
	"github.com/flxtilla/cxre/state"
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

func callStatus(status int) state.Manage {
	return func(s state.State) {
		_, _ = s.Call("status", status)
	}
}

func routestring(status int) string {
	return fmt.Sprintf("/call%d", status)
}

func testStatus(t *testing.T, status int, method, expects string) {
	a := AppForTest(t, "statuses")

	exp, _ := txst.NewExpectation(
		status,
		method,
		routestring(status),
		func(t *testing.T) state.Manage {
			return callStatus(status)
		},
	)
	exp.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			if bytes.Compare(r.Body.Bytes(), []byte(expects)) != 0 {
				t.Errorf("Status test expected %s, but got %s.", expects, r.Body.String())
			}
		},
	)

	txst.SimplePerformer(t, a, exp).Perform()
}

func TestStatus(t *testing.T) {
	for _, m := range txst.METHODS {
		testStatus(t, 418, m, "418 I'm a teapot")
	}
}

func test500(method string, t *testing.T) {
	a := AppForTest(t, "panic")
	exp, _ := txst.NewExpectation(
		500,
		method,
		routestring(500),
		func(t *testing.T) state.Manage {
			return func(s state.State) {
				panic("Test panic!")
			}
		},
	)
	exp.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			if !strings.Contains(r.Body.String(), "Test panic!") {
				t.Errorf(`Status test 500 expected to contain "Test Panic!", but did not.`)
			}
		},
	)
	txst.SimplePerformer(t, a, exp).Perform()
}

func Test500(t *testing.T) {
	for _, m := range txst.METHODS {
		test500(m, t)
	}
}

func Custom404(s state.State) {
	s.Call("serve_plain", 404, "I AM NOT FOUND :: 404")
}

func Custom418(s state.State) {
	s.Call("serve_plain", 418, "I AM TEAPOT :: 418")
}

type mockStatus struct {
	code     int
	managers []state.Manage
}

func (s *mockStatus) Code() int {
	return s.code
}

func (s *mockStatus) Managers() []state.Manage {
	return s.managers
}

func customStatus(t *testing.T, method string, expects string, status int, m state.Manage, raw bool) {
	a := AppForTest(t, "customstatus")

	if raw {
		a.STATUS(status, m)
	} else {
		st := &mockStatus{status, []state.Manage{m}}
		a.SetStatus(st)
	}

	exp, _ := txst.NewExpectation(
		status,
		"GET",
		routestring(status),
		func(t *testing.T) state.Manage {
			return func(s state.State) {
				_, _ = s.Call("status", status)
			}
		},
	)
	exp.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			if bytes.Compare(r.Body.Bytes(), []byte(expects)) != 0 {
				t.Errorf("Custom status %d test expected %s, but got %s.", status, expects, r.Body.String())
			}
		},
	)

	txst.SimplePerformer(t, a, exp).Perform()
}

func TestCustomStatus(t *testing.T) {
	for _, m := range txst.METHODS {
		customStatus(t, m, "I AM NOT FOUND :: 404", 404, Custom404, false)
		customStatus(t, m, "I AM TEAPOT :: 418", 418, Custom418, true)
	}
}
