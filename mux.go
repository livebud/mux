package mux

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strings"

	"github.com/livebud/mux/ast"
	"github.com/livebud/mux/internal/radix"
	"github.com/livebud/slot"
)

var (
	ErrDuplicate = radix.ErrDuplicate
	ErrNoMatch   = radix.ErrNoMatch
)

type Router interface {
	http.Handler
	Get(route string, handler http.Handler) error
	Post(route string, handler http.Handler) error
	Put(route string, handler http.Handler) error
	Patch(route string, handler http.Handler) error
	Delete(route string, handler http.Handler) error
	Layout(route string, handler http.Handler) error
	Error(route string, handler http.Handler) error
	Group(route string) Router
	Routes() []*Route
}

type Match struct {
	Method  string
	Route   *ast.Route
	Path    string
	Slots   []*radix.Slot
	Handler http.Handler
	Layout  *LayoutHandler
	Error   *ErrorHandler
}

func New() *router {
	return &router{
		base:    "",
		methods: map[string]*radix.Tree{},
		layouts: radix.New(),
		errors:  radix.New(),
	}
}

type router struct {
	base    string
	methods map[string]*radix.Tree
	layouts *radix.Tree
	errors  *radix.Tree
}

var _ http.Handler = (*router)(nil)
var _ Router = (*router)(nil)

// Get route
func (rt *router) Get(route string, handler http.Handler) error {
	return rt.set(http.MethodGet, route, handler)
}

// Post route
func (rt *router) Post(route string, handler http.Handler) error {
	return rt.set(http.MethodPost, route, handler)
}

// Put route
func (rt *router) Put(route string, handler http.Handler) error {
	return rt.set(http.MethodPut, route, handler)
}

// Patch route
func (rt *router) Patch(route string, handler http.Handler) error {
	return rt.set(http.MethodPatch, route, handler)
}

// Delete route
func (rt *router) Delete(route string, handler http.Handler) error {
	return rt.set(http.MethodDelete, route, handler)
}

// Layout attaches a layout handler to a given route
func (rt *router) Layout(route string, handler http.Handler) error {
	return rt.layouts.Insert(route, handler)
}

// Error attaches an error handler to a given route
func (rt *router) Error(route string, handler http.Handler) error {
	return rt.errors.Insert(route, handler)
}

// Routable is an interface for adding routes
type Routable interface {
	Routes(rt Router)
}

// Add routes from routables
func (rt *router) Add(routable Routable) {
	routable.Routes(rt)
}

// Group routes within a route
func (rt *router) Group(route string) Router {
	return &router{
		base:    strings.TrimSuffix(path.Join(rt.base, route), "/"),
		methods: rt.methods,
		layouts: rt.layouts,
		errors:  rt.errors,
	}
}

// ServeHTTP implements http.Handler
func (rt *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := rt.Middleware(http.NotFoundHandler())
	handler.ServeHTTP(w, r)
}

// Middleware will return next on no match
func (rt *router) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Match the path
		match, err := rt.Match(r.Method, r.URL.Path)
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
		// Chain the handlers together
		var handlers []http.Handler
		handlers = append(handlers, match.Handler)
		if match.Layout != nil {
			handlers = append(handlers, match.Layout.Handler)
		}
		// Batch the handlers together
		handler := slot.Batch(handlers...)
		// Call the handler
		handler.ServeHTTP(w, r)
	})
}

// Set a handler manually
func (rt *router) Set(method string, route string, handler http.Handler) error {
	if !isMethod(method) {
		return fmt.Errorf("router: %q is not a valid HTTP method", method)
	}
	return rt.set(method, route, handler)
}

type Route struct {
	Method  string
	Route   string
	Handler http.Handler
	Layout  *LayoutHandler
	Error   *ErrorHandler
}

type LayoutHandler struct {
	Route   string
	Handler http.Handler
}

type ErrorHandler struct {
	Route   string
	Handler http.Handler
}

func (r *Route) String() string {
	return fmt.Sprintf("%s %s", r.Method, r.Route)
}

func (rt *router) Find(method, route string) (*Route, error) {
	tree, ok := rt.methods[method]
	if !ok {
		return nil, fmt.Errorf("router: %w found for %s %s", ErrNoMatch, method, route)
	}
	node, err := tree.Find(route)
	if err != nil {
		return nil, err
	}
	r := &Route{
		Method:  method,
		Route:   node.Route.String(),
		Handler: node.Handler,
	}
	// Find the layout handler
	r.Layout, err = rt.findLayout(method, node.Route.String())
	if err != nil {
		return nil, err
	}
	// Find the error handler
	r.Error, err = rt.findError(method, node.Route.String())
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Routes lists all the routes
func (rt *router) Routes() (routes []*Route) {
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
func (rt *router) Match(method, path string) (*Match, error) {
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
	// Find the layout handler
	match.Layout, err = rt.findLayout(method, m.Route.String())
	if err != nil {
		return nil, err
	}
	// Find the error handler
	match.Error, err = rt.findError(method, m.Route.String())
	if err != nil {
		return nil, err
	}
	return match, nil
}

// FindLayout finds the layout for a given path
func (rt *router) findLayout(method, route string) (*LayoutHandler, error) {
	if method != http.MethodGet {
		return nil, nil
	}
	layout, err := rt.layouts.FindByPrefix(route)
	if err != nil {
		if !errors.Is(err, ErrNoMatch) {
			return nil, err
		}
		return nil, nil
	}
	return &LayoutHandler{
		Route:   layout.Route.String(),
		Handler: layout.Handler,
	}, nil
}

// Find the error handler
func (rt *router) findError(method, route string) (*ErrorHandler, error) {
	if method != http.MethodGet {
		return nil, nil
	}
	errorHandler, err := rt.errors.FindByPrefix(route)
	if err != nil {
		if !errors.Is(err, ErrNoMatch) {
			return nil, err
		}
		return nil, nil
	}
	return &ErrorHandler{
		Route:   errorHandler.Route.String(),
		Handler: errorHandler.Handler,
	}, nil
}

var methodSort = map[string]int{
	http.MethodGet:    0,
	http.MethodPost:   1,
	http.MethodPut:    2,
	http.MethodPatch:  3,
	http.MethodDelete: 4,
}

// Set the route
func (rt *router) set(method, route string, handler http.Handler) error {
	return rt.insert(method, path.Join(rt.base, route), handler)
}

// Insert the route into the method's radix tree
func (rt *router) insert(method, route string, handler http.Handler) error {
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
