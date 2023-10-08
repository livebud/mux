package mux

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/livebud/mux/ast"
	"github.com/livebud/mux/internal/radix"
)

var (
	ErrDuplicate = radix.ErrDuplicate
	ErrNoMatch   = radix.ErrNoMatch
)

type Match struct {
	Method  string
	Route   *ast.Route
	Path    string
	Slots   []*radix.Slot
	Handler http.Handler
	Layout  *Partial
	Error   *Partial
}

func New() *Router {
	return &Router{
		methods: map[string]*radix.Tree{},
		layouts: radix.New(),
		errors:  radix.New(),
	}
}

type Router struct {
	methods map[string]*radix.Tree
	layouts *radix.Tree
	errors  *radix.Tree
}

var _ http.Handler = (*Router)(nil)

// Get route
func (rt *Router) Get(route string, handler http.Handler) error {
	return rt.set(http.MethodGet, route, handler)
}

// Post route
func (rt *Router) Post(route string, handler http.Handler) error {
	return rt.set(http.MethodPost, route, handler)
}

// Put route
func (rt *Router) Put(route string, handler http.Handler) error {
	return rt.set(http.MethodPut, route, handler)
}

// Patch route
func (rt *Router) Patch(route string, handler http.Handler) error {
	return rt.set(http.MethodPatch, route, handler)
}

// Delete route
func (rt *Router) Delete(route string, handler http.Handler) error {
	return rt.set(http.MethodDelete, route, handler)
}

// Layout attaches a layout handler to a given route
func (rt *Router) Layout(route string, handler http.Handler) error {
	return rt.layouts.Insert(route, handler)
}

// Error attaches an error handler to a given route
func (rt *Router) Error(route string, handler http.Handler) error {
	return rt.errors.Insert(route, handler)
}

// ServeHTTP implements http.Handler
func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := rt.Middleware(http.NotFoundHandler())
	handler.ServeHTTP(w, r)
}

// Middleware will return next on no match
func (rt *Router) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tree, ok := rt.methods[r.Method]
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		// Match the path
		match, err := tree.Match(r.URL.Path)
		if err != nil {
			if errors.Is(err, radix.ErrNoMatch) {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Add the slots
		if len(match.Slots) > 0 {
			query := r.URL.Query()
			for _, slot := range match.Slots {
				query.Set(slot.Key, slot.Value)
			}
			r.URL.RawQuery = query.Encode()
		}
		// Call the handler
		match.Handler.ServeHTTP(w, r)
	})
}

// Set a handler manually
func (rt *Router) Set(method string, route string, handler http.Handler) error {
	if !isMethod(method) {
		return fmt.Errorf("router: %q is not a valid HTTP method", method)
	}
	return rt.set(method, route, handler)
}

type Route struct {
	Method  string
	Route   string
	Handler http.Handler
	Layout  *Partial
	Error   *Partial
}

type Partial struct {
	Route   string
	Handler http.Handler
}

func (r *Route) String() string {
	return fmt.Sprintf("%s %s", r.Method, r.Route)
}

func (rt *Router) Find(method, route string) (*Route, error) {
	tree, ok := rt.methods[method]
	if !ok {
		return nil, fmt.Errorf("router: %w found for %s %s", ErrNoMatch, method, route)
	}
	node, err := tree.Find(route)
	if err != nil {
		return nil, err
	}
	return &Route{
		Method:  method,
		Route:   node.Route.String(),
		Handler: node.Handler,
	}, nil
}

// Routes lists all the routes
func (rt *Router) Routes() (routes []*Route) {
	for method, tree := range rt.methods {
		tree.Each(func(node *radix.Node) bool {
			if node.Route == nil {
				return true
			}
			routes = append(routes, &Route{
				Method:  method,
				Route:   node.Route.String(),
				Handler: node.Handler,
			})
			return true
		})
	}
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Method != routes[j].Method {
			return methodSort[routes[i].Method] < methodSort[routes[j].Method]
		}
		return routes[i].Route < routes[j].Route
	})
	return routes
}

// Match a route from a method and path
func (rt *Router) Match(method, path string) (*Match, error) {
	tree, ok := rt.methods[method]
	if !ok {
		return nil, fmt.Errorf("router: %w found for %s %s", ErrNoMatch, method, path)
	}
	m, err := tree.Match(path)
	if err != nil {
		return nil, err
	} else if m.Handler == nil {
		// Internal route without a handler
		return nil, fmt.Errorf("router: %w found for %s %s", ErrNoMatch, method, path)
	}
	match := &Match{
		Method:  method,
		Route:   m.Route,
		Path:    m.Path,
		Slots:   m.Slots,
		Handler: m.Handler,
	}
	// Find the layout
	if method == http.MethodGet {
		if layout, err := rt.layouts.FindByPrefix(match.Route.String()); err != nil && !errors.Is(err, radix.ErrNoMatch) {
			return nil, err
		} else if layout != nil {
			match.Layout = &Partial{
				Route:   layout.Route.String(),
				Handler: layout.Handler,
			}
		}
		// Find the error
		if error, err := rt.errors.FindByPrefix(match.Route.String()); err != nil && !errors.Is(err, radix.ErrNoMatch) {
			return nil, err
		} else if error != nil {
			match.Error = &Partial{
				Route:   error.Route.String(),
				Handler: error.Handler,
			}
		}
	}
	return match, nil
}

var methodSort = map[string]int{
	http.MethodGet:    0,
	http.MethodPost:   1,
	http.MethodPut:    2,
	http.MethodPatch:  3,
	http.MethodDelete: 4,
}

// Set the route
func (rt *Router) set(method, route string, handler http.Handler) error {
	return rt.insert(method, route, handler)
}

// Insert the route into the method's radix tree
func (rt *Router) insert(method, route string, handler http.Handler) error {
	if _, ok := rt.methods[method]; !ok {
		rt.methods[method] = radix.New()
	}
	return rt.methods[method].Insert(route, handler)
}

// isMethod returns true if method is a valid HTTP method
func isMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost,
		http.MethodPut, http.MethodPatch, http.MethodDelete,
		http.MethodConnect, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}
