package mux

import (
	"fmt"
	"net/http"

	"github.com/matthewmueller/enroute"
)

type tree struct {
	Tree     *enroute.Tree
	Handlers map[string]http.Handler
}

func (t *tree) Insert(route string, handler http.Handler) error {
	if err := t.Tree.Insert(route, route); err != nil {
		return err
	}
	t.Handlers[route] = handler
	return nil
}

func (t *tree) Find(method, route string) (*Route, error) {
	node, err := t.Tree.Find(route)
	if err != nil {
		return nil, err
	}
	handler, ok := t.Handlers[route]
	if !ok {
		return nil, fmt.Errorf("router: handler not found for %s %s", method, route)
	}
	return &Route{
		Method:  method,
		Route:   node.Label,
		Handler: handler,
	}, nil
}

func (t *tree) Match(method, path string) (*Match, error) {
	m, err := t.Tree.Match(path)
	if err != nil {
		return nil, err
	}
	handler, ok := t.Handlers[m.Value]
	if !ok {
		return nil, fmt.Errorf("router: no handler provided for %s %s", method, path)
	}
	return &Match{
		Method:  method,
		Route:   m.Route,
		Path:    m.Path,
		Slots:   m.Slots,
		Handler: handler,
	}, nil
}

func (t *tree) Routes(method string) (routes []*Route) {
	t.Tree.Each(func(node *enroute.Node) bool {
		if node.Label == "" {
			return true
		}
		handler, ok := t.Handlers[node.Value]
		if !ok {
			return true
		}
		routes = append(routes, &Route{
			Method:  method,
			Route:   node.Label,
			Handler: handler,
		})
		return true
	})
	return routes
}
