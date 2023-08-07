package token

import (
	"strconv"
	"strings"
)

type Type string

type Token struct {
	Type  Type
	Text  string
	Start int
	Line  int
}

func (t *Token) String() string {
	s := new(strings.Builder)
	s.WriteString(string(t.Type))
	if t.Text != "" && t.Text != string(t.Type) {
		s.WriteString(":")
		s.WriteString(strconv.Quote(t.Text))
	}
	return s.String()
}

const (
	End        Type = "end"
	Error      Type = "error"
	Regexp     Type = "regexp"
	Path       Type = "path"
	Slot       Type = "slot"
	Slash      Type = "/"
	OpenCurly  Type = "{"
	CloseCurly Type = "}"
	Question   Type = "?"
	Star       Type = "*"
	Pipe       Type = "|"
)

// func Expand(tokens []Token) (tokensets [][]Token) {
// 	for i, token := range tokens {
// 		switch token.Type {
// 		case Question:
// 			tokensets = append(tokensets, stripTokenTrail(tokens[:i]))
// 			tokens[i] = Token{Type: Slot, Text: token.Text, Start: token.Start, Line: token.Line}
// 		case Star:
// 			tokensets = append(tokensets, stripTokenTrail(tokens[:i]))
// 			tokens[i] = Token{Type: Slot, Text: token.Text, Start: token.Start, Line: token.Line}
// 		}
// 	}
// 	tokensets = append(tokensets, tokens)
// 	return tokensets
// }

// // strip token trail removes path tokens up to either a slot or a slash
// // e.g. /:id. => /:id
// //
// //	/a/b => /a
// func stripTokenTrail(tokens []Token) []Token {
// 	for i := len(tokens) - 1; i >= 0; i-- {
// 		if tokens[i].Type != Slash {
// 			return tokens[0:i]
// 		}
// 	}
// 	return tokens
// }
