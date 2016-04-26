package engine

import (
	"strings"
	"unicode"
)

type Param struct {
	Key   string
	Value string
}

type Params []Param

func (ps Params) ByName(name string) string {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value
		}
	}
	return ""
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func countParams(path string) uint8 {
	var n uint
	for i := 0; i < len(path); i++ {
		if path[i] != ':' && path[i] != '*' {
			continue
		}
		n++
	}
	if n >= 255 {
		return 255
	}
	return uint8(n)
}

type nodeType uint8

const (
	static   nodeType = 0
	param    nodeType = 1
	catchAll nodeType = 2
)

type node struct {
	path      string
	wildChild bool
	nType     nodeType
	maxParams uint8
	indices   []byte
	children  []*node
	rule      Rule
	priority  uint32
}

func (n *node) incrementChildPrio(i int) int {
	n.children[i].priority++
	prio := n.children[i].priority

	// adjust position (move to front)
	for j := i - 1; j >= 0 && n.children[j].priority < prio; j-- {
		// swap node positions
		tmpN := n.children[j]
		n.children[j] = n.children[i]
		n.children[i] = tmpN
		tmpI := n.indices[j]
		n.indices[j] = n.indices[i]
		n.indices[i] = tmpI

		i--
	}
	return i
}

func (n *node) addRoute(path string, rule Rule) {
	n.priority++
	numParams := countParams(path)

	// non-empty tree
	if len(n.path) > 0 || len(n.children) > 0 {
	WALK:
		for {
			// Update maxParams of the current node
			if numParams > n.maxParams {
				n.maxParams = numParams
			}

			// Find the longest common prefix.
			// This also implies that the commom prefix contains no ':' or '*'
			// since the existing key can't contain this chars.
			i := 0
			for max := min(len(path), len(n.path)); i < max && path[i] == n.path[i]; i++ {
			}

			// Split edge
			if i < len(n.path) {
				child := node{
					path:      n.path[i:],
					wildChild: n.wildChild,
					indices:   n.indices,
					children:  n.children,
					rule:      n.rule,
					priority:  n.priority - 1,
				}

				// Update maxParams (max of all children)
				for i := range child.children {
					if child.children[i].maxParams > child.maxParams {
						child.maxParams = child.children[i].maxParams
					}
				}

				n.children = []*node{&child}
				n.indices = []byte{n.path[i]}
				n.path = path[:i]
				n.rule = nil
				n.wildChild = false
			}

			// Make new node a child of this node
			if i < len(path) {
				path = path[i:]

				if n.wildChild {
					n = n.children[0]
					n.priority++

					// Update maxParams of the child node
					if numParams > n.maxParams {
						n.maxParams = numParams
					}
					numParams--

					// Check if the wildcard matches
					if len(path) >= len(n.path) && n.path == path[:len(n.path)] {
						// check for longer wildcard, e.g. :name and :names
						if len(n.path) >= len(path) || path[len(n.path)] == '/' {
							continue WALK
						}
					}

					panic("conflict with wildcard route")
				}

				c := path[0]

				// slash after param
				if n.nType == param && c == '/' && len(n.children) == 1 {
					n = n.children[0]
					n.priority++
					continue WALK
				}

				// Check if a child with the next path byte exists
				for i, index := range n.indices {
					if c == index {
						i = n.incrementChildPrio(i)
						n = n.children[i]
						continue WALK
					}
				}

				// Otherwise insert it
				if c != ':' && c != '*' {
					n.indices = append(n.indices, c)
					child := &node{
						maxParams: numParams,
					}
					n.children = append(n.children, child)
					n.incrementChildPrio(len(n.indices) - 1)
					n = child
				}
				n.insertChild(numParams, path, rule)
				return

			} else if i == len(path) { // Make node a (in-path) leaf
				if n.rule != nil {
					panic("a Rule is already registered for this path")
				}
				n.rule = rule
			}
			return
		}
	} else { // Empty tree
		n.insertChild(numParams, path, rule)
	}
}

func (n *node) insertChild(numParams uint8, path string, rule Rule) {
	var offset int

	// find prefix until first wildcard (beginning with ':'' or '*'')
	for i, max := 0, len(path); numParams > 0; i++ {
		c := path[i]
		if c != ':' && c != '*' {
			continue
		}

		// Check if this Node existing children which would be
		// unreachable if we insert the wildcard here
		if len(n.children) > 0 {
			panic("wildcard route conflicts with existing children")
		}

		// find wildcard end (either '/' or path end)
		end := i + 1
		for end < max && path[end] != '/' {
			end++
		}

		if end-i < 2 {
			panic("wildcards must be named with a non-empty name")
		}

		if c == ':' { // param
			// split path at the beginning of the wildcard
			if i > 0 {
				n.path = path[offset:i]
				offset = i
			}

			child := &node{
				nType:     param,
				maxParams: numParams,
			}
			n.children = []*node{child}
			n.wildChild = true
			n = child
			n.priority++
			numParams--

			// if the path doesn't end with the wildcard, then there
			// will be another non-wildcard subpath starting with '/'
			if end < max {
				n.path = path[offset:end]
				offset = end

				child := &node{
					maxParams: numParams,
					priority:  1,
				}
				n.children = []*node{child}
				n = child
			}

		} else { // catchAll
			if end != max || numParams > 1 {
				panic("catch-all routes are only allowed at the end of the path")
			}

			if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
				panic("catch-all conflicts with existing handle for the path segment root")
			}

			// currently fixed width 1 for '/'
			i--
			if path[i] != '/' {
				panic("no / before catch-all")
			}

			n.path = path[offset:i]

			// first node: catchAll node with empty path
			child := &node{
				wildChild: true,
				nType:     catchAll,
				maxParams: 1,
			}
			n.children = []*node{child}
			n.indices = []byte{path[i]}
			n = child
			n.priority++

			// second node: node holding the variable
			child = &node{
				path:      path[i:],
				nType:     catchAll,
				maxParams: 1,
				rule:      rule,
				priority:  1,
			}
			n.children = []*node{child}

			return
		}
	}

	// insert remaining path part and handle to the leaf
	n.path = path[offset:]
	n.rule = rule
}

func (n *node) getValue(path string) (rule Rule, p Params, tsr bool) {
walk: // Outer loop for walking the tree
	for {
		if len(path) > len(n.path) {
			if path[:len(n.path)] == n.path {
				path = path[len(n.path):]
				// If this node does not have a wildcard (param or catchAll)
				// child,  we can just look up the next child node and continue
				// to walk down the tree
				if !n.wildChild {
					c := path[0]
					for i, index := range n.indices {
						if c == index {
							n = n.children[i]
							continue walk
						}
					}

					// Nothing found.
					// We can recommend to redirect to the same URL without a
					// trailing slash if a leaf exists for that path.
					tsr = (path == "/" && n.rule != nil)
					return

				}

				// handle wildcard child
				n = n.children[0]
				switch n.nType {
				case param:
					// find param end (either '/' or path end)
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}

					// save param value
					if p == nil {
						// lazy allocation
						p = make(Params, 0, n.maxParams)
					}
					i := len(p)
					p = p[:i+1] // expand slice within preallocated capacity
					p[i].Key = n.path[1:]
					p[i].Value = path[:end]

					// we need to go deeper!
					if end < len(path) {
						if len(n.children) > 0 {
							path = path[end:]
							n = n.children[0]
							continue walk
						}

						// ... but we can't
						tsr = (len(path) == end+1)
						return
					}

					if rule = n.rule; rule != nil {
						return
					} else if len(n.children) == 1 {
						// No handle found. Check if a handle for this path + a
						// trailing slash exists for TSR recommendation
						n = n.children[0]
						tsr = (n.path == "/" && n.rule != nil)
					}

					return

				case catchAll:
					// save param value
					if p == nil {
						// lazy allocation
						p = make(Params, 0, n.maxParams)
					}
					i := len(p)
					p = p[:i+1] // expand slice within preallocated capacity
					p[i].Key = n.path[2:]
					p[i].Value = path

					rule = n.rule
					return

				default:
					panic("Invalid node type")
				}
			}
		} else if path == n.path {
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if rule = n.rule; rule != nil {
				return
			}

			// No handle found. Check if a handle for this path + a
			// trailing slash exists for trailing slash recommendation
			for i, index := range n.indices {
				if index == '/' {
					n = n.children[i]
					tsr = (n.path == "/" && n.rule != nil) ||
						(n.nType == catchAll && n.children[0].rule != nil)
					return
				}
			}

			return
		}

		// Nothing found. We can recommend to redirect to the same URL with an
		// extra trailing slash if a leaf exists for that path
		tsr = (path == "/") ||
			(len(n.path) == len(path)+1 && n.path[len(path)] == '/' &&
				path == n.path[:len(n.path)-1] && n.rule != nil)
		return
	}
}

func (n *node) findCaseInsensitivePath(path string, fixTrailingSlash bool) (ciPath []byte, found bool) {
	ciPath = make([]byte, 0, len(path)+1) // preallocate enough memory

	// Outer loop for walking the tree
	for len(path) >= len(n.path) && strings.ToLower(path[:len(n.path)]) == strings.ToLower(n.path) {
		path = path[len(n.path):]
		ciPath = append(ciPath, n.path...)

		if len(path) > 0 {
			// If this node does not have a wildcard (param or catchAll) child,
			// we can just look up the next child node and continue to walk down
			// the tree
			if !n.wildChild {
				r := unicode.ToLower(rune(path[0]))
				for i, index := range n.indices {
					// must use recursive approach since both index and
					// ToLower(index) could exist. We must check both.
					if r == unicode.ToLower(rune(index)) {
						out, found := n.children[i].findCaseInsensitivePath(path, fixTrailingSlash)
						if found {
							return append(ciPath, out...), true
						}
					}
				}

				// Nothing found. We can recommend to redirect to the same URL
				// without a trailing slash if a leaf exists for that path
				found = (fixTrailingSlash && path == "/" && n.rule != nil)
				return

			} else {
				n = n.children[0]

				switch n.nType {
				case param:
					// find param end (either '/' or path end)
					k := 0
					for k < len(path) && path[k] != '/' {
						k++
					}

					// add param value to case insensitive path
					ciPath = append(ciPath, path[:k]...)

					// we need to go deeper!
					if k < len(path) {
						if len(n.children) > 0 {
							path = path[k:]
							n = n.children[0]
							continue
						} else { // ... but we can't
							if fixTrailingSlash && len(path) == k+1 {
								return ciPath, true
							}
							return
						}
					}

					if n.rule != nil {
						return ciPath, true
					} else if fixTrailingSlash && len(n.children) == 1 {
						// No handle found. Check if a handle for this path + a
						// trailing slash exists
						n = n.children[0]
						if n.path == "/" && n.rule != nil {
							return append(ciPath, '/'), true
						}
					}
					return

				case catchAll:
					return append(ciPath, path...), true

				default:
					panic("Invalid node type")
				}
			}
		} else {
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if n.rule != nil {
				return ciPath, true
			}

			// No handle found.
			// Try to fix the path by adding a trailing slash
			if fixTrailingSlash {
				for i, index := range n.indices {
					if index == '/' {
						n = n.children[i]
						if (n.path == "/" && n.rule != nil) ||
							(n.nType == catchAll && n.children[0].rule != nil) {
							return append(ciPath, '/'), true
						}
						return
					}
				}
			}
			return
		}
	}

	// Nothing found.
	// Try to fix the path by adding / removing a trailing slash
	if fixTrailingSlash {
		if path == "/" {
			return ciPath, true
		}
		if len(path)+1 == len(n.path) && n.path[len(path)] == '/' &&
			strings.ToLower(path) == strings.ToLower(n.path[:len(path)]) &&
			n.rule != nil {
			return append(ciPath, n.path...), true
		}
	}
	return
}
