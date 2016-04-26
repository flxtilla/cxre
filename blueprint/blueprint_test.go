package blueprint_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/flxtilla/app"
	"github.com/flxtilla/cxre/blueprint"
	"github.com/flxtilla/cxre/route"
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

func testBlueprint(method string, a *app.App, b blueprint.Blueprint, t *testing.T) {
	var passed bool
	var passone bool
	var passmultiple []bool
	var inc int
	var incis bool

	bm0 := func(s state.State) {
		if inc == 0 {
			incis = true
			inc++
		} else {
			incis = false
		}
	}

	bm1 := func(s state.State) {
		if inc == 1 {
			incis = true
			inc++
		} else {
			incis = false
		}
		passone = true
		passmultiple = append(passmultiple, true)
	}

	bm2 := func(s state.State) {
		if inc == 2 {
			incis = true
			inc++
		} else {
			incis = false
		}
	}

	b.UseAt(0, bm0)

	b.Use(bm1, bm1, bm1)

	b.UseAt(5, bm2)

	m := func(s state.State) { passed = true }

	reflect.ValueOf(b).MethodByName(method).Call([]reflect.Value{reflect.ValueOf("/test_blueprint"), reflect.ValueOf(m)})

	a.Attach(b)

	a.Configure()

	expected := "/blueprint/test_blueprint"

	txst.ZeroExpectationPerformer(t, a, 200, method, expected).Perform()

	if passed != true {
		t.Errorf("%s blueprint route: %s was not invoked.", method, expected)
	}

	if inc != 3 || !incis {
		t.Errorf("Error setting and cycling through blueprint level Manage functions: %+v", b)
	}

	if passone != true {
		t.Errorf("Blueprint level Manage %#v was not invoked: %t.", bm1, passone)
	}

	if len(passmultiple) > 1 {
		t.Errorf("Blueprint level Manage %#v was duplicated.", bm1)
	}

	if passmultiple[0] != true {
		t.Errorf("Blueprint level Manage %#v was used in error.", bm1)
	}
}

func TestStandAloneBlueprint(t *testing.T) {
	for _, m := range txst.METHODS {
		a := AppForTest(t, "testStandAoneBlueprint")

		b := blueprint.New("/blueprint", blueprint.NewHandles(a.Handle), blueprint.NewMakes(a.StateFunction(a)))

		testBlueprint(m, a, b, t)
	}
}

func TestAppSpawnedBlueprint(t *testing.T) {
	for _, m := range txst.METHODS {
		a := AppForTest(t, "testAppSpawnedBlueprint")

		b := a.New("/blueprint")

		testBlueprint(m, a, b, t)
	}
}

func attachBlueprints(method string, t *testing.T) {
	var passed0, passed1, passed2 bool

	bm0 := func(c state.State) { passed0 = true }

	bm1 := func(c state.State) { passed1 = true }

	bm2 := func(c state.State) { passed2 = true }

	a := AppForTest(t, "testAttachBlueprints")

	b0 := blueprint.New("/", blueprint.NewHandles(a.Handle), blueprint.NewMakes(a.StateFunction(a)))

	zero := route.New(route.DefaultRouteConf(method, "/zero/:param", []state.Manage{bm0}))

	b0.Manage(zero)

	b1 := a.New("/blueprint")

	one := route.New(route.DefaultRouteConf(method, "/route/one", []state.Manage{bm1}))

	b1.Manage(one)

	b2 := blueprint.New("/blueprint", blueprint.NewHandles(a.Handle), blueprint.NewMakes(a.StateFunction(a)))

	two := route.New(route.DefaultRouteConf(method, "/route/two", []state.Manage{bm2}))

	b2.Manage(two)

	a.Attach(b0, b1, b2)

	a.Configure()

	txst.ZeroExpectationPerformer(t, a, 200, method, "/zero/test").Perform()
	txst.ZeroExpectationPerformer(t, a, 200, method, "/blueprint/route/one").Perform()
	txst.ZeroExpectationPerformer(t, a, 200, method, "/blueprint/route/two").Perform()

	if passed0 != true && passed1 != true && passed2 != true {
		t.Errorf("Blueprint routes were not merged properly.")
	}

	var paths []string

	bps := a.ListBlueprints()

	for _, bp := range bps {
		for _, rt := range bp.All() {
			paths = append(paths, rt.Path)
		}
	}

	existsIn := func(s string, l []string) bool {
		for _, x := range l {
			if s == x {
				return true
			}
		}
		return false
	}

	for _, expected := range []string{"/zero/:param", "/blueprint/route/one", "/blueprint/route/two"} {
		if !existsIn(expected, paths) {
			t.Errorf("Expected route with path %s was not found in added routes.", expected)
		}
	}
}

func TestBlueprintAttach(t *testing.T) {
	for _, m := range txst.METHODS {
		attachBlueprints(m, t)
	}
}

func chainBlueprints(method string, t *testing.T) {
	var x1, x2, x3 bool
	var y int
	a := AppForTest(t, "testChainedBlueprints")
	a.Use(func(c state.State) {
		x1 = true
		y = 1
	})
	b := a.New("/blueprintone")
	b.Use(func(c state.State) {
		if x1 {
			x2 = true
			y = 2
		}
	})
	c := b.New("/blueprinttwo")
	third := route.New(route.DefaultRouteConf(method, "/third", []state.Manage{func(c state.State) {}}))
	c.Manage(third)
	c.Use(func(c state.State) {
		if x1 && x2 {
			x3 = true
			y = 3
		}
	})
	a.Configure()
	txst.ZeroExpectationPerformer(t, a, 200, method, "/blueprintone/blueprinttwo/third").Perform()
	if !x1 && !x2 && !x3 && !(y == 3) {
		t.Errorf("Blueprint Manage chain error, chained test blueprint did not execute expected Manage.")
	}
}

func TestChainBlueprints(t *testing.T) {
	for _, m := range txst.METHODS {
		chainBlueprints(m, t)
	}
}

func mountBlueprint(method string, t *testing.T) {
	var passed bool

	a := AppForTest(t, "testMountBlueprint")

	b := blueprint.New("/mount", blueprint.NewHandles(a.Handle), blueprint.NewMakes(a.StateFunction(a)))

	m := func(c state.State) { passed = true }

	one := route.New(route.DefaultRouteConf(method, "/mounted/1", []state.Manage{m}))

	two := route.New(route.DefaultRouteConf(method, "/mounted/2", []state.Manage{m}))

	b.Manage(one)

	b.Manage(two)

	a.Mount("/test/one", b)

	a.Mount("/test/two", b)

	a.Attach(b)

	a.Configure()

	err := a.Mount("/cannot", b)

	if err == nil {
		t.Errorf("mounting a registered blueprint return no error")
	}

	perform := func(expected string, method string, app *app.App) {
		txst.ZeroExpectationPerformer(t, app, 200, method, expected).Perform()

		if passed == false {
			t.Errorf(fmt.Sprintf("%s blueprint route: %s was not invoked.", method, expected))
		}

		passed = false
	}

	perform("/mount/mounted/1", method, a)
	perform("/mount/mounted/2", method, a)
	perform("/test/one/mount/mounted/1", method, a)
	perform("/test/two/mount/mounted/1", method, a)
	perform("/test/one/mount/mounted/2", method, a)
	perform("/test/two/mount/mounted/2", method, a)
}

func TestMountBlueprint(t *testing.T) {
	for _, m := range txst.METHODS {
		mountBlueprint(m, t)
	}
}
