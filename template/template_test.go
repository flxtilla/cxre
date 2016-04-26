package template

import (
	"fmt"
	"html/template"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flxtilla/app"
	"github.com/flxtilla/cxre/extension"
	"github.com/flxtilla/cxre/state"
	"github.com/flxtilla/cxre/template/resources"
	"github.com/flxtilla/txst"
)

func testLocation() string {
	wd, _ := os.Getwd()
	ld, _ := filepath.Abs(wd)
	return ld
}

func testTemplateDirectory() string {
	return filepath.Join(testLocation(), "resources", "templates")
}

func trimTemplates(t []string) []string {
	var templates []string
	for _, f := range t {
		templates = append(templates, filepath.Base(f))
	}
	return templates
}

func stringTemplates(t []string) string {
	templates := trimTemplates(t)
	return strings.Join(templates, ",")
}

var tplFuncs map[string]interface{} = map[string]interface{}{
	"Hello": func(s string) string { return fmt.Sprintf("Hello World!: %s", s) },
}

func mkFunction(k string, v interface{}) extension.Function {
	return extension.NewFunction(k, v)
}

func tplFuncsConf(tf map[string]interface{}) app.ConfigurationFn {
	return func(a *app.App) error {
		a.AddTemplateFunctions(tf)
		ext := extension.New(
			"template_test_functions",
			mkFunction("fn_string", fnString),
			mkFunction("fn_html", fnHtml),
		)
		a.Extend(ext)
		return nil
	}
}

func fnString(s state.State) string {
	return fmt.Sprintf("returned STRING")
}

func fnHtml(s state.State) template.HTML {
	return "<div>returned HTML</div>"
}

func templateContains(t *testing.T, body string, must string) {
	if !strings.Contains(body, must) {
		t.Errorf(`Test template was not rendered correctly, expecting %s but it is not present:
		%s
		`, must, body)
	}
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

func existsIn(s string, l []string) bool {
	for _, x := range l {
		if s == x {
			return true
		}
	}
	return false
}

func TestDefaultTemplating(t *testing.T) {
	a := AppForTest(
		t,
		"testDefaultTemplating",
		app.Assets(resources.ResourceFS),
		tplFuncsConf(tplFuncs),
	)

	a.TemplateDirs(testTemplateDirectory())

	var expected []string = []string{"layout.html", "test.html", "layout_asset.html", "test_asset.html"}
	var existing []string = trimTemplates(a.ListTemplates())

	for _, ex := range expected {
		if !existsIn(ex, existing) {
			t.Errorf(`Existing templates do not contain %s`, ex)
		}
	}

	exp1, _ := txst.NewExpectation(
		200,
		"GET",
		"/template",
		func(t *testing.T) state.Manage {
			return func(s state.State) {
				ret := make(map[string]interface{})
				ret["Title"] = "rendered from test template test.html"
				s.Flash("test_flash", "TEST_FLASH_ONE")
				s.Flash("test_flash", "TEST_FLASH_TWO")
				//	s.Call("set", "set_in_ctx", "SET_IN_CTX")
				s.Call("render_template", "test.html", ret)
			}
		},
	)
	exp1.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			lookfor := []string{
				`<div>TEST TEMPLATE</div>`,
				`Hello World!: TEST`,
				`returned STRING`,
				`<div>returned HTML</div>`,
				`[TEST_FLASH_ONE TEST_FLASH_TWO]`,
				// `/template?value1%3Dadditional`,
				// `Unable to get url for route \does\not\exist\p\s\get with params [param /a/splat/fragment].`,
				// `SET_IN_CTX`,
			}

			for _, lf := range lookfor {
				templateContains(t, r.Body.String(), lf)
			}
		},
	)

	exp2, _ := txst.NewExpectation(
		200,
		"GET",
		"/asset_template",
		func(t *testing.T) state.Manage {
			return func(s state.State) {
				s.Call("render_template", "test_asset.html", "rendered from test template test_asset.html")
			}
		},
	)
	exp2.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			//templateContains(t, r.Body.String(), `<title>rendered from test template test_asset.html</title>`)
		},
	)

	txst.MultiPerformer(t, a, exp1, exp2).Perform()
}

type testTemplatr struct{}

func (tt *testTemplatr) TemplateDirs(s ...string) []string {
	return []string{"test_templatr_dirs"}
}

func (tt *testTemplatr) ListTemplates() []string {
	return []string{"test_templatr_templates"}
}

func (tt *testTemplatr) AddTemplateFunctions(fns map[string]interface{}) error {
	return nil
}

func (tt *testTemplatr) SetTemplateFunctions() {}

func (tt *testTemplatr) Render(w io.Writer, s string, i interface{}) error {
	_, err := w.Write([]byte("test templator"))
	if err != nil {
		return err
	}
	return nil
}

func (tt *testTemplatr) RenderTemplate(s state.State, template string, data interface{}) error {
	s.Push(func(ps state.State) {
		td := NewTemplateData(ps, data)
		tt.Render(ps.RWriter(), template, td)
	})
	return nil
}

func TestTemplatr(t *testing.T) {
	tt := &testTemplatr{}
	a := AppForTest(
		t,
		"testExternalTemplator",
		app.Assets(resources.ResourceFS),
	)
	a.SwapTemplatr(tt)
	exp, _ := txst.NewExpectation(
		200,
		"GET",
		"/templator/",
		func(t *testing.T) state.Manage {
			return func(s state.State) {
				s.Call("render_template", "test.html", "test data")
			}
		},
	)
	exp.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			b := r.Body.String()
			if b != "test templator" {
				t.Errorf(`Test external templator rendered %s, not "test templator"`, b)
			}

		},
	)
	txst.SimplePerformer(t, a, exp).Perform()
}
