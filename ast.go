package main

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Program es el nodo raíz del AST
type Program struct {
	Keyword   string `@Keyword`
	ID        string `@Ident`
	Semicolon string `@";"`

	Vars *Vars `@@?`
	// TODO: Implementar P_FUNCS

	Main    string `@Keyword`
	Body    *Body  `@@?`
	EndStmt string `@Keyword`
}

// Vars representa <VARS>
type Vars struct {
	VarKeyword string `@Keyword`
	FVar       FVar   `@@`
}

// FVar representa <F_VAR>
type FVar struct {
	ID    string   `@Ident`
	RID   []string `(@"," @Ident)*`
	Colon string   `@":"`
	Type  Type     `@@`
	Semi  string   `@";"`
}

// Type representa <TYPE>
type Type struct {
	Name string `@Type`
}

// Body representa <BODY>
type Body struct {
	LBrace string      `@"{"`
	PStat  []Statement `@@*`
	RBrace string      `@"}"`
}

// Statement representa <STATEMENT>
type Statement struct {
	Assign    *Assign    `@@`
	Condition *Condition `@@`
	Cycle     *Cycle     `@@`
	Print     *Print     `@@`
}

// Assign representa <ASSIGN>
type Assign struct {
	ID     string     `@Ident`
	Assign string     `@"="`
	Expr   Expression `@@`
	Semi   string     `@";"`
}

// Condition representa <CONDITION>
type Condition struct {
	IfKeyword string     `@Keyword`
	LParen    string     `@"("`
	Expr      Expression `@@`
	RParen    string     `@")"`
	Body      Body       `@@`
	Else      *Else      `@@?`
	Semi      string     `@";"`
}

// Else representa <ELSE>
type Else struct {
	ElseKeyword string `@Keyword`
	Body        Body   `@@`
}

// Cycle representa <CYCLE>
type Cycle struct {
	WhileKeyword string     `@Keyword`
	LParen       string     `@"("`
	Expr         Expression `@@`
	RParen       string     `@")"`
	DoKeyword    string     `@Keyword`
	Body         Body       `@@`
	Semi         string     `@";"`
}

// Print representa <PRINT>
// <PRINT> -> print ( <PRINT_P> ) ;
// PrintP debe ser Expression o String
type Print struct {
	PrintKeyword string   `@Keyword`
	LParen       string   `@"("`
	PrintP       []string `(@String | @Ident | @Int | @Float)+` // Al menos uno
	RParen       string   `@")"`
	Semi         string   `@";"`
}

// Expression representa <EXPRESION>
// Versión simplificada que evita recursión izquierda
type Expression struct {
	// Puede ser un identificador, número, o string
	Ident  string `@Ident?`
	Int    string `@Int?`
	Float  string `@Float?`
	String string `@String?`
}

func MustBuildParser() *participle.Parser[Program] {
	def := lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Keyword", Pattern: `(?i)\b(program|var|main|end|void|while|do|if|else|print)\b`},
		{Name: "Type", Pattern: `(?i)\b(int|float)\b`},
		{Name: "String", Pattern: `"(\\"|[^"])*"`},
		{Name: "Int", Pattern: `\d+`},
		{Name: "Float", Pattern: `\d+\.\d+`},
		{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
		{Name: "Punctuation", Pattern: `[;:(),{}[\]]`},
		{Name: "Operator", Pattern: `[=+\-*/<>!]+`},
		{Name: "whitespace", Pattern: `\s+`},
	})

	parser := participle.MustBuild[Program](
		participle.Lexer(def),
		participle.UseLookahead(2),
		participle.CaseInsensitive("Keyword", "Main", "EndStmt"),
	)

	return parser
}
