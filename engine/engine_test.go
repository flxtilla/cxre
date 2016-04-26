package engine

import (
	"errors"
	"net/http"
	"reflect"
	"testing"
)

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(int) {}

func TestRouter(t *testing.T) {
	t.Parallel()

	engine := DefaultEngine(func(http.ResponseWriter, *http.Request, *Result) {})

	routed := false

	engine.Handle("GET", "/user/:name", func(rw http.ResponseWriter, rq *http.Request, rs *Result) {
		routed = true
		want := Params{Param{"name", "gopher"}}
		if !reflect.DeepEqual(rs.Params, want) {
			t.Fatalf("wrong wildcard values: want %v, got %v", want, rs.Params)
		}
	})

	w := new(mockResponseWriter)

	req, _ := http.NewRequest("GET", "/user/gopher", nil)
	engine.ServeHTTP(w, req)

	if !routed {
		t.Fatal("routing failed")
	}
}

type handlerStruct struct {
	handeled *bool
}

func (h handlerStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	*h.handeled = true
}

func TestRouterAPI(t *testing.T) {
	var get, post, put, patch, delete bool
	//var handler, handlerFunc bool

	//httpHandler := handlerStruct{&handler}

	router := DefaultEngine(func(http.ResponseWriter, *http.Request, *Result) {})

	router.Handle("GET", "/GET", func(rw http.ResponseWriter, rq *http.Request, rs *Result) {
		get = true
	})
	router.Handle("POST", "/POST", func(rw http.ResponseWriter, rq *http.Request, rs *Result) {
		post = true
	})
	router.Handle("PUT", "/PUT", func(rw http.ResponseWriter, rq *http.Request, rs *Result) {
		put = true
	})
	router.Handle("PATCH", "/PATCH", func(rw http.ResponseWriter, rq *http.Request, rs *Result) {
		patch = true
	})
	router.Handle("DELETE", "/DELETE", func(rw http.ResponseWriter, rq *http.Request, rs *Result) {
		delete = true
	})

	//router.Handler("GET", "/Handler", httpHandler)
	//router.HandlerFunc("GET", "/HandlerFunc", func(w http.ResponseWriter, r *http.Request) {
	//	handlerFunc = true
	//})

	w := new(mockResponseWriter)

	r, _ := http.NewRequest("GET", "/GET", nil)
	router.ServeHTTP(w, r)
	if !get {
		t.Error("routing GET failed")
	}

	r, _ = http.NewRequest("POST", "/POST", nil)
	router.ServeHTTP(w, r)
	if !post {
		t.Error("routing POST failed")
	}

	r, _ = http.NewRequest("PUT", "/PUT", nil)
	router.ServeHTTP(w, r)
	if !put {
		t.Error("routing PUT failed")
	}

	r, _ = http.NewRequest("PATCH", "/PATCH", nil)
	router.ServeHTTP(w, r)
	if !patch {
		t.Error("routing PATCH failed")
	}

	r, _ = http.NewRequest("DELETE", "/DELETE", nil)
	router.ServeHTTP(w, r)
	if !delete {
		t.Error("routing DELETE failed")
	}

	//r, _ = http.NewRequest("GET", "/Handler", nil)
	//router.ServeHTTP(w, r)
	//if !handler {
	//	t.Error("routing Handler failed")
	//}

	//r, _ = http.NewRequest("GET", "/HandlerFunc", nil)
	//router.ServeHTTP(w, r)
	//if !handlerFunc {
	//	t.Error("routing HandlerFunc failed")
	//}
}

func TestRouterRoot(t *testing.T) {
	router := DefaultEngine(func(http.ResponseWriter, *http.Request, *Result) {})
	recv := catchPanic(func() {
		router.Handle("GET", "noSlashRoot", nil)
	})
	if recv == nil {
		t.Fatal("registering path not beginning with '/' did not panic")
	}
}

func TestRouterLookup(t *testing.T) {
	routed := false
	wantHandle := func(_ http.ResponseWriter, _ *http.Request, _ *Result) {
		routed = true
	}
	wantParams := Params{Param{"name", "gopher"}}

	router := DefaultEngine(func(http.ResponseWriter, *http.Request, *Result) {})

	// try empty router first
	rslt := router.lookup("GET", "/nope")
	if rslt.Code != 404 {
		t.Fatalf("Got Result for a registered Rule, not a status: %+v", rslt)
	}
	if rslt.TSR {
		t.Error("Got wrong TSR recommendation!")
	}

	// insert route and try again
	router.Handle("GET", "/user/:name", wantHandle)

	rslt = router.lookup("GET", "/user/gopher")
	if rslt.Rule == nil {
		t.Fatal("Got no handle!")
	} else {
		rslt.Rule(nil, nil, nil)
		if !routed {
			t.Fatal("Routing failed!")
		}
	}

	if !reflect.DeepEqual(rslt.Params, wantParams) {
		t.Fatalf("Wrong parameter values: want %v, got %v", wantParams, rslt.Params)
	}

	rslt = router.lookup("GET", "/user/gopher/")
	if rslt.Code != 301 {
		t.Fatalf("Got Result for a registered Rule, not a status: %+v", rslt)
	}
	if !rslt.TSR {
		t.Error("Got no TSR recommendation!")
	}

	rslt = router.lookup("GET", "/nope")
	if rslt.Code != 404 {
		t.Fatalf("Got Result for a registered Rule, not a status: %+v", rslt)
	}
	if rslt.TSR {
		t.Error("Got wrong TSR recommendation!")
	}
}

type mockFileSystem struct {
	opened bool
}

func (mfs *mockFileSystem) Open(name string) (http.File, error) {
	mfs.opened = true
	return nil, errors.New("this is just a mock")
}
