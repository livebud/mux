package mux

import (
	"cmp"
	"fmt"
	"net/http"
	"path"
	"slices"
	"strings"

	"github.com/matthewmueller/enroute"
)

var (
	ErrDuplicate = enroute.ErrDuplicate
	ErrNoMatch   = enroute.ErrNoMatch
	ErrInvalid   = enroute.ErrInvalid
)

type Middleware interface {
	Middleware(next http.Handler) http.Handler
}

// Use adapts middleware functions into Middleware
type Use func(next http.Handler) http.Handler

func (fn Use) Middleware(next http.Handler) http.Handler {
	return fn(next)
}

// Mountable interface for mounting routes
type Mountable interface {
	Mount(router *Router)
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
		methods: map[string]*enroute.Tree[http.Handler]{},
	}
}

type Router struct {
	base    string
	stack   []Middleware
	methods map[string]*enroute.Tree[http.Handler]
}

var _ http.Handler = (*Router)(nil)

// Use middleware
func (rt *Router) Use(mws ...Middleware) *Router {
	return &Router{
		base:    rt.base,
		stack:   slices.Concat(rt.stack, mws),
		methods: rt.methods,
	}
}

// Group routes within a route
func (rt *Router) Group(route string) *Router {
	return &Router{
		base:    strings.TrimSuffix(path.Join(rt.base, route), "/"),
		stack:   rt.stack,
		methods: rt.methods,
	}
}

// Mount routes
func (rt *Router) Mount(mnt Mountable) *Router {
	mnt.Mount(rt)
	return rt
}

// Get route. Panics if the route is invalid or conflicts with an existing
// route. Use Set to get an error instead.
func (rt *Router) Get(route string, handler http.Handler) {
	rt.mustSet(http.MethodGet, route, handler)
}

// Post route. Panics if the route is invalid or conflicts with an existing
// route. Use Set to get an error instead.
func (rt *Router) Post(route string, handler http.Handler) {
	rt.mustSet(http.MethodPost, route, handler)
}

// Put route. Panics if the route is invalid or conflicts with an existing
// route. Use Set to get an error instead.
func (rt *Router) Put(route string, handler http.Handler) {
	rt.mustSet(http.MethodPut, route, handler)
}

// Patch route. Panics if the route is invalid or conflicts with an existing
// route. Use Set to get an error instead.
func (rt *Router) Patch(route string, handler http.Handler) {
	rt.mustSet(http.MethodPatch, route, handler)
}

// Delete route. Panics if the route is invalid or conflicts with an existing
// route. Use Set to get an error instead.
func (rt *Router) Delete(route string, handler http.Handler) {
	rt.mustSet(http.MethodDelete, route, handler)
}

// Set a handler for any HTTP method. Unlike Get, Post, etc., Set returns an
// error instead of panicking.
func (rt *Router) Set(method string, route string, handler http.Handler) error {
	if !isMethod(method) {
		return fmt.Errorf("router: %q is not a valid HTTP method", method)
	}
	return rt.set(method, route, handler)
}

// mustSet panics because a bad route is a programmer error, not a runtime
// condition. It panics with the error so callers can recover and errors.Is it.
func (rt *Router) mustSet(method, route string, handler http.Handler) {
	if err := rt.set(method, route, handler); err != nil {
		panic(err)
	}
}

func (rt *Router) set(method, route string, handler http.Handler) error {
	tr := rt.methods[method]
	if tr == nil {
		tr = enroute.New[http.Handler]()
		rt.methods[method] = tr
	}
	return tr.Insert(path.Join(rt.base, route), Compose(rt.stack...).Middleware(handler))
}

// ServeHTTP implements http.Handler. Requests with no matching route get a
// 405 with an Allow header if the path matches under another method,
// otherwise a 404. OPTIONS requests for a known path get a 204 with an Allow
// header, unless an OPTIONS route handles them.
func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.Middleware(http.HandlerFunc(rt.notFound)).ServeHTTP(w, r)
}

func (rt *Router) notFound(w http.ResponseWriter, r *http.Request) {
	var allow []string
	for method, tree := range rt.methods {
		if _, err := tree.Match(r.URL.Path); err != nil {
			continue
		}
		allow = append(allow, method)
		if method == http.MethodGet {
			allow = append(allow, http.MethodHead)
		}
	}
	if len(allow) == 0 {
		http.NotFound(w, r)
		return
	}
	allow = append(allow, http.MethodOptions)
	slices.Sort(allow)
	w.Header().Set("Allow", strings.Join(slices.Compact(allow), ", "))
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

// Middleware turns the router into middleware where if there are no matches
// it will call the next middleware in the stack. Unlike ServeHTTP, a path that
// matches under another method also calls next, since a later handler may
// serve that method.
func (rt *Router) Middleware(next http.Handler) http.Handler {
	stack := Compose(rt.stack...)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Match only fails with ErrNoMatch, so fall through to next
		match, err := rt.Match(r.Method, r.URL.Path)
		if err != nil {
			stack.Middleware(next).ServeHTTP(w, r)
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
	node, err := tree.Find(path.Join(rt.base, route))
	if err != nil {
		return nil, err
	}
	return &Route{Method: method, Route: node.Label, Handler: node.Value}, nil
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
		for _, r := range tree.Routes() {
			routes = append(routes, &Route{Method: method, Route: r.Route, Handler: r.Value})
		}
	}
	slices.SortFunc(routes, func(a, b *Route) int {
		return cmp.Or(
			cmp.Compare(methodSort[a.Method], methodSort[b.Method]),
			strings.Compare(a.Route, b.Route),
			// Methods missing from methodSort tie with GET, so break the tie
			strings.Compare(a.Method, b.Method),
		)
	})
	return routes
}

// Match a route from a method and path. HEAD requests fall back to GET routes
// when there's no HEAD route.
func (rt *Router) Match(method, path string) (*Match, error) {
	match, err := rt.match(method, path)
	if err != nil && method == http.MethodHead {
		return rt.match(http.MethodGet, path)
	}
	return match, err
}

func (rt *Router) match(method, path string) (*Match, error) {
	tree, ok := rt.methods[method]
	if !ok {
		return nil, fmt.Errorf("router: %w found for %s %s", ErrNoMatch, method, path)
	}
	m, err := tree.Match(path)
	if err != nil {
		return nil, err
	}
	return &Match{Method: method, Route: m.Route, Path: m.Path, Slots: m.Slots, Handler: m.Value}, nil
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

// Compose a stack of middleware into one middleware
func Compose(stack ...Middleware) Middleware {
	return Use(func(next http.Handler) http.Handler {
		for _, s := range slices.Backward(stack) {
			next = s.Middleware(next)
		}
		return next
	})
}
