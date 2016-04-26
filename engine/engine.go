package engine

import (
	"fmt"
	"net/http"

	"github.com/flxtilla/cxre/xrr"
)

// A routing Rule requires a http.ResponseWriter, *http.Request, and a Result
// instance.
type Rule func(http.ResponseWriter, *http.Request, *Result)

// The Engine interface encapsulates routing management and satisfies the
// net/http ServeHTTP interface function.
type Engine interface {
	Handle(string, string, Rule)
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type conf struct {
	RedirectTrailingSlash bool
	RedirectFixedPath     bool
}

type engine struct {
	*conf
	trees      map[string]*node
	StatusRule Rule
}

func defaultConf() *conf {
	return &conf{
		RedirectTrailingSlash: true,
		RedirectFixedPath:     true,
	}
}

// Provided a default status Rule, DefaultEngine returns a default engine
// instance.
func DefaultEngine(status Rule) *engine {
	return &engine{
		conf:       defaultConf(),
		StatusRule: status,
	}
}

// The default engine Handle function takes method string, a path string, and a
// Rule.
func (e *engine) Handle(method string, path string, r Rule) {
	if method != "STATUS" && path[0] != '/' {
		panic("path must begin with '/'")
	}

	if method == "STATUS" && path == "DEFAULT" {
		e.StatusRule = r
	}

	if e.trees == nil {
		e.trees = make(map[string]*node)
	}

	root := e.trees[method]

	if root == nil {
		root = new(node)
		e.trees[method] = root
	}

	root.addRoute(path, r)
}

func (e *engine) lookup(method, path string) *Result {
	if root := e.trees[method]; root != nil {
		if rule, params, tsr := root.getValue(path); rule != nil {
			return NewResult(200, rule, params, tsr)
		} else if method != "CONNECT" && path != "/" {
			code := 301
			if method != "GET" {
				code = 307
			}
			if tsr && e.RedirectTrailingSlash {
				var newpath string
				if path[len(path)-1] == '/' {
					newpath = path[:len(path)-1]
				} else {
					newpath = path + "/"
				}
				return NewResult(code, func(rw http.ResponseWriter, rq *http.Request, rs *Result) {
					rq.URL.Path = newpath
					http.Redirect(rw, rq, rq.URL.String(), code)
				}, nil, tsr)
			}
			if e.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					e.RedirectTrailingSlash,
				)
				if found {
					return NewResult(code, func(rw http.ResponseWriter, rq *http.Request, rs *Result) {
						rq.URL.Path = string(fixedPath)
						http.Redirect(rw, rq, rq.URL.String(), code)
					}, nil, tsr)
				}
			}
		}
	}
	for method := range e.trees {
		if method == method {
			continue
		}
		handle, _, _ := e.trees[method].getValue(path)
		if handle != nil {
			return e.status(statusPath(405, path))
		}
	}
	return e.status(statusPath(404, path))
}

func statusPath(code int, path string) (int, string) {
	return code, fmt.Sprintf("/%d/%s", code, path)
}

func (e *engine) status(code int, path string) *Result {
	if root := e.trees["STATUS"]; root != nil {
		if rule, params, tsr := root.getValue(path); rule != nil {
			return NewResult(code, rule, params, tsr)
		}
	}
	return NewResult(code, e.defaultStatus(code), nil, false)
}

func defaultStatusRule(rw http.ResponseWriter, rq *http.Request, rs *Result) {
	rw.WriteHeader(rs.Code())
	rw.Write([]byte(fmt.Sprintf("%d %s", rs.Code, http.StatusText(rs.Code()))))
}

func (e *engine) defaultStatus(code int) Rule {
	if e.StatusRule == nil {
		return defaultStatusRule
	}
	return e.StatusRule
}

func (e *engine) rcvr(rw http.ResponseWriter, rq *http.Request) {
	if rcv := recover(); rcv != nil {
		s := e.status(500, rq.URL.Path)
		s.Xrror("%s", xrr.ErrorTypePanic, xrr.Stack(3), rcv)
		s.Rule(rw, rq, s)
	}
}

// The default engine ServeHTTP function.
func (e *engine) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	defer e.rcvr(rw, rq)
	rslt := e.lookup(rq.Method, rq.URL.Path)
	rslt.Rule(rw, rq, rslt)
}
