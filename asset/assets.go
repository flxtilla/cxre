package asset

import "net/http"

// AssetFS is an interface to a grouping of binary assets created by the pack
// tool providing functionality for retrieving contained assets and information
// about contained assets.
type AssetFS interface {
	Asset(string) ([]byte, error)
	AssetHttp(string) (http.File, error)
	AssetDir(string) ([]string, error)
	AssetNames() []string
}

// Assets is an interface to any grouping of AssetsFS, with functionality for
// accessing specific assets within this grouping, and adding additional
// AssetFS.
type Assets interface {
	GetAsset(string) (http.File, error)
	GetAssetByte(string) ([]byte, error)
	SetAssetFS(...AssetFS)
	ListAssetFS() []AssetFS
}

// New provides a default assets instance satisfying the Assets interface, with
// provided AssetFSs.
func New(af ...AssetFS) Assets {
	as := &assets{
		a: make([]AssetFS, 0),
	}
	as.SetAssetFS(af...)
	return as
}

type assets struct {
	a []AssetFS
}

var assetUnavailable = Xrror("Asset %s unavailable").Out

// The default assets GetAssets function takes a string name and returns an
// http.File version of the assset if it exists and an error.
func (a *assets) GetAsset(requested string) (http.File, error) {
	for _, x := range a.a {
		f, err := x.AssetHttp(requested)
		if err == nil {
			return f, nil
		}
	}
	return nil, assetUnavailable(requested)
}

// The default assets GetAssetByte function takes a string name and returns a
// byte representation of the asset if it exists and error.
func (a *assets) GetAssetByte(requested string) ([]byte, error) {
	for _, x := range a.a {
		b, err := x.Asset(requested)
		if err == nil {
			return b, nil
		}
	}
	return nil, assetUnavailable(requested)
}

// The default assets SetAssetFS function takes any number of AssetFS to add to
// the assets instance.
func (a *assets) SetAssetFS(af ...AssetFS) {
	a.a = append(a.a, af...)
}

// The default assets ListAssetFS function returns a slice of AssetFS contained
// in the assets instance.
func (a *assets) ListAssetFS() []AssetFS {
	return a.a
}
