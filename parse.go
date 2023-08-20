package mux

import (
	"github.com/livebud/mux/ast"
	"github.com/livebud/mux/internal/parser"
)

// Parse a route
func Parse(route string) (*ast.Route, error) {
	return parser.Parse(route)
}
