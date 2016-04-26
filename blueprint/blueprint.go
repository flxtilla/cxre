package blueprint

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/flxtilla/cxre/engine"
	"github.com/flxtilla/cxre/route"
	"github.com/flxtilla/cxre/state"
	"github.com/flxtilla/cxre/static"
	"github.com/flxtilla/cxre/status"
)

// Blueprint is an interface for common route bundling in a flotilla app.
type Blueprint interface {
	SetupState
	Handles
	Makes
	Prefix() string
	route.Routes
	Exists(*route.Route) bool
	New(string, ...state.Manage) Blueprint
	Use(...state.Manage)
	UseAt(int, ...state.Manage)
	Managers() []state.Manage
	Parent(...Blueprint)
	Descendents() []Blueprint
	MethodManager
	status.Statusr
}

type SetupState interface {
	Register()
	Registered() bool
	Held() []*route.Route
}

type setupstate struct {
	registered bool
	deferred   []func()
	held       []*route.Route
}

func (s *setupstate) Register() {
	s.runDeferred()
	s.registered = true
}

func (s *setupstate) runDeferred() {
	for _, fn := range s.deferred {
		fn()
	}
	s.deferred = nil
}

// The default setupstate Registered function, returning a boolean value.
func (s *setupstate) Registered() bool {
	return s.registered
}

// The default setupstate Held function, returning a slice of Route currently held.
func (s *setupstate) Held() []*route.Route {
	return s.held
}

// A HandleFn is any function taking a string method, a string path, and an
// engine.Rule.
type HandleFn func(string, string, engine.Rule)

// The Handles interface provides a Handling function of the HandleFn type that
// may be used and/or passed around by a Blueprint.
type Handles interface {
	Handling(string, string, engine.Rule)
}

type handles struct {
	handle HandleFn
}

// NewHandles returns a default Handles interface, provided a HandleFn.
func NewHandles(hf HandleFn) Handles {
	return &handles{hf}
}

// The default Handles Handling function.
func (h *handles) Handling(method string, path string, rule engine.Rule) {
	h.handle(method, path, rule)
}

// The Makes interface provides a Making function that returns a state.Make
// function to be used and/or passed around by a Blueprint.
type Makes interface {
	Making() state.Make
}

type makes struct {
	makes state.Make
}

// NewMakes returns a default Makes interface, provided a state.Make function.
func NewMakes(m state.Make) Makes {
	return &makes{m}
}

// The default makes Making function, returning a state.Make function.
func (m *makes) Making() state.Make {
	return m.makes
}

type blueprint struct {
	*setupstate
	Handles
	Makes
	prefix      string
	descendents []Blueprint
	route.Routes
	managers []state.Manage
	MethodManager
	status.Statusr
}

// New returns a new Blueprint with the provided string prefix, Handles & Makes
// interfaces.
func New(prefix string, h Handles, m Makes) Blueprint {
	return newBlueprint(prefix, h, m)
}

func newBlueprint(prefix string, h Handles, m Makes) *blueprint {
	return &blueprint{
		setupstate: &setupstate{},
		Handles:    h,
		Makes:      m,
		prefix:     prefix,
		Routes:     route.NewRoutes(),
		Statusr:    status.New(m),
	}
}

// The default blueprint Prefix returns a string useful for identification or
// path building.
func (b *blueprint) Prefix() string {
	return b.prefix
}

// The default blueprint Exists returns a boolean indicating if the provided
// Route is managed by the Blueprint.
func (b *blueprint) Exists(rt *route.Route) bool {
	for _, r := range b.Routes.All() {
		if (rt.Path == r.Path) && (rt.Method == r.Method) {
			return true
		}
	}
	return false
}

func (b *blueprint) pathFor(path string) string {
	joined := filepath.ToSlash(filepath.Join(b.prefix, path))
	// Append a '/' if the last component had one, but only if it's not there already
	if len(path) > 0 && path[len(path)-1] == '/' && joined[len(joined)-1] != '/' {
		return joined + "/"
	}
	return joined
}

func combineManagers(b Blueprint, managers []state.Manage) []state.Manage {
	h := make([]state.Manage, 0)
	h = append(h, b.Managers()...)
	for _, manage := range managers {
		if !manageExists(h, manage) {
			h = append(h, manage)
		}
	}
	return h
}

// New returns a new Blueprint as a child of the parent Blueprint.
func (b *blueprint) New(component string, managers ...state.Manage) Blueprint {
	prefix := b.pathFor(component)
	newb := newBlueprint(prefix, b.Handles, b.Makes)
	newb.managers = combineManagers(b, managers)
	b.descendents = append(b.descendents, newb)
	return newb
}

func isFunc(fn interface{}) bool {
	return reflect.ValueOf(fn).Kind() == reflect.Func
}

func equalFunc(a, b interface{}) bool {
	if !isFunc(a) || !isFunc(b) {
		panic("flotilla : funcEqual -- type error!")
	}
	av := reflect.ValueOf(&a).Elem()
	bv := reflect.ValueOf(&b).Elem()
	return av.InterfaceData() == bv.InterfaceData()
}

func manageExists(inside []state.Manage, outside state.Manage) bool {
	for _, fn := range inside {
		if equalFunc(fn, outside) {
			return true
		}
	}
	return false
}

// The default blueprint Use function, takes any number of state.Manage
// functions. Any route managed by the Blueprint will use these functions, and
// use them before the Routes own state functions.
func (b *blueprint) Use(managers ...state.Manage) {
	for _, manage := range managers {
		if !manageExists(b.managers, manage) {
			b.managers = append(b.managers, manage)
		}
	}
}

// The default blueprint UseAt function, taking an integer index and any number
// of state.Manage functions. Any route managed by the Blueprint will use these
// functions, and use them before the Routes own state functions. Added state
// functions are placed at the index provided in the existing Blueprint state
// function list, or at the end if the index does not exist.
func (b *blueprint) UseAt(index int, managers ...state.Manage) {
	if index > len(b.managers) {
		b.Use(managers...)
		return
	}

	var newh []state.Manage

	for _, manage := range managers {
		if !manageExists(b.managers, manage) {
			newh = append(newh, manage)
		}
	}

	before := b.managers[:index]
	after := append(newh, b.managers[index:]...)
	b.managers = append(before, after...)
}

func (b *blueprint) add(r *route.Route) {
	b.Routes.SetRoute(r)
}

func (b *blueprint) hold(r *route.Route) {
	b.held = append(b.held, r)
}

func (b *blueprint) push(register func(), rt *route.Route) {
	if b.registered {
		register()
	} else {
		if rt != nil {
			b.hold(rt)
		}
		b.deferred = append(b.deferred, register)
	}
}

// The default blueprint Parent function will add the provided blueprints as
// descendents of the Blueprint.
func (b *blueprint) Parent(bs ...Blueprint) {
	b.descendents = append(b.descendents, bs...)
}

// The default blueprint Descendents function returns an array of Blueprint as
// direct decendents of the calling Blueprint.
func (b *blueprint) Descendents() []Blueprint {
	return b.descendents
}

// Teh default blueprint Managers function returns an array of state.Manage
// functions attached to the Blueprint.
func (b *blueprint) Managers() []state.Manage {
	return b.managers
}

func reManage(rt *route.Route, b Blueprint) {
	var ms []state.Manage
	existing := rt.Managers
	rt.Managers = nil
	ms = append(ms, combineManagers(b, existing)...)
	rt.Managers = ms
}

func registerRouteConf(b *blueprint) route.RouteConf {
	return func(rt *route.Route) error {
		reManage(rt, b)
		rt.Path = b.pathFor(rt.Base)
		rt.Registered = true
		rt.Makes = b.Makes
		return nil
	}
}

// MethodManager provides an interface for route management.
type MethodManager interface {
	Manage(rt *route.Route)
	GET(string, ...state.Manage)
	POST(string, ...state.Manage)
	DELETE(string, ...state.Manage)
	PATCH(string, ...state.Manage)
	PUT(string, ...state.Manage)
	OPTIONS(string, ...state.Manage)
	HEAD(string, ...state.Manage)
	STATIC(static.Static, string, ...string)
	STATUS(int, ...state.Manage)
}

// The default blueprint method manager Manage function adds a route to the Blueprint.
func (b *blueprint) Manage(rt *route.Route) {
	register := func() {
		rt.Configure(registerRouteConf(b))
		if !b.Exists(rt) {
			rt.Managers = append([]state.Manage{b.statusExtension}, rt.Managers...)
			b.add(rt)
			b.Handling(rt.Method, rt.Path, rt.Rule)
		}
	}
	b.push(register, rt)
}

// The default blueprint GET method manager function.
func (b *blueprint) GET(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("GET", path, managers)))
}

// The default blueprint method manager POST function.
func (b *blueprint) POST(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("POST", path, managers)))
}

// The default blueprint method manager DELETE function.
func (b *blueprint) DELETE(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("DELETE", path, managers)))
}

// The default blueprint method manager PATCH function.
func (b *blueprint) PATCH(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("PATCH", path, managers)))
}

// The default blueprint method manager PUT function.
func (b *blueprint) PUT(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("PUT", path, managers)))
}

// The default blueprint method manager OPTIONS function.
func (b *blueprint) OPTIONS(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("OPTIONS", path, managers)))
}

// The default blueprint method manager HEAD function.
func (b *blueprint) HEAD(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("HEAD", path, managers)))
}

func dropTrailing(path string, trailing string) string {
	if fp := strings.Split(path, "/"); fp[len(fp)-1] == trailing {
		return strings.Join(fp[0:len(fp)-1], "/")
	}
	return path
}

// The default blueprint STATIC function, taking a static.Static interface,
// path string, and any number of directories.
func (b *blueprint) STATIC(s static.Static, path string, dirs ...string) {
	if len(dirs) > 0 {
		for _, dir := range dirs {
			b.push(func() { s.StaticDirs(dropTrailing(dir, "*filepath")) }, nil)
		}
	}
	register := func() {
		rt := route.New(route.StaticRouteConf("GET", path, []state.Manage{s.StaticManage}))
		rt.Configure(registerRouteConf(b))
		b.add(rt)
		b.Handling(rt.Method, rt.Path, rt.Rule)
	}
	b.push(register, nil)
}

func formatStatusPath(code, prefix string) string {
	if prefix == "/" {
		return fmt.Sprintf("/%s/*filepath", code)
	}
	return fmt.Sprintf("/%s/%s/*filepath", code, prefix)
}

func (b *blueprint) statusExtension(s state.State) {
	s.Extend(b.StateStatusExtension())
}

// The default blueprint STATUS function, taking an integer status code and any
// number of state.Manage functions.
func (b *blueprint) STATUS(code int, managers ...state.Manage) {
	b.SetRawStatus(code, managers...)
	b.push(func() {
		b.Handling(
			"STATUS",
			formatStatusPath(strconv.Itoa(code), b.prefix),
			b.StatusRule(),
		)
	},
		nil)
}
