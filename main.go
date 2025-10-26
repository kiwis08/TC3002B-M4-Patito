package main

import (
	"fmt"
	"os"
)

func main() {
	// Construir el parser
	parser := MustBuildParser()

	// Programa de ejemplo con variables
	example := `
program ejemplo ;
var x, y, z : int ;
main
end
`

	fmt.Println("=== Scanner y Parser para Micro-Lenguaje Imperativo ===")
	fmt.Print("\nPrograma de entrada:")
	fmt.Print(example)
	fmt.Println("\n--- Parsing ---")

	// Parsear el programa
	program, err := parser.ParseString("", example)
	if err != nil {
		fmt.Printf("Error al parsear: %v\n", err)
		os.Exit(1)
	}

	// Mostrar resultados
	fmt.Println("✓ Parsing exitoso!")
	fmt.Printf("\nAST generado:\n")
	fmt.Printf("  Keyword (program): %s\n", program.Keyword)
	fmt.Printf("  ID del programa: %s\n", program.ID)

	if program.Vars != nil {
		fmt.Printf("  ✓ Tiene sección de variables\n")
		// TODO: Mostrar detalles de las variables parseadas
	}

	fmt.Printf("  Keyword (main): %s\n", program.Main)
	fmt.Printf("  Keyword (end): %s\n", program.EndStmt)

	fmt.Println("\n✓ Scanner y Parser implementados correctamente")
}
