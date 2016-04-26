package route

import (
	"github.com/flxtilla/cxre/xrr"
)

// Routes is an interface for bundling multiple routes.
type Routes interface {
	GetRoute(string) (*Route, error)
	SetRoute(*Route)
	All() []*Route
	Map() map[string]*Route
}

// NewRoutes returns a package default Routes instance.
func NewRoutes() Routes {
	return &routes{
		r: make(map[string]*Route),
	}
}

type routes struct {
	r map[string]*Route
}

var nonExistent = xrr.NewXrror(`Route "%s" does not exist`).Out

// GetRoute returns a Route and an error, one or the other will be nil.
func (r *routes) GetRoute(key string) (*Route, error) {
	if rt, ok := r.r[key]; ok {
		return rt, nil
	}
	return nil, nonExistent(key)
}

// SetRoute adds the provided Route instance to this default Routes.
func (r *routes) SetRoute(rt *Route) {
	r.r[rt.Name()] = rt
}

// All returns a slice of all Route instances managed by this default Routes.
func (r *routes) All() []*Route {
	var ret []*Route
	for _, v := range r.r {
		ret = append(ret, v)
	}
	return ret
}

// Map returns a map of string keys to Route instances contained in this
// default Routes.
func (r *routes) Map() map[string]*Route {
	return r.r
}
