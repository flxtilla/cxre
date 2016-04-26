package route

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flxtilla/cxre/engine"
	"github.com/flxtilla/cxre/state"
)

type Makes interface {
	Making() state.Make
}

type Route struct {
	name, Method, Base, Path string
	Registered, Static       bool
	Managers                 []state.Manage
	Makes
}

// New returns a Route instance with the given configuration.
func New(conf ...RouteConf) *Route {
	rt := &Route{}
	err := rt.Configure(conf...)
	if err != nil {
		panic(fmt.Sprintf("route configuration error: %s", err.Error()))
	}
	return rt
}

// RouteConf is a route configuration function type taking a Route instance and
// returning an error.
type RouteConf func(*Route) error

// Configure runs the given functions to configure the Route instance.
func (rt *Route) Configure(conf ...RouteConf) error {
	var err error
	for _, c := range conf {
		err = c(rt)
	}
	return err
}

// DefaultRouteConf returns a route configuration function for method, path,
// and managers.
func DefaultRouteConf(method string, path string, managers []state.Manage) RouteConf {
	return func(rt *Route) error {
		rt.Method = method
		rt.Base = path
		rt.Managers = managers
		return nil
	}
}

// StaticRouteConf returns default route configuration function for a static
// route method, path, and managers.
func StaticRouteConf(method string, path string, managers []state.Manage) RouteConf {
	return func(rt *Route) error {
		rt.Method = method
		rt.Static = true
		if fp := strings.Split(path, "/"); fp[len(fp)-1] != "*filepath" {
			rt.Base = filepath.ToSlash(filepath.Join(path, "/*filepath"))
		} else {
			rt.Base = path
		}
		rt.Managers = managers
		return nil
	}
}

// Rule provides the engine Rule for the route.
func (rt *Route) Rule(rw http.ResponseWriter, rq *http.Request, rs *engine.Result) {
	stateFn := rt.Making()
	state := stateFn(rw, rq, rs, rt.Managers)
	state.Run()
	state.Cancel()
}

// Name returns the route name.
func (rt *Route) Name() string {
	if rt.name == "" {
		return Named(rt)
	}
	return rt.name
}

// Rename will rename the Route to the provided string.
func (rt *Route) Rename(name string) {
	rt.name = name
}

// Given a Route instance, Named will return a route name based on the route's
// paths & method.
func Named(rt *Route) string {
	n := strings.Split(rt.Path, "/")
	n = append(n, strings.ToLower(rt.Method))
	for index, value := range n {
		if regSplat.MatchString(value) {
			n[index] = "{s}"
		}
		if regParam.MatchString(value) {
			n[index] = "{p}"
		}
	}
	return strings.Join(n, `\`)
}

var regParam = regexp.MustCompile(`:[^/#?()\.\\]+|\(\?P<[a-zA-Z0-9]+>.*\)`)
var regSplat = regexp.MustCompile(`\*[^/#?()\.\\]+|\(\?P<[a-zA-Z0-9]+>.*\)`)

// Url returns a *url.Url for the route, provided the string parameters.
func (rt *Route) Url(params ...string) (*url.URL, error) {
	paramCount := len(params)
	i := 0
	rurl := regParam.ReplaceAllStringFunc(rt.Path, func(m string) string {
		var val string
		if i < paramCount {
			val = params[i]
		}
		i += 1
		return fmt.Sprintf(`%s`, val)
	})
	rurl = regSplat.ReplaceAllStringFunc(rurl, func(m string) string {
		splat := params[i:(len(params))]
		i += len(splat)
		return fmt.Sprintf(strings.Join(splat, "/"))
	})
	u, err := url.Parse(rurl)
	if err != nil {
		return nil, err
	}
	if i < len(params) && rt.Method == "GET" {
		providedquerystring := params[i:(len(params))]
		var querystring []string
		qsi := 0
		for qi, qs := range providedquerystring {
			if len(strings.Split(qs, "=")) != 2 {
				qs = fmt.Sprintf("value%d=%s", qi+1, qs)
			}
			querystring = append(querystring, url.QueryEscape(qs))
			qsi += 1
		}
		u.RawQuery = strings.Join(querystring, "&")
	}
	return u, nil
}
