package pack

import (
	"fmt"
	"io"
)

func writeDebug(w io.Writer, c *Config, toc []Asset) error {
	err := writeDebugHeader(w)
	if err != nil {
		return err
	}

	for i := range toc {
		err = writeDebugAsset(w, c, &toc[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func writeDebugHeader(w io.Writer) error {
	_, err := fmt.Fprintf(w, `import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)
// bindataRead reads the given file from disk. It returns an error on failure.
func bindataRead(path, name string) ([]byte, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset %%s at %%s: %%v", name, path, err)
	}
	return buf, err
}
type asset struct {
	bytes []byte
	info  os.FileInfo
}
`)
	return err
}

func writeDebugAsset(w io.Writer, c *Config, asset *Asset) error {
	pathExpr := fmt.Sprintf("%q", asset.Path)
	if c.Dev {
		pathExpr = fmt.Sprintf("filepath.Join(rootDir, %q)", asset.Name)
	}

	_, err := fmt.Fprintf(w, `// %s reads file data from disk. It returns an error on failure.
func %s() (*asset, error) {
	path := %s
	name := %q
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}
	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %%s at %%s: %%v", name, path, err)
	}
	a := &asset{bytes: bytes, info: fi}
	return a, err
}
`, asset.Func, asset.Func, pathExpr, asset.Name)
	return err
}
