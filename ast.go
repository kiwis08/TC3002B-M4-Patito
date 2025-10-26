package main

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Program es el nodo raíz del AST
// Representa: program id ; <VARS> <P_FUNCS> main <Body> end
type Program struct {
	Keyword   string `@Keyword`
	ID        string `@Ident`
	Semicolon string `@";"`

	Vars *Vars `@@?` // Opcional: <VARS>
	// TODO: Implementar P_FUNCS
	// Por ahora los omitimos

	Main string `@Keyword`
	// Body omitido por ahora hasta implementar la estructura completa
	EndStmt string `@Keyword`
}

// Vars representa <VARS>
// <VARS> -> var <F_VAR> <P_VAR>
// Para manejar múltiples declaraciones, aceptamos cero o más FVar
type Vars struct {
	VarKeyword string `@Keyword` // "var"
	FVar       FVar   `@@`
}

// FVar representa <F_VAR>
// <F_VAR> -> id <R_ID> : <TYPE> ;
type FVar struct {
	ID    string   `@Ident`
	RID   []string `(@"," @Ident)*` // Lista de más IDs
	Colon string   `@":"`
	Type  Type     `@@`
	Semi  string   `@";"`
}

// Type representa <TYPE>
// <TYPE> -> int | float
type Type struct {
	Name string `@Type`
}

// Body representa <Body>
// Placeholder por ahora - solo capturamos tokens como stmts simples
type Body struct {
	// TODO: Definir la estructura completa de <Body>
	// Por ahora solo indicamos que existe
}

// MustBuildParser crea el parser usando participle con lexer personalizado
func MustBuildParser() *participle.Parser[Program] {
	// Definir el lexer con tokens del lenguaje
	def := lexer.MustSimple([]lexer.SimpleRule{
		// Keywords deben ir antes de Ident para prioridad
		{Name: "Keyword", Pattern: `(?i)\b(program|var|main|end)\b`},
		{Name: "Type", Pattern: `(?i)\b(int|float)\b`},
		{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
		{Name: "Punctuation", Pattern: `[;:(),]`},
		{Name: "whitespace", Pattern: `\s+`},
	})

	parser := participle.MustBuild[Program](
		participle.Lexer(def),
		participle.UseLookahead(2),
		participle.CaseInsensitive("Keyword", "Main", "EndStmt"),
	)

	return parser
}
