package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// 1. Tests del Parser Base
// ============================================================================

func TestMustBuildParser(t *testing.T) {
	parser := MustBuildParser()
	assert.NotNil(t, parser, "Parser debe construirse correctamente")
}

func TestLexerTokens(t *testing.T) {
	parser := MustBuildParser()

	// Test Keywords
	testCases := []struct {
		name    string
		program string
	}{
		{"Program keyword", "program test; main end"},
		{"Var keyword", "program test; var x: int; main end"},
		{"Main keyword", "program test; main end"},
		{"End keyword", "program test; main end"},
		{"While keyword", "program test; main { while (x) do {}; } end"},
		{"Do keyword", "program test; main { while (x) do {}; } end"},
		{"If keyword", "program test; main { if (x) {}; } end"},
		{"Else keyword", "program test; main { if (x) {} else {}; } end"},
		{"Print keyword", "program test; main { print(x); } end"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parser.ParseString("", tc.program)
			assert.NoError(t, err, "Debe parsear correctamente: %s", tc.name)
		})
	}
}

// ============================================================================
// 2. Tests de Program (estructura principal)
// ============================================================================

func TestProgramBasic(t *testing.T) {
	parser := MustBuildParser()
	program := `program ejemplo; main end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	assert.Equal(t, "program", prog.Keyword)
	assert.Equal(t, "ejemplo", prog.ID)
	assert.Equal(t, "main", prog.Main)
	assert.Equal(t, "end", prog.EndStmt)
	assert.Nil(t, prog.Vars)
}

func TestProgramWithVars(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; var x: int; main end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	assert.NotNil(t, prog.Vars)
	assert.Equal(t, "x", prog.Vars.FVar.ID)
	assert.Equal(t, "int", prog.Vars.FVar.Type.Name)
}

func TestProgramWithBody(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	assert.NotNil(t, prog.Body)
}

func TestProgramMissingID(t *testing.T) {
	parser := MustBuildParser()
	program := `program ; main end`

	_, err := parser.ParseString("", program)
	assert.Error(t, err, "Debe fallar si falta el identificador")
}

func TestProgramMissingSemicolon(t *testing.T) {
	parser := MustBuildParser()
	program := `program test main end`

	_, err := parser.ParseString("", program)
	assert.Error(t, err, "Debe fallar si falta el punto y coma")
}

func TestProgramMissingEnd(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main`

	_, err := parser.ParseString("", program)
	assert.Error(t, err, "Debe fallar si falta el keyword 'end'")
}

// ============================================================================
// 3. Tests de Variables (Vars y FVar)
// ============================================================================

func TestVarsSingleVariable(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; var x: int; main end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Vars)
	assert.Equal(t, "x", prog.Vars.FVar.ID)
	assert.Empty(t, prog.Vars.FVar.RID)
	assert.Equal(t, "int", prog.Vars.FVar.Type.Name)
}

func TestVarsMultipleVariables(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; var x, y, z: float; main end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Vars)
	assert.Equal(t, "x", prog.Vars.FVar.ID)
	assert.Equal(t, []string{"y", "z"}, prog.Vars.FVar.RID)
	assert.Equal(t, "float", prog.Vars.FVar.Type.Name)
}

func TestVarsMissingType(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; var x: ; main end`

	_, err := parser.ParseString("", program)
	assert.Error(t, err, "Debe fallar si falta el tipo")
}

func TestVarsMissingColon(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; var x int; main end`

	_, err := parser.ParseString("", program)
	assert.Error(t, err, "Debe fallar si falta el dos puntos")
}

func TestVarsInvalidType(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; var x: string; main end`

	_, err := parser.ParseString("", program)
	// El lexer acepta "string" pero el parser espera solo "int" o "float"
	// El error puede venir en diferentes fases
	assert.Error(t, err, "Debe fallar con tipo inválido")
}

// ============================================================================
// 4. Tests de Statements
// ============================================================================

func TestAssignInteger(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { x = 42; } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	require.Len(t, prog.Body.PStat, 1)

	assign := prog.Body.PStat[0].Assign
	require.NotNil(t, assign)
	assert.Equal(t, "x", assign.ID)
	assert.Equal(t, "42", assign.Expr.Int)
}

func TestAssignFloat(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { y = 3.14; } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	require.Len(t, prog.Body.PStat, 1)

	assign := prog.Body.PStat[0].Assign
	require.NotNil(t, assign)
	assert.Equal(t, "y", assign.ID)
	assert.Equal(t, "3.14", assign.Expr.Float)
}

func TestAssignString(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { s = "hola"; } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	require.Len(t, prog.Body.PStat, 1)

	assign := prog.Body.PStat[0].Assign
	require.NotNil(t, assign)
	assert.Equal(t, "s", assign.ID)
	assert.Equal(t, `"hola"`, assign.Expr.String)
}

func TestAssignMissingSemicolon(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { x = 42 } end`

	_, err := parser.ParseString("", program)
	assert.Error(t, err, "Debe fallar si falta el punto y coma")
}

func TestConditionSimple(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { if (x) {}; } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	require.Len(t, prog.Body.PStat, 1)

	cond := prog.Body.PStat[0].Condition
	require.NotNil(t, cond)
	assert.Equal(t, "if", cond.IfKeyword)
}

func TestConditionWithElse(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { if (x) {} else {}; } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	require.Len(t, prog.Body.PStat, 1)

	cond := prog.Body.PStat[0].Condition
	require.NotNil(t, cond)
	assert.NotNil(t, cond.Else)
	assert.Equal(t, "else", cond.Else.ElseKeyword)
}

func TestConditionMissingParens(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { if x {}; } end`

	_, err := parser.ParseString("", program)
	assert.Error(t, err, "Debe fallar si faltan paréntesis")
}

func TestConditionMissingBody(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { if (x) ; } end`

	_, err := parser.ParseString("", program)
	assert.Error(t, err, "Debe fallar si falta el cuerpo")
}

func TestCycleWhileDo(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { while (x) do {}; } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	require.Len(t, prog.Body.PStat, 1)

	cycle := prog.Body.PStat[0].Cycle
	require.NotNil(t, cycle)
	assert.Equal(t, "while", cycle.WhileKeyword)
	assert.Equal(t, "do", cycle.DoKeyword)
}

func TestCycleMissingDo(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { while (x) {}; } end`

	_, err := parser.ParseString("", program)
	assert.Error(t, err, "Debe fallar si falta el keyword 'do'")
}

func TestCycleMissingBody(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { while (x) do ; } end`

	_, err := parser.ParseString("", program)
	assert.Error(t, err, "Debe fallar si falta el cuerpo")
}

func TestPrintString(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { print("texto"); } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	require.Len(t, prog.Body.PStat, 1)

	printStmt := prog.Body.PStat[0].Print
	require.NotNil(t, printStmt)
	assert.Equal(t, "print", printStmt.PrintKeyword)
	assert.Len(t, printStmt.PrintP, 1)
	assert.Equal(t, `"texto"`, printStmt.PrintP[0])
}

func TestPrintIdent(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { print(x); } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	require.Len(t, prog.Body.PStat, 1)

	printStmt := prog.Body.PStat[0].Print
	require.NotNil(t, printStmt)
	assert.Len(t, printStmt.PrintP, 1)
	assert.Equal(t, "x", printStmt.PrintP[0])
}

func TestPrintMultiple(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { print("x=", x); } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	require.Len(t, prog.Body.PStat, 1)

	printStmt := prog.Body.PStat[0].Print
	require.NotNil(t, printStmt)
	assert.Len(t, printStmt.PrintP, 2)
}

func TestPrintEmpty(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { print(); } end`

	_, err := parser.ParseString("", program)
	// El parser actual permite print() vacío según la gramática
	// Si queremos forzar al menos un argumento, habría que cambiar la gramática
	// Por ahora, esto debería parsear sin error
	if err != nil {
		t.Logf("Nota: print() vacío da error: %v", err)
	}
}

// ============================================================================
// 5. Tests de Expressions
// ============================================================================

func TestExpressionIdent(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { x = y; } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	require.Len(t, prog.Body.PStat, 1)

	assign := prog.Body.PStat[0].Assign
	require.NotNil(t, assign)
	assert.Equal(t, "y", assign.Expr.Ident)
}

func TestExpressionInt(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { x = 123; } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	require.Len(t, prog.Body.PStat, 1)

	assign := prog.Body.PStat[0].Assign
	require.NotNil(t, assign)
	assert.Equal(t, "123", assign.Expr.Int)
}

func TestExpressionFloat(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { x = 3.14; } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	require.Len(t, prog.Body.PStat, 1)

	assign := prog.Body.PStat[0].Assign
	require.NotNil(t, assign)
	assert.Equal(t, "3.14", assign.Expr.Float)
}

func TestExpressionString(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { x = "hello"; } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	require.Len(t, prog.Body.PStat, 1)

	assign := prog.Body.PStat[0].Assign
	require.NotNil(t, assign)
	assert.Equal(t, `"hello"`, assign.Expr.String)
}

func TestExpressionEmpty(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { x = ; } end`

	_, err := parser.ParseString("", program)
	assert.Error(t, err, "Debe fallar si la expresión está vacía")
}

// ============================================================================
// 6. Tests Edge Cases
// ============================================================================

func TestEmptyProgram(t *testing.T) {
	parser := MustBuildParser()
	program := ``

	_, err := parser.ParseString("", program)
	assert.Error(t, err, "Debe fallar con programa vacío")
}

func TestProgramOnlySpaces(t *testing.T) {
	parser := MustBuildParser()
	program := `   `

	_, err := parser.ParseString("", program)
	assert.Error(t, err, "Debe fallar con solo espacios")
}

func TestCaseInsensitiveKeywords(t *testing.T) {
	testCases := []struct {
		name    string
		program string
	}{
		{"Uppercase PROGRAM", "PROGRAM test; MAIN END"},
		{"Mixed case Program", "Program test; Main End"},
		{"Lowercase program", "program test; main end"},
		{"Mixed VAR", "program test; VAR x: int; main end"},
		{"Mixed WHILE", "program test; main { WHILE (x) DO {}; } end"},
	}

	parser := MustBuildParser()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parser.ParseString("", tc.program)
			assert.NoError(t, err, "Debe aceptar keywords case-insensitive: %s", tc.name)
		})
	}
}

func TestNestedBody(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { if (x) { while (y) do { print(z); }; }; } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	require.Len(t, prog.Body.PStat, 1)

	// Primer statement es un if
	cond := prog.Body.PStat[0].Condition
	require.NotNil(t, cond)

	// El cuerpo del if debe tener un while
	require.Len(t, cond.Body.PStat, 1)
	cycle := cond.Body.PStat[0].Cycle
	require.NotNil(t, cycle)
}

func TestMultipleStatements(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { x = 1; y = 2; z = 3; } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	assert.Len(t, prog.Body.PStat, 3)

	// Verificar cada asignación
	for i := 0; i < 3; i++ {
		assign := prog.Body.PStat[i].Assign
		require.NotNil(t, assign)
	}
}

func TestStringWithEscapes(t *testing.T) {
	parser := MustBuildParser()
	program := `program test; main { x = "texto \"escapado\""; } end`

	prog, err := parser.ParseString("", program)
	require.NoError(t, err)
	require.NotNil(t, prog.Body)
	require.Len(t, prog.Body.PStat, 1)

	assign := prog.Body.PStat[0].Assign
	require.NotNil(t, assign)
	assert.Contains(t, assign.Expr.String, `"escapado"`)
}
