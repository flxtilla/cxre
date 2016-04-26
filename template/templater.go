package template

import (
	"io"
	"strings"

	"github.com/flxtilla/cxre/asset"
	"github.com/flxtilla/cxre/state"
	"github.com/flxtilla/cxre/store"
	"github.com/flxtilla/cxre/xrr"
	"github.com/thrisp/djinn"
)

type Templater interface {
	TemplateDirs(...string) []string
	ListTemplates() []string
	AddTemplateFunctions(fns map[string]interface{}) error
	SetTemplateFunctions()
	Render(io.Writer, string, interface{}) error
	RenderTemplate(state.State, string, interface{}) error
}

type templater struct {
	*djinn.Djinn
	s store.Store
	a asset.Assets
	*functions
}

func DefaultTemplater(s store.Store, a asset.Assets) Templater {
	t := &templater{
		Djinn: djinn.Empty(),
		s:     s,
		a:     a,
	}
	t.functions = &functions{t, false, nil}
	t.AddConfig(djinn.Loaders(newLoader(t)))
	return t
}

func doAdd(s string, ss []string) []string {
	if isAppendable(s, ss) {
		ss = append(ss, s)
	}
	return ss
}

func isAppendable(s string, ss []string) bool {
	for _, x := range ss {
		if x == s {
			return false
		}
	}
	return true
}

func (t *templater) TemplateDirs(added ...string) []string {
	dirs := t.s.List("TEMPLATE_DIRECTORIES")
	if added != nil {
		for _, add := range added {
			dirs = doAdd(add, dirs)
		}
		t.s.Add("TEMPLATE_DIRECTORIES", strings.Join(dirs, ","))
	}
	return dirs
}

func (t *templater) ListTemplates() []string {
	var ret []string
	for _, l := range t.Djinn.GetLoaders() {
		ts := l.ListTemplates()
		ret = append(ret, ts...)
	}
	return ret
}

type functions struct {
	t                 *templater
	set               bool
	templateFunctions map[string]interface{}
}

func addTplFunc(f *functions, name string, fn interface{}) {
	if f.templateFunctions == nil {
		f.templateFunctions = make(map[string]interface{})
	}
	f.templateFunctions[name] = fn
}

func (f *functions) SetTemplateFunctions() {
	f.t.AddConfig(djinn.TemplateFunctions(f.templateFunctions))
	f.set = true
}

var AlreadySet = xrr.NewXrror("cannot set template functions: already set")

func (f *functions) AddTemplateFunctions(fns map[string]interface{}) error {
	if !f.set {
		for k, v := range fns {
			addTplFunc(f, k, v)
		}
		return nil
	}
	return AlreadySet
}

func (t *templater) RenderTemplate(s state.State, template string, data interface{}) error {
	s.Push(func(ps state.State) {
		td := NewTemplateData(ps, data)
		t.Render(ps.RWriter(), template, td)
	})
	return nil
}
