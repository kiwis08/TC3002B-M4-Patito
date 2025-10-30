package patito

import (
	"Patito/lexer"
	"Patito/parser"
)

// Adapter wraps the generated parser and exposes a ParseString method
// so tests can depend on a simple interface.
type Adapter struct {
	p *parser.Parser
}

// MustBuildParser returns an Adapter (not the raw *parser.Parser)
func MustBuildParser() *Adapter {
	return &Adapter{p: parser.NewParser()}
}

// ParseString matches the signature your tests expect.
func (a *Adapter) ParseString(_filename, src string) (interface{}, error) {
	return a.p.Parse(lexer.NewLexer([]byte(src)))
}

// ParseString is a compatibility helper for tests that call pwrap.ParseString.
func ParseString(a *Adapter, filename, src string) (interface{}, error) {
	return a.ParseString(filename, src)
}
