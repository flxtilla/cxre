package asset_test

import (
	"testing"

	"github.com/flxtilla/app"
	"github.com/flxtilla/cxre/asset/test/resources"
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

func TestAssets(t *testing.T) {
	a := AppForTest(t, "assetsTest", app.Assets(resources.ResourceFS))

	var err error

	_, err = a.GetAsset("test_asset.html")
	if err != nil {
		t.Errorf("GetAsset error: %s", err.Error())
	}

	_, err = a.GetAsset("does_not_exist")
	if err == nil {
		t.Error("GetAsset error: error was nil where it should not be")
	}

	_, err = a.GetAssetByte("assets/templates/test_asset.html")
	if err != nil {
		t.Errorf("GetAssetByte error: %s", err.Error())
	}

	_, err = a.GetAssetByte("test_asset.html")
	if err == nil {
		t.Error("GetAssetByte error: error was nil, but should not be")
	}

	l := a.ListAssetFS()[0]
	ll := len(l.AssetNames())
	if ll != 4 {
		t.Errorf("Asset list length is %d, not 4: %s", ll, l)
	}
}
