package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"Patito/lexer"
	"Patito/parser"
	"Patito/semantic"
	"Patito/vm"
)

func main() {
	var data []byte
	var err error

	filename := os.Args[1]
	data, err = os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		os.Exit(1)
	}

	// Crear contexto semántico
	ctx := semantic.NewContext()

	// Generar GOTO al inicio del programa (será completado cuando se encuentre main)
	semantic.ProcessProgramStart(ctx)

	// Crear parser y asignar contexto
	p := parser.NewParser()
	p.Context = ctx

	// Parsear
	if _, err := p.Parse(lexer.NewLexer(data)); err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err)
		os.Exit(1)
	}

	compile := false
	verbose := false
	outputFile := ""

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--compile" || arg == "-c" {
			compile = true
		} else if !strings.HasPrefix(arg, "-") {
			outputFile = arg
		}
		if arg == "--verbose" || arg == "-v" {
			verbose = true
		}
	}

	if compile {
		fmt.Print("Compiling...")
		// Generar archivo .patitoc
		if outputFile == "" {
			// Generar nombre de salida basado en el archivo de entrada
			baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
			outputFile = baseName + ".patitoc"
			if verbose {
				fmt.Printf("Filename will be: %s", outputFile)
			}
		}

		if err := vm.SavePatitoc(ctx, outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "error generating .patitoc: %v\n", err)
			os.Exit(1)
		}

		if verbose {
			fmt.Printf("✓ Compiled successfully: %s -> %s\n", filename, outputFile)
		} else {
			fmt.Printf("✓ Compiled successfully\n")
		}
		fmt.Printf("  Quadruples: %d\n", ctx.Quadruples.Size())
		fmt.Printf("  Constants: %d\n", len(ctx.ConstantTable.Entries()))
		fmt.Printf("  Functions: %d\n", len(ctx.Directory.Functions))
	} else {
		// Modo por defecto: mostrar cuádruplos
		fmt.Println("OK: parsed Patito successfully")
		fmt.Println()
		fmt.Print(ctx.Quadruples.String())
	}
}
