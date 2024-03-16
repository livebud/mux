package mux

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strings"

	"github.com/matthewmueller/enroute"
)

var (
	ErrDuplicate = enroute.ErrDuplicate
	ErrNoMatch   = enroute.ErrNoMatch
)

type Interface interface {
	Get(route string, handler http.Handler) error
	Post(route string, handler http.Handler) error
	Put(route string, handler http.Handler) error
	Patch(route string, handler http.Handler) error
	Delete(route string, handler http.Handler) error
	Set(method, route string, handler http.Handler) error
}

type Match struct {
	Method  string
	Route   string
	Path    string
	Slots   []*enroute.Slot
	Handler http.Handler
}

func New() *Router {
	return &Router{
		base:    "",
		methods: map[string]*tree{},
	}
}

type Router struct {
	base    string
	methods map[string]*tree
}

var _ http.Handler = (*Router)(nil)
var _ Interface = (*Router)(nil)

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

// Group routes within a route
func (rt *Router) Group(route string) *Router {
	return &Router{
		base:    strings.TrimSuffix(path.Join(rt.base, route), "/"),
		methods: rt.methods,
	}
}

// ServeHTTP implements http.Handler
func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := rt.Middleware(http.NotFoundHandler())
	handler.ServeHTTP(w, r)
}

// Middleware will return next on no match
func (rt *Router) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Match the path
		match, err := rt.Match(r.Method, r.URL.Path)
		if err != nil {
			if errors.Is(err, enroute.ErrNoMatch) {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Add the slots as query params
		if len(match.Slots) > 0 {
			query := r.URL.Query()
			for _, slot := range match.Slots {
				query.Set(slot.Key, slot.Value)
			}
			r.URL.RawQuery = query.Encode()
		}
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
}

func (r *Route) String() string {
	return fmt.Sprintf("%s %s", r.Method, r.Route)
}

func (rt *Router) Find(method, route string) (*Route, error) {
	tree, ok := rt.methods[method]
	if !ok {
		return nil, fmt.Errorf("router: %w found for %s %s", ErrNoMatch, method, route)
	}
	return tree.Find(method, route)
}

var methodSort = map[string]int{
	http.MethodGet:    0,
	http.MethodPost:   1,
	http.MethodPut:    2,
	http.MethodPatch:  3,
	http.MethodDelete: 4,
}

// Routes lists all the routes
func (rt *Router) Routes() (routes []*Route) {
	for method, tree := range rt.methods {
		routes = append(routes, tree.Routes(method)...)
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
	return tree.Match(method, path)
}

// Set the route
func (rt *Router) set(method, route string, handler http.Handler) error {
	return rt.insert(method, path.Join(rt.base, route), handler)
}

// Insert the route into the method's radix tree
func (rt *Router) insert(method, route string, handler http.Handler) error {
	tr := rt.methods[method]
	if tr == nil {
		tr = &tree{
			Tree:     enroute.New(),
			Handlers: map[string]http.Handler{},
		}
		rt.methods[method] = tr
	}
	return tr.Insert(route, handler)
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
