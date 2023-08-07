package ast

import (
	"regexp"
	"strings"
)

type Node interface {
	String() string
}

var (
	_ Node = (*Route)(nil)
	_ Node = (*Slash)(nil)
	_ Node = (*Path)(nil)
	_ Node = (*RequiredSlot)(nil)
	_ Node = (*OptionalSlot)(nil)
	_ Node = (*WildcardSlot)(nil)
	_ Node = (*RegexpSlot)(nil)
)

type Routes []Route

type Route struct {
	Sections Sections
}

func (r *Route) String() string {
	s := new(strings.Builder)
	for _, section := range r.Sections {
		s.WriteString(section.String())
	}
	return s.String()
}

func trimRightSlash(r *Route) *Route {
	for i := len(r.Sections) - 1; i >= 0; i-- {
		if _, ok := r.Sections[i].(*Slash); !ok {
			r.Sections = r.Sections[:i+1]
			break
		}
	}
	return r
}

func (r *Route) Expand() (routes []*Route) {
	// Clone the route
	route := &Route{
		Sections: append(Sections{}, r.Sections...),
	}
	for i, section := range route.Sections {
		switch s := section.(type) {
		case *OptionalSlot:
			// Create route before the optional slot
			routes = append(routes, trimRightSlash(&Route{
				Sections: route.Sections[:i],
			}))
			// Create a new route with the slot required
			route.Sections[i] = &RequiredSlot{Key: s.Key}
		case *WildcardSlot:
			// Create route before the wildcard slot
			routes = append(routes, trimRightSlash(&Route{
				Sections: route.Sections[:i],
			}))
		}
	}
	routes = append(routes, route)
	return routes
}

// Section of the route
type Section interface {
	Node
	Len() int
	Match(path string) (index int)
}

type Sections []Section

func (sections Sections) At(n int) string {
	for _, section := range sections {
		switch s := section.(type) {
		case *Slash:
			if n == 0 {
				return "/"
			}
			n--
		case *Path:
			for _, char := range s.Value {
				if n == 0 {
					return string(char)
				}
				n--
			}
		case *RequiredSlot, *OptionalSlot, *WildcardSlot, *RegexpSlot:
			if n == 0 {
				return "{slot}"
			}
			n--
		}
	}
	return ""
}

func (sections Sections) Len() (n int) {
	for _, section := range sections {
		n += section.Len()
	}
	return n
}

func (sections Sections) Split(at int) []Sections {
	sections = append(Sections{}, sections...)
	for i, section := range sections {
		switch s := section.(type) {
		case *Slash:
			if at != 0 {
				at--
				continue
			}
			if i > 0 && i < len(sections) {
				return []Sections{sections[:i], sections[i:]}
			}
			return []Sections{sections}
		case *Path:
			for j := range s.Value {
				if at != 0 {
					at--
					continue
				}
				left, right := s.Value[:j], s.Value[j:]
				// At the edge
				if left == "" || right == "" {
					if i > 0 && i < len(sections) {
						return []Sections{sections[:i], sections[i:]}
					}
					return []Sections{sections}
				}
				// Split the path in two
				leftPath := &Path{Value: left}
				rightPath := &Path{Value: right}
				leftSections := append(append(Sections{}, sections[:i]...), leftPath)
				rightSections := append(append(Sections{}, rightPath), sections[i+1:]...)
				return []Sections{leftSections, rightSections}
			}
		case *RequiredSlot, *OptionalSlot, *WildcardSlot, *RegexpSlot:
			if at != 0 {
				at--
				continue
			}
			if i > 0 && i < len(sections) {
				return []Sections{sections[:i], sections[i:]}
			}
			return []Sections{sections}
		}
	}
	return []Sections{sections}
}

func (sections Sections) String() string {
	s := new(strings.Builder)
	for _, section := range sections {
		s.WriteString(section.String())
	}
	return s.String()
}

var (
	_ Section = (*Slash)(nil)
	_ Section = (*Path)(nil)
	_ Section = (*OptionalSlot)(nil)
	_ Section = (*WildcardSlot)(nil)
	_ Section = (*RegexpSlot)(nil)
)

type Slash struct {
	Value string
}

func (s *Slash) String() string {
	return "/"
}

func (p *Slash) Len() int {
	return 1
}

func (p *Slash) Match(path string) (index int) {
	if len(path) == 0 {
		return -1
	}
	if path[0] == '/' {
		return 1
	}
	return -1
}

type Path struct {
	Value string
}

func (p *Path) String() string {
	return p.Value
}

func (p *Path) Len() int {
	return len(p.Value)
}

func (p *Path) Match(path string) (index int) {
	valueLen := p.Len()
	if len(path) < valueLen {
		return -1
	}
	prefix := strings.ToLower(path[:valueLen])
	if prefix != p.Value {
		return -1
	}
	return valueLen
}

type Slot interface {
	Node
	Section
	slot()
}

var (
	_ Slot = (*RequiredSlot)(nil)
	_ Slot = (*OptionalSlot)(nil)
	_ Slot = (*WildcardSlot)(nil)
	_ Slot = (*RegexpSlot)(nil)
)

type RequiredSlot struct {
	Key string
}

func (s *RequiredSlot) Len() int {
	return 1
}

func (s *RequiredSlot) slot() {}

func (s *RequiredSlot) String() string {
	return "{" + s.Key + "}"
}

func (s *RequiredSlot) Match(path string) (index int) {
	return -1
	// valueLen := s.Len()
	// if len(path) < valueLen {
	// 	return -1
	// }
	// prefix := strings.ToLower(path[:valueLen])
	// if prefix != p.Value {
	// 	return -1
	// }
	// return valueLen
}

type OptionalSlot struct {
	Key string
}

func (s *OptionalSlot) Len() int {
	return 1
}

func (s *OptionalSlot) slot() {}

func (o *OptionalSlot) String() string {
	return "{" + o.Key + "?}"
}

func (s *OptionalSlot) Match(path string) (index int) {
	return -1
}

type WildcardSlot struct {
	Key string
}

func (s *WildcardSlot) Len() int {
	return 1
}

func (s *WildcardSlot) slot() {}

func (w *WildcardSlot) String() string {
	return "{" + w.Key + "*}"
}

func (s *WildcardSlot) Match(path string) (index int) {
	return -1
}

type RegexpSlot struct {
	Key     string
	Pattern *regexp.Regexp
}

func (s *RegexpSlot) Len() int {
	return 1
}

func (s *RegexpSlot) slot() {}

func (r *RegexpSlot) String() string {
	return "{" + r.Key + "|" + r.Pattern.String() + "}"
}

func (s *RegexpSlot) Match(path string) (index int) {
	return -1
}
