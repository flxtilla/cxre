package flash

import (
	"testing"

	"github.com/flxtilla/app"
	"github.com/flxtilla/cxre/state"
	"github.com/flxtilla/cxre/txst"
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

func TestFlash(t *testing.T) {
	a := AppForTest(t, "testFlash")
	exp1, _ := txst.NewExpectation(
		200,
		"GET",
		"/ft1/",
		func(t *testing.T) state.Manage {
			return func(s state.State) {
				//c.Call("flash", "testing", "test flash message")
			}
		},
	)
	exp2, _ := txst.NewExpectation(
		200,
		"GET",
		"/ft2/",
		func(t *testing.T) state.Manage {
			return func(s state.State) {
				//fl := Flshr(c)
				//v := fl.Write("testing")
				//if !strings.Contains(v[0], "test flash message") {
				//	t.Errorf(`flash messaging expected "test flash message" as first message, but it was %s`, v[0])
				//}
				//fl.WriteAll()
				//c.Call("flash", "testing_two", "second test flash message")
			}
		},
	)
	exp3, _ := txst.NewExpectation(
		200,
		"GET",
		"/ft3/",
		func(t *testing.T) state.Manage {
			return func(s state.State) {
				//fl := Flshr(c)
				//nv := fl.Write("testing")
				//if nv != nil {
				//	t.Errorf(`flasher wrote %s for category "testing", but was expecting a nil value`, nv)
				//}
				//v := fl.WriteAll()
				//expected := v["testing_two"][0]
				//if !strings.Contains(expected, "second test flash message") {
				//	t.Errorf(`flash messaging expected "second test flash message" as first message for "testing_two", but it was %s`, expected)
				//}
				//v2 := fl.Write("testing_two")
				//if v2 != nil {
				//	t.Errorf(`flasher wrote %s for category "testing_two", but was expecting a nil value`, v2)
				//}
			}
		},
	)

	txst.SessionPerformer(t, a, exp1, exp2, exp3).Perform()
}
