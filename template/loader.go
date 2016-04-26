package template

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/flxtilla/cxre/xrr"
)

type loader struct {
	*templater
	FileExtensions []string
}

func newLoader(t *templater) *loader {
	return &loader{
		templater:      t,
		FileExtensions: []string{".html", ".dji"},
	}
}

// ValidFileExtension returns a boolean for extension provided indicating if
// the Loader allows the extension type.
func (l *loader) ValidFileExtension(ext string) bool {
	for _, extension := range l.FileExtensions {
		if extension == ext {
			return true
		}
	}
	return false
}

func (l *loader) assetTemplates() []string {
	var ret []string
	for _, assetfs := range l.a.ListAssetFS() {
		for _, f := range assetfs.AssetNames() {
			if l.ValidFileExtension(filepath.Ext(f)) {
				ret = append(ret, f)
			}
		}
	}
	return ret
}

// ListTemplates lists templates in the Loader.
func (l *loader) ListTemplates() []string {
	var ret []string
	for _, p := range l.s.List("template_directories") {
		files, _ := ioutil.ReadDir(p)
		for _, f := range files {
			if l.ValidFileExtension(filepath.Ext(f.Name())) {
				ret = append(ret, fmt.Sprintf("%s/%s", p, f.Name()))
			}
		}
	}
	ret = append(ret, l.assetTemplates()...)
	return ret
}

var TemplateDoesNotExist = xrr.NewXrror("Template %s does not exist.").Out

// Load a template by string name from the flotilla Loader.
func (l *loader) Load(name string) (string, error) {
	for _, p := range l.s.List("template_directories") {
		f := filepath.Join(p, name)
		if l.ValidFileExtension(filepath.Ext(f)) {
			if _, err := os.Stat(f); err == nil {
				file, err := os.Open(f)
				r, err := ioutil.ReadAll(file)
				return string(r), err
			}
			if r, err := l.a.GetAsset(name); err == nil {
				r, err := ioutil.ReadAll(r)
				return string(r), err
			}
		}
	}
	return "", TemplateDoesNotExist(name)
}
