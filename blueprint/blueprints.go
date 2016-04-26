package blueprint

import (
	"path/filepath"

	"github.com/flxtilla/cxre/state"
	"github.com/flxtilla/cxre/xrr"
)

// Blueprints is an anchor interface that implements a root Blueprint, in
// addition to methods for listing, determining existence, attachement, and
// mounting of other blueprints used by a flotilla application.
type Blueprints interface {
	Blueprint
	ListBlueprints() []Blueprint
	BlueprintExists(string) (Blueprint, bool)
	Attach(...Blueprint)
	Mount(string, ...Blueprint) error
}

type blueprints struct {
	Blueprint
}

// NewBlueprints returns a default Blueprints provided a string prefix, a
// HandleFn, and a state.Make function.
func NewBlueprints(prefix string, fn HandleFn, mk state.Make) Blueprints {
	return &blueprints{
		Blueprint: newBlueprint(prefix, NewHandles(fn), NewMakes(mk)),
	}
}

// ListBlueprints returns an array of attached and/or mounted Blueprint.
func (b *blueprints) ListBlueprints() []Blueprint {
	type IterC func(bs []Blueprint, fn IterC)

	var bps []Blueprint

	bps = append(bps, b)

	iter := func(bs []Blueprint, fn IterC) {
		for _, x := range bs {
			bps = append(bps, x)
			fn(x.Descendents(), fn)
		}
	}

	iter(b.Descendents(), iter)

	return bps
}

// Given a string prefix, BlueprintExists returns the Blueprint and boolean for
// existence of a Blueprint; if false Blueprint will be nil.
func (b *blueprints) BlueprintExists(prefix string) (Blueprint, bool) {
	for _, bp := range b.ListBlueprints() {
		if bp.Prefix() == prefix {
			return bp, true
		}
	}
	return nil, false
}

// Given any number of Blueprints, Attach integrates each with the App, either
// managing the Blueprint, or if the blueprint prefix exists, attaching routes
// to the appropriate Blueprint.
func (b *blueprints) Attach(blueprints ...Blueprint) {
	for _, blueprint := range blueprints {
		existing, exists := b.BlueprintExists(blueprint.Prefix())
		if !exists {
			blueprint.Register()
			b.Parent(blueprint)
		}
		if exists {
			for _, rt := range blueprint.Held() {
				existing.Manage(rt)
			}
		}
	}
}

var alreadyRegistered = xrr.NewXrror("only unregistered blueprints may be mounted; %s is already registered").Out

// Mount attaches each provided Blueprint to the given string mount point.
func (b *blueprints) Mount(point string, blueprints ...Blueprint) error {
	var bs []Blueprint
	for _, blueprint := range blueprints {
		if blueprint.Registered() {
			return alreadyRegistered(blueprint.Prefix)
		}

		newPrefix := filepath.ToSlash(filepath.Join(point, blueprint.Prefix()))

		nbp := newBlueprint(newPrefix, blueprint, blueprint)

		nbp.managers = combineManagers(b, blueprint.Managers())

		for _, rt := range blueprint.Held() {
			nbp.Manage(rt)
		}

		bs = append(bs, nbp)
	}
	b.Attach(bs...)
	return nil
}
