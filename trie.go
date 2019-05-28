package main

import (
	"log"
	"net/http"
	"regexp"
	"strings"
)

// type Node struct {
// 	//name string
// 	parent   *Node
// 	children map[string]*Node
// 	handler  http.HandlerFunc
// 	wildcard bool
// }

// type Trie struct {
// 	root *Node
// }

// func NewTrie(rootNode *Node) *Trie {
// 	return &Trie{
// 		root: rootNode,
// 	}
// }

// func (t *Trie) Add(path string, newNode *Node) error {
// 	// can't add an empty string or /
// 	if path == "" || !strings.HasPrefix(path, "/") || path == "/" {
// 		return fmt.Errorf("Path cannot be empty and must start with /")
// 	}

// 	// replace any silliness
// 	path = cleanPath(path)

// 	// regexy expansions
// 	path = parsePath(path)

// 	// "/foo/bar".split == ["", "foo", "bar"]
// 	parts := strings.Split(path, "/")

// 	n := t.root

// 	for _, v := range parts {
// 		if v != "" {
// 			t := n
// 			n = n.children[v]
// 			if n == nil {
// 				newNode.parent = t
// 				t.children[v] = newNode
// 				break
// 			}
// 		}
// 	}

// 	return nil
// }

// func (t *Trie) Fetch(path string) http.HandlerFunc {

// 	if path == "/" {
// 		return t.root.handler
// 	}

// 	path = parsePath(path)
// 	parts := strings.Split(path, "/")
// 	n := t.root

// 	for _, v := range parts {
// 		if v != "" {
// 			n = n.children[v]
// 			if n == nil {
// 				return nil
// 			}
// 		}
// 	}

// 	return nil
// }

// func parsePath(path string) string {
// 	// :param ([^/]+)
// 	// *
// 	return path
// }

// func cleanPath(path string) string {
// 	return strings.Replace(path, "//", "/", -1)
// }

// WalkFunc defines some action to take on the given key and value during
// a Trie Walk. Returning a non-nil error will terminate the Walk.
type WalkFunc func(key string, value interface{}) error

// StringSegmenter takes a string key with a starting index and returns
// the first segment after the start and the ending index. When the end is
// reached, the returned nextIndex should be -1.
// Implementations should NOT allocate heap memory as Trie Segmenters are
// called upon Gets. See PathSegmenter.
type StringSegmenter func(key string, start int) (segment string, nextIndex int)

// PathSegmenter segments string key paths by slash separators. For example,
// "/a/b/c" -> ("/a", 2), ("/b", 4), ("/c", -1) in successive calls. It does
// not allocate any heap memory.
func PathSegmenter(path string, start int) (segment string, next int) {
	if len(path) == 0 || start < 0 || start > len(path)-1 {
		return "", -1
	}
	end := strings.IndexRune(path[start+1:], '/') // next '/' after 0th rune
	if end == -1 {
		return path[start:], -1
	}
	return path[start : start+end+1], start + end + 1
}

type Path struct {
	isRegex bool
	path    string
}

// PathTrie is a trie of paths with string keys and interface{} values.

// PathTrie is a trie of string keys and interface{} values. Internal nodes
// have nil values so stored nil values cannot be distinguished and are
// excluded from walks. By default, PathTrie will segment keys by forward
// slashes with PathSegmenter (e.g. "/a/b/c" -> "/a", "/b", "/c"). A custom
// StringSegmenter may be used to customize how strings are segmented into
// nodes. A classic trie might segment keys by rune (i.e. unicode points).
type PathTrie struct {
	segmenter StringSegmenter // key segmenter, must not cause heap allocs
	value     interface{}
	children  map[string]*PathTrie
}

// PathTrie node and the part string key of the child the path descends into.
type nodeStr struct {
	node *PathTrie
	part string
}

// NewPathTrie allocates and returns a new *PathTrie.
func NewPathTrie() *PathTrie {
	return &PathTrie{
		segmenter: PathSegmenter,
		children:  make(map[string]*PathTrie),
	}
}

// GetPattern asshat
func (trie *PathTrie) GetPattern(key string) interface{} {
	node := trie
	//label := regexp.MustCompile(":[a-z]+")

	for part, i := trie.segmenter(key, 0); ; part, i = trie.segmenter(key, i) {

		node = node.children[part]

		if node == nil {
			return nil
		}
		if i == -1 {
			break
		}

		for k := range node.children {
			log.Print(k, "<--")
		}

	}
	return node.value
}

// Get returns the value stored at the given key. Returns nil for internal
// nodes or for nodes with a value of nil.
func (trie *PathTrie) Get(key string) interface{} {
	node := trie
	label := regexp.MustCompile(":[a-z]+")

	for part, i := trie.segmenter(key, 0); ; part, i = trie.segmenter(key, i) {

		log.Print("part, i => ", part, ", ", i)
		if label.Match([]byte(part)) {
			log.Print("part match!")
		} else {
			log.Print(part, " does not match ", ":[a-z]+")
		}
		log.Print("---")

		node = node.children[part]
		if node == nil {
			return nil
		}
		if i == -1 {
			break
		}
	}
	return node.value
}

// Put inserts the value into the trie at the given key, replacing any
// existing items. It returns true if the put adds a new value, false
// if it replaces an existing value.
// Note that internal nodes have nil values so a stored nil value will not
// be distinguishable and will not be included in Walks.
func (trie *PathTrie) Put(key string, value interface{}) bool {

	node := trie

	for part, i := trie.segmenter(key, 0); ; part, i = trie.segmenter(key, i) {
		log.Print(part)
		child, _ := node.children[part]

		if child == nil {
			child = NewPathTrie()
			node.children[part] = child
		}
		node = child
		if i == -1 {
			break
		}
	}

	// does node have an existing value?
	isNewVal := node.value == nil
	node.value = value
	return isNewVal
}

// Delete removes the value associated with the given key. Returns true if a
// node was found for the given key. If the node or any of its ancestors
// becomes childless as a result, it is removed from the trie.
func (trie *PathTrie) Delete(key string) bool {
	var path []nodeStr // record ancestors to check later
	node := trie
	for part, i := trie.segmenter(key, 0); ; part, i = trie.segmenter(key, i) {
		path = append(path, nodeStr{part: part, node: node})
		node = node.children[part]
		if node == nil {
			// node does not exist
			return false
		}
		if i == -1 {
			break
		}
	}
	// delete the node value
	node.value = nil
	// if leaf, remove it from its parent's children map. Repeat for ancestor path.
	if node.isLeaf() {
		// iterate backwards over path
		for i := len(path) - 1; i >= 0; i-- {
			parent := path[i].node
			part := path[i].part
			delete(parent.children, part)
			if parent.value != nil || !parent.isLeaf() {
				// parent has a value or has other children, stop
				break
			}
		}
	}
	return true // node (internal or not) existed and its value was nil'd
}

// Walk iterates over each key/value stored in the trie and calls the given
// walker function with the key and value. If the walker function returns
// an error, the walk is aborted.
// The traversal is depth first with no guaranteed order.
func (trie *PathTrie) Walk(walker WalkFunc) error {
	return trie.walk("", walker)
}

func (trie *PathTrie) walk(key string, walker WalkFunc) error {
	if trie.value != nil {
		if err := walker(key, trie.value); err != nil {
			return err
		}
	}
	for part, child := range trie.children {
		if err := child.walk(key+part, walker); err != nil {
			return err
		}
	}
	return nil
}

func (trie *PathTrie) isLeaf() bool {
	return len(trie.children) == 0
}

func redirectHandler2(w http.ResponseWriter, r *http.Request, location string) []byte {
	w.Header().Set("x-foo", r.URL.String())
	w.Header().Set("Location", location)
	w.WriteHeader(301)
	return []byte("Redirecting...")
}

func redirectHandler(loc string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(301)
		w.Header().Set("x-sigma-match", r.URL.String())
		w.Header().Set("x-sigma", "redirect")
		w.Header().Set("Location", loc)
		w.Write([]byte("Redirecting..."))
	}
}

func (trie *PathTrie) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s]", r.URL)

	key := r.URL.String()

	redirect := trie.GetPattern(key)

	if redirect != nil {
		// cast it to what it actually is...
		redirect := redirect.(func(w http.ResponseWriter, r *http.Request))
		redirect(w, r)
	} else {
		w.Header().Set("X-Sigma", "revproxy")
		w.Write([]byte("Revproxy..."))
	}
}

func main() {
	t := NewPathTrie()

	t.Put("/images/:id/butts", redirectHandler("https://news.bbc.co.uk/images/:id"))
	//t.Put("/images/butts", redirectHandler("https://news.bbc.co.uk/images/:id"))

	http.ListenAndServe(":8080", t)
}
