package static_test

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flxtilla/app"
	"github.com/flxtilla/cxre/state"
	"github.com/flxtilla/cxre/static/resources"
	"github.com/flxtilla/txst"
)

func testLocation() string {
	wd, _ := os.Getwd()
	ld, _ := filepath.Abs(wd)
	return ld
}

func AppForTest(t *testing.T, name string, conf ...app.ConfigurationFn) *app.App {
	conf = append(conf, app.Mode("Testing", true))
	a := app.New(name, conf...)
	err := a.Configure()
	if err != nil {
		t.Errorf("Error in app configuration: %s", err.Error())
	}
	return a
}

func TestStatic(t *testing.T) {
	a := AppForTest(
		t,
		"testStatic",
		app.Assets(resources.ResourceFS),
	)

	a.STATIC(a.Environment, "/resources/static/*filepath", "resources")

	a.StaticDirs("resources")

	txst.ZeroExpectationPerformer(t, a, 200, "GET", "/resources/static/css/static.css").Perform()

	exp1, _ := txst.NewExpectation(
		200,
		"GET",
		"/static/css/static.css",
	)
	exp1.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			b := r.Body.String()
			if !strings.Contains(b, "test css file") {
				t.Error(`Test css file did not return "test css file"`)
			}
		},
	)
	exp1.SetPreRegister(true)

	exp2, _ := txst.NewExpectation(
		200,
		"GET",
		"/static/css/css/css/css_asset.css",
	)
	exp2.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			b := r.Body.String()
			expected := "/* Test binary css asset */"
			if !strings.Contains(b, expected) {
				t.Error(`Test css asset file did not return %s`, expected)
			}
		},
	)
	exp2.SetPreRegister(true)
	txst.MultiPerformer(t, a, exp1, exp2).Perform()

	txst.ZeroExpectationPerformer(t, a, 404, "GET", "/static/css/no.css").Perform()
}

type testStatic struct{}

func (ts *testStatic) StaticDirs(d ...string) []string {
	return []string{""}
}

func (ts *testStatic) Exists(s state.State, str string) bool {
	return true
}

func (ts *testStatic) StaticManage(s state.State) {
	s.Call("write_to_response", "from external staticr")
}

func TestCustomStaticor(t *testing.T) {
	ss := &testStatic{}
	a := AppForTest(
		t,
		"testExternalStaticor",
	)
	a.SwapStaticr(ss)
	exp, _ := txst.NewExpectation(
		200,
		"GET",
		"/test/staticr/",
		func(t *testing.T) state.Manage {
			return a.StaticManage
		},
	)
	exp.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			b := r.Body.String()
			if b != "from external staticr" {
				t.Error(`Test external staticr did not return "from external staticr"`)
			}

		},
	)
	txst.SimplePerformer(t, a, exp).Perform()
}
