package static

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flxtilla/cxre/asset"
	"github.com/flxtilla/cxre/state"
	"github.com/flxtilla/cxre/store"
)

type Static interface {
	Staticr
	SwapStaticr(Staticr)
}

type static struct {
	Staticr
}

func New(s store.Store, a asset.Assets) Static {
	return &static{
		Staticr: defaultStaticr(s, a),
	}
}

func (s *static) SwapStaticr(st Staticr) {
	s.Staticr = st
}

type Staticr interface {
	StaticDirs(...string) []string
	Exists(state.State, string) bool
	StaticManage(state.State)
}

type staticr struct {
	s store.Store
	a asset.Assets
}

func defaultStaticr(s store.Store, a asset.Assets) Staticr {
	return &staticr{
		s: s,
		a: a,
	}
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

func (st *staticr) StaticDirs(added ...string) []string {
	dirs := st.s.List("STATIC_DIRECTORIES")
	if added != nil {
		for _, add := range added {
			dirs = doAdd(add, dirs)
		}
		st.s.Add("STATIC_DIRECTORIES", strings.Join(dirs, ","))
	}
	return dirs
}

func (st *staticr) appStaticFile(requested string, s state.State) bool {
	exists := false
	for _, dir := range st.s.List("static_directories") {
		filepath.Walk(dir, func(path string, _ os.FileInfo, _ error) (err error) {
			if filepath.Base(path) == requested {
				f, _ := os.Open(path)
				serveStatic(s, f)
				exists = true
			}
			return err
		})
	}
	return exists
}

func (st *staticr) appAssetFile(requested string, s state.State) bool {
	exists := false
	f, err := st.a.GetAsset(requested)
	if err == nil {
		serveStatic(s, f)
		exists = true
	}
	return exists
}

func (st *staticr) Exists(s state.State, requested string) bool {
	exists := st.appStaticFile(requested, s)
	if !exists {
		exists = st.appAssetFile(requested, s)
	}
	return exists
}

func (st *staticr) StaticManage(s state.State) {
	if !st.Exists(s, requestedFile(s)) {
		abortStatic(s)
	} else {
		s.Call("header_now")
	}
}

func requestedFile(s state.State) string {
	rq := s.Request()
	return filepath.Base(rq.URL.Path)
}

func abortStatic(s state.State) {
	s.Call("abort", 404)
}

func serveStatic(s state.State, f http.File) {
	s.Call("serve_file", f)
}
