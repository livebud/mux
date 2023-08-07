package radix

import (
	"fmt"
	"strings"

	"github.com/livebud/router/internal/ast"
	"github.com/livebud/router/internal/parser"
)

var ErrDuplicate = fmt.Errorf("route already exists")
var ErrNoMatch = fmt.Errorf("no match")

func New() *Tree {
	return &Tree{}
}

type Tree struct {
	root *Node
}

func (t *Tree) Insert(route string) error {
	r, err := parser.Parse(trimTrailingSlash(route))
	if err != nil {
		return err
	}
	// Expand optional and wildcard routes
	for _, route := range r.Expand() {
		if err := t.insert(route); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tree) insert(route *ast.Route) error {
	if t.root == nil {
		t.root = &Node{
			Route:    route,
			sections: route.Sections,
		}
		return nil
	}
	// fmt.Println("inserting", route)
	return t.root.insert(route, route.Sections)
}

type Node struct {
	Route    *ast.Route
	sections ast.Sections
	children []*Node
}

// type Nodes []*Node

func (n *Node) insert(route *ast.Route, sections ast.Sections) error {
	lcp := longestCommonPrefix(n.sections, sections)
	if lcp < n.sections.Len() {
		// Split the node's sections
		parts := n.sections.Split(lcp)
		// Create a new node with the parent's sections after the lcp.
		splitChild := &Node{
			Route:    n.Route,
			sections: parts[1],
			children: n.children,
		}
		n.sections = parts[0]
		n.children = []*Node{splitChild}
		n.Route = route
		// Finally we can add the child
		if lcp < sections.Len() {
			newChild := &Node{
				Route:    route,
				sections: sections.Split(lcp)[1],
			}
			// Replace the parent's sections with the lcp.
			n.children = append(n.children, newChild)
			n.Route = nil
		}
		return nil
	}
	// Route already exists
	if lcp == sections.Len() {
		return fmt.Errorf("%w: %q", ErrDuplicate, n.sections.String())
	}
	// Check children for a match
	remainingSections := sections.Split(lcp)[1]
	for _, child := range n.children {
		if child.sections.At(0) == remainingSections.At(0) {
			return child.insert(route, remainingSections)
		}
	}
	n.children = append(n.children, &Node{
		Route:    route,
		sections: remainingSections,
	})
	return nil
}

type Slot struct {
	Key   string
	Value string
}

type Match struct {
	Route *ast.Route
	Slots []*Slot
}

func (m *Match) String() string {
	s := new(strings.Builder)
	s.WriteString(m.Route.String())
	for _, slot := range m.Slots {
		s.WriteString(" ")
		s.WriteString(slot.Key)
		s.WriteString("=")
		s.WriteString(slot.Value)
	}
	return s.String()
}

func (t *Tree) Match(path string) (*Match, error) {
	path = trimTrailingSlash(path)
	// A tree without any routes shouldn't panic
	if t.root == nil || len(path) == 0 || path[0] != '/' {
		return nil, fmt.Errorf("%w for %q", ErrNoMatch, path)
	}
	match, ok := t.root.Match(path)
	if !ok {
		return nil, fmt.Errorf("%w for %q", ErrNoMatch, path)
	}
	return match, nil
}

func (n *Node) Match(path string) (*Match, bool) {
	for _, section := range n.sections {
		index := section.Match(path)
		if index < 0 {
			return nil, false
		}
		path = path[index:]
	}
	if len(path) == 0 {
		return &Match{
			Route: n.Route,
		}, true
	}
	for _, child := range n.children {
		if match, ok := child.Match(path); ok {
			return match, true
		}
	}
	return nil, false
}

func (t *Tree) String() string {
	return t.string(t.root, "")
}

func (t *Tree) string(n *Node, indent string) string {
	route := n.sections.String()
	var mods []string
	if n.Route != nil {
		mods = append(mods, "routable="+n.Route.String())
	}
	mod := ""
	if len(mods) > 0 {
		mod = " [" + strings.Join(mods, ", ") + "]"
	}
	out := fmt.Sprintf("%s%s%s\n", indent, route, mod)
	for i := 0; i < len(route); i++ {
		indent += "•"
	}
	for _, child := range n.children {
		out += t.string(child, indent)
	}
	return out
}

func longestCommonPrefix(a, b ast.Sections) int {
	index := 0
	max := min(a.Len(), b.Len())
	for i := 0; i < max && a.At(i) == b.At(i); i++ {
		index++
	}
	return index
}

func trimTrailingSlash(input string) string {
	if input == "/" {
		return input
	}
	return strings.TrimRight(input, "/")
}
