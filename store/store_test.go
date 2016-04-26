package store_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flxtilla/app"
	"github.com/flxtilla/cxre/state"
	"github.com/flxtilla/cxre/store"
	"github.com/flxtilla/cxre/store/test/resources"
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

func testLocation() string {
	wd, _ := os.Getwd()
	ld, _ := filepath.Abs(wd)
	return ld
}

func testConfFile() string {
	return filepath.Join(testLocation(), "test", "resources", "flotilla.conf")
}

func stored(s state.State) store.Store {
	var st store.Store
	if s, err := s.Call("store"); err == nil {
		st = s.(store.Store)
		return st
	}
	return nil
}

func TestStore(t *testing.T) {
	exp, _ := txst.NewExpectation(
		200,
		"GET",
		"/store",
		func(t *testing.T) state.Manage {
			return func(s state.State) {
				stor := stored(s)

				stor.Add("ADDED", "TRUE")

				adv := stor.Bool("ADDED")
				if adv != true {
					t.Errorf(`Store "ADDED" was not "true", but was %t`, adv)
				}

				noRead := stor.String("UNREAD_VALUE")
				if noRead != "" {
					t.Errorf(`Store item value exists, but should not.`)
				}

				df := struct {
					s   string
					b   bool
					f   float64
					i   int
					i64 int64
					l   []string
				}{
					stor.String("D"),
					stor.Bool("D"),
					stor.Float("D"),
					stor.Int("D"),
					stor.Int64("D"),
					stor.List("D"),
				}

				if df.s != "" || df.b != false || df.f != 0.0 || df.i != 0 || df.i64 != -1 || df.l != nil {
					t.Errorf(`Store item did not return correct default values: %+v`, df)
				}

				confString := stor.String("CONFSTRING")
				if confString != "ONE" {
					t.Errorf(`Store item value was not "ONE", but was %s`, confString)
				}

				bv := stor.Bool("CONFBOOL")
				if bv != true {
					t.Errorf(`Store "CONFBOOL" was not "true", but was %t`, bv)
				}

				fv := stor.Float("CONFFLOAT")
				if fv != 3.33333 {
					t.Errorf(`Store "CONFLOAT" was not "3.33333", but was %f`, fv)
				}

				iv := stor.Int("CONFINT")
				if iv != 2 {
					t.Errorf(`Store "CONFINT" value was not 2, but was %d`, iv)
				}

				iiv := stor.Int64("CONFINT64")
				if iiv != 99999 {
					t.Errorf(`Store "CONFINT64" value was not "99999", but was %s`, iiv)
				}

				sv := stor.String("SECTION_BLUE")
				if sv != "bondi" {
					t.Errorf(`Store item from section value was not "bondi", but was %s`, sv)
				}

				lv := stor.List("CONFLIST")
				have := strings.Join(lv, ",")
				expected := strings.Join([]string{"a", "b", "c", "d"}, ",")
				if bytes.Compare([]byte(have), []byte(expected)) != 0 {
					t.Errorf(`Store "CONFLIST" was not [a,b,c,d], but was [%s]`, have)
				}

				av := stor.String("CONFASSET")
				if av != "FROM_ASSET" {
					t.Errorf(`Store "CONFASSET" was not "FROM_ASSET", but was %s`, av)
				}
			}
		})

	a := AppForTest(
		t,
		"testStore",
		app.Assets(resources.ResourceFS),
	)
	a.Load(testConfFile())
	ac, _ := a.GetAssetByte("assets/bin.conf")
	a.LoadByte(ac, "bin.conf")
	a.Configure()

	txst.SimplePerformer(t, a, exp).Perform()
}
