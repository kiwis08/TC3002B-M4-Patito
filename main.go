package main

import (
	"fmt"
	"io"
	"os"

	"Patito/lexer"
	"Patito/parser"
)

func main() {
	var data []byte
	var err error
	if len(os.Args) == 2 {
		data, err = os.ReadFile(os.Args[1])
	} else {
		data, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "read error:", err)
		os.Exit(1)
	}

	p := parser.NewParser()
	if _, err := p.Parse(lexer.NewLexer(data)); err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err)
		os.Exit(1)
	}

	fmt.Println("OK: parsed Patito successfully")
}
