package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Node struct {
	//name string
	parent   *Node
	children map[string]*Node
	handler  http.HandlerFunc
	wildcard bool
}

type Trie struct {
	root *Node
}

func NewTrie(rootNode *Node) *Trie {
	return &Trie{
		root: rootNode,
	}
}

func (t *Trie) Add(path string, newNode *Node) error {
	// can't add an empty string or /
	if path == "" || !strings.HasPrefix(path, "/") || path == "/" {
		return fmt.Errorf("Path cannot be empty and must start with /")
	}

	// replace any silliness
	path = cleanPath(path)

	// regexy expansions
	path = parsePath(path)

	// "/foo/bar".split == ["", "foo", "bar"]
	parts := strings.Split(path, "/")

	n := t.root

	for _, v := range parts {
		if v != "" {
			t := n
			n = n.children[v]
			if n == nil {
				newNode.parent = t
				t.children[v] = newNode
				break
			}
		}
	}

	return nil
}

func (t *Trie) Fetch(path string) http.HandlerFunc {

	if path == "/" {
		return t.root.handler
	}

	path = parsePath(path)
	parts := strings.Split(path, "/")
	n := t.root

	for _, v := range parts {
		if v != "" {
			n = n.children[v]
			if n == nil {
				return nil
			}
		}
	}

	return nil
}

func parsePath(path string) string {
	// :param ([^/]+)
	// *
	return path
}

func cleanPath(path string) string {
	return strings.Replace(path, "//", "/", -1)
}

func main() {

	redirectHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "I love /")
	}

	rootNode := &Node{
		wildcard: true,
		handler: redirectHandler,
		children: make(map[string]*Node),
	}

	t := NewTrie(rootNode)

	fooBar := &Node{
		handler: func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "I love /foo/bar")
		},
	}

	t.Add("/foo/:bar", fooBar)

	node := t.Fetch("/")
	log.Print(node)

	node = t.Fetch("/bar")
	log.Print(node)

	node = t.Fetch("/foo/bar")
	log.Print(node)

}
