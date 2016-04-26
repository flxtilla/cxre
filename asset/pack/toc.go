package pack

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type assetTree struct {
	Asset    Asset
	Children map[string]*assetTree
}

func newAssetTree() *assetTree {
	tree := &assetTree{}
	tree.Children = make(map[string]*assetTree)
	return tree
}

func (node *assetTree) child(name string) *assetTree {
	rv, ok := node.Children[name]
	if !ok {
		rv = newAssetTree()
		node.Children[name] = rv
	}
	return rv
}

func (root *assetTree) Add(route []string, asset Asset) {
	for _, name := range route {
		root = root.child(name)
	}
	root.Asset = asset
}

func ident(w io.Writer, n int) {
	for i := 0; i < n; i++ {
		w.Write([]byte{'\t'})
	}
}

func (root *assetTree) funcOrNil() string {
	if root.Asset.Func == "" {
		return "nil"
	} else {
		return root.Asset.Func
	}
}

func (root *assetTree) writeGoMap(w io.Writer, nident int) {
	fmt.Fprintf(w, "&bintree{%s, map[string]*bintree{", root.funcOrNil())

	if len(root.Children) > 0 {
		io.WriteString(w, "\n")

		// Sort to make output stable between invocations
		filenames := make([]string, len(root.Children))
		i := 0
		for filename, _ := range root.Children {
			filenames[i] = filename
			i++
		}
		sort.Strings(filenames)

		for _, p := range filenames {
			ident(w, nident+1)
			fmt.Fprintf(w, `"%s": `, p)
			root.Children[p].writeGoMap(w, nident+1)
		}
		ident(w, nident)
	}

	io.WriteString(w, "}}")
	if nident > 0 {
		io.WriteString(w, ",")
	}
	io.WriteString(w, "\n")
}

func (root *assetTree) WriteAsGoMap(w io.Writer) error {
	_, err := fmt.Fprint(w, `type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = `)
	root.writeGoMap(w, 0)
	return err
}

func writeTOCTree(w io.Writer, toc []Asset) error {
	_, err := fmt.Fprintf(w, `// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func (b *bindataFS) AssetDir(name string) ([]string, error) {
	node := b.tree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %%s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %%s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

`)
	if err != nil {
		return err
	}
	tree := newAssetTree()
	for i := range toc {
		pathList := strings.Split(toc[i].Name, "/")
		tree.Add(pathList, toc[i])
	}
	return tree.WriteAsGoMap(w)
}

// writeTOC writes the table of contents file.
func writeTOC(w io.Writer, c *Config, toc []Asset) error {
	err := writeTOCHeader(w, c)
	if err != nil {
		return err
	}

	for i := range toc {
		err = writeTOCAsset(w, &toc[i])
		if err != nil {
			return err
		}
	}

	return writeTOCFooter(w)
}

// writeTOCHeader writes the table of contents file header.
func writeTOCHeader(w io.Writer, c *Config) error {
	_, err := fmt.Fprintf(w, `var %s *bindataFS = &bindataFS{
	prefix: %q,
	tree:   _bintree,
	data:   _bindata,
}

type bindataFS struct {
	prefix string
	tree   *bintree
	data   bindata
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func (b *bindataFS) Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := b.data[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %%s can't read by error: %%v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %%s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func (b *bindataFS) MustAsset(name string) []byte {
	a, err := b.Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func (b *bindataFS) AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := b.data[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %%s can't read by error: %%v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %%s not found", name)
}

// AssetNames returns the names of the assets.
func (b *bindataFS) AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

type bindata map[string]func() (*asset, error)

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = bindata{
`, c.FSName, c.FSPrefix)
	return err
}

// writeTOCAsset write a TOC entry for the given asset.
func writeTOCAsset(w io.Writer, asset *Asset) error {
	_, err := fmt.Fprintf(w, "\t%q: %s,\n", asset.Name, asset.Func)
	return err
}

// writeTOCFooter writes the table of contents file footer.
func writeTOCFooter(w io.Writer) error {
	_, err := fmt.Fprintf(w, `}

`)
	return err
}

func WriteTOCPost(w io.Writer) error {
	_, err := fmt.Fprintf(w, `func (b *bindataFS) HasAsset(requested string) (string, bool) {
	for _, filename := range b.AssetNames() {
		if path.Base(filename) == requested {
			return filename, true
		}
	}
	return "", false
}

func (b *bindataFS) AssetHttp(requested string) (http.File, error) {
	if has, ok := b.HasAsset(requested); ok {
		f, err := b.open(has)
		return f, err
	}
	return nil, errors.New(fmt.Sprintf("Asset %s unavailable", requested))
}

func (b *bindataFS) open(name string) (http.File, error) {
	name = path.Join(b.prefix, name)
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}
	if children, err := b.AssetDir(name); err == nil {
		return NewAssetDirectory(name, children, b), nil
	}
	bf, err := b.Asset(name)
	if err != nil {
		return nil, err
	}
	return NewAssetFile(name, bf), nil
}

type AssetDirectory struct {
	AssetFile
	ChildrenRead int
	Children     []os.FileInfo
}

func NewAssetDirectory(name string, children []string, fs *bindataFS) *AssetDirectory {
	fileinfos := make([]os.FileInfo, 0, len(children))
	for _, child := range children {
		_, err := fs.AssetDir(filepath.Join(name, child))
		fileinfos = append(fileinfos, &FakeFile{child, err == nil, 0})
	}
	return &AssetDirectory{
		AssetFile{
			bytes.NewReader(nil),
			ioutil.NopCloser(nil),
			FakeFile{name, true, 0},
		},
		0,
		fileinfos}
}

func (f *AssetDirectory) Readdir(count int) ([]os.FileInfo, error) {
	if count <= 0 {
		return f.Children, nil
	}
	if f.ChildrenRead+count > len(f.Children) {
		count = len(f.Children) - f.ChildrenRead
	}
	rv := f.Children[f.ChildrenRead : f.ChildrenRead+count]
	f.ChildrenRead += count
	return rv, nil
}

func (f *AssetDirectory) Stat() (os.FileInfo, error) {
	return f, nil
}

type AssetFile struct {
	*bytes.Reader
	io.Closer
	FakeFile
}

func NewAssetFile(name string, content []byte) *AssetFile {
	return &AssetFile{
		bytes.NewReader(content),
		ioutil.NopCloser(nil),
		FakeFile{name, false, int64(len(content))},
	}
}

func (f *AssetFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, errors.New("not a directory")
}

func (f *AssetFile) Size() int64 {
	return f.FakeFile.Size()
}

func (f *AssetFile) Stat() (os.FileInfo, error) {
	return f, nil
}

type FakeFile struct {
	Path string
	Dir  bool
	Len  int64
}

func (f *FakeFile) Name() string {
	_, name := filepath.Split(f.Path)
	return name
}

func (f *FakeFile) Mode() os.FileMode {
	mode := os.FileMode(0644)
	if f.Dir {
		return mode | os.ModeDir
	}
	return mode
}

func (f *FakeFile) ModTime() time.Time {
	return time.Unix(0, 0)
}

func (f *FakeFile) Size() int64 {
	return f.Len
}

func (f *FakeFile) IsDir() bool {
	return f.Mode().IsDir()
}

func (f *FakeFile) Sys() interface{} {
	return nil
}
`)
	return err
}
