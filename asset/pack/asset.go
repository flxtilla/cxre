package pack

type Asset struct {
	Path string // Full file path.
	Name string // Key used in TOC -- name by which asset is referenced.
	Func string // Function name for the procedure returning the asset contents.
}
