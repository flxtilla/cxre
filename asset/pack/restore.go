package pack

import (
	"fmt"
	"io"
)

func writeRestore(w io.Writer) error {
	_, err := fmt.Fprintf(w, `
// RestoreAsset restores an asset under the given directory
func (b *bindataFS) RestoreAsset(dir, name string) error {
	data, err := b.Asset(name)
	if err != nil {
		return err
	}
	info, err := b.AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

// RestoreAssets restores an asset under the given directory recursively
func (b *bindataFS) RestoreAssets(dir, name string) error {
	children, err := b.AssetDir(name)
	// File
	if err != nil {
		return b.RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = b.RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}
`)
	return err
}
