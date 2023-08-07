package parser

import (
	"fmt"
	"regexp"

	"github.com/livebud/router/internal/ast"
	"github.com/livebud/router/internal/lexer"
	"github.com/livebud/router/internal/token"
)

func New(l *lexer.Lexer) *Parser {
	return &Parser{l: l}
}

func Parse(input string) (*ast.Route, error) {
	p := New(lexer.New(input))
	return p.Parse()
}

type Parser struct {
	l *lexer.Lexer
}

func (p *Parser) Parse() (*ast.Route, error) {
	return p.parseRoute()
}

func (p *Parser) tokenText() string {
	return p.l.Token.Text
}

func (p *Parser) tokenType() token.Type {
	return p.l.Token.Type
}

func (p *Parser) parseRoute() (*ast.Route, error) {
	route := new(ast.Route)
	for p.next() {
		section, err := p.parseSection()
		if err != nil {
			return nil, err
		}
		route.Sections = append(route.Sections, section)
	}
	return route, nil
}

func (p *Parser) parseSection() (ast.Section, error) {
	switch p.tokenType() {
	case token.Error:
		return nil, fmt.Errorf(p.tokenText())
	case token.Slash:
		return p.parseSlash()
	case token.Path:
		return p.parsePath()
	case token.OpenCurly:
		return p.parseSlot()
	default:
		return nil, fmt.Errorf("unexpected token %s", p.tokenType())
	}
}

func (p *Parser) parseSlash() (*ast.Slash, error) {
	return &ast.Slash{}, nil
}

func (p *Parser) parsePath() (*ast.Path, error) {
	return &ast.Path{
		Value: p.tokenText(),
	}, nil
}

func (p *Parser) parseSlot() (ast.Slot, error) {
	if err := p.expect(token.Slot); err != nil {
		return nil, err
	}
	key := p.tokenText()
	switch {
	case p.accept(token.Question):
		return p.parseOptionalSlot(key)
	case p.accept(token.Star):
		return p.parseWildcardSlot(key)
	case p.accept(token.Pipe):
		return p.parseRegexpSlot(key)
	default:
		return p.parseRequiredSlot(key)
	}
}

func (p *Parser) parseOptionalSlot(key string) (*ast.OptionalSlot, error) {
	node := &ast.OptionalSlot{
		Key: key,
	}
	if err := p.expect(token.CloseCurly); err != nil {
		return nil, err
	}
	return node, nil
}

func (p *Parser) parseWildcardSlot(key string) (*ast.WildcardSlot, error) {
	node := &ast.WildcardSlot{
		Key: key,
	}
	if err := p.expect(token.CloseCurly); err != nil {
		return nil, err
	}
	if err := p.expect(token.End); err != nil {
		return nil, fmt.Errorf("wildcard slots must be at the end of the path")
	}
	return node, nil
}

func (p *Parser) parseRegexpSlot(key string) (*ast.RegexpSlot, error) {
	node := &ast.RegexpSlot{
		Key: key,
	}
	if err := p.expect(token.Regexp); err != nil {
		return nil, err
	}
	regex, err := regexp.Compile(p.tokenText())
	if err != nil {
		return nil, err
	}
	node.Pattern = regex
	if err := p.expect(token.CloseCurly); err != nil {
		return nil, err
	}
	return node, nil
}

func (p *Parser) parseRequiredSlot(key string) (*ast.RequiredSlot, error) {
	node := &ast.RequiredSlot{
		Key: key,
	}
	if err := p.expect(token.CloseCurly); err != nil {
		return nil, err
	}
	return node, nil
}

func (p *Parser) next() bool {
	return p.l.Next()
}

// Returns true if all the given tokens are next
func (p *Parser) check(tokens ...token.Type) bool {
	for i, token := range tokens {
		if p.l.Peak(i+1).Type != token {
			return false
		}
	}
	return true
}

// Returns true and advances lexer if all the given tokens are next
func (p *Parser) accept(tokens ...token.Type) bool {
	if !p.check(tokens...) {
		return false
	}
	for i := 0; i < len(tokens); i++ {
		p.l.Next()
	}
	return true
}

// Returns an error if all the given tokens are not next
func (p *Parser) expect(tokens ...token.Type) error {
	for i, tok := range tokens {
		peaked := p.l.Peak(i + 1)
		if peaked.Type == token.Error {
			return fmt.Errorf(peaked.Text)
		} else if peaked.Type != tok {
			return fmt.Errorf("expected %s, got %s", tok, peaked.Type)
		}
	}
	for i := 0; i < len(tokens); i++ {
		p.l.Next()
	}
	return nil
}
