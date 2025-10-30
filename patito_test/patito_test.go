package parser_test

import (
	"testing"

	pwrap "Patito/pkg/parser"

	"github.com/stretchr/testify/assert"
)

func mustParser(t *testing.T) *parserWithWrap {
	t.Helper()
	return &parserWithWrap{p: pwrap.MustBuildParser()}
}

type parserWithWrap struct {
	p interface {
		ParseString(string, string) (interface{}, error)
	}
}

// --- Helpers (thin) ---

func parseOK(t *testing.T, code string) {
	t.Helper()
	p := pwrap.MustBuildParser()
	_, err := pwrap.ParseString(p, "", code)
	assert.NoError(t, err, "Debe parsear OK:\n%s", code)
}

func parseErr(t *testing.T, code string) {
	t.Helper()
	p := pwrap.MustBuildParser()
	_, err := pwrap.ParseString(p, "", code)
	assert.Error(t, err, "Debe fallar el parseo:\n%s", code)
}

// 1) Palabras clave y estructura base

func TestProgram_Minimo(t *testing.T) {
	parseOK(t, `program p; main { } end`)
}

func TestProgram_ConVars(t *testing.T) {
	parseOK(t, `program p; var x:int; main { } end`)
}

func TestProgram_FaltaID(t *testing.T) {
	parseErr(t, `program ; main { } end`)
}

func TestProgram_FaltaPuntoYComa(t *testing.T) {
	parseErr(t, `program p main { } end`)
}

func TestProgram_FaltaEnd(t *testing.T) {
	parseErr(t, `program p; main { }`)
}

// 2) Variables

func TestVars_UnaLinea(t *testing.T) {
	parseOK(t, `program p; var x:int; main { } end`)
}

func TestVars_MultiplesNombres(t *testing.T) {
	parseOK(t, `program p; var x,y,z:float; main { } end`)
}

func TestVars_TipoInvalido(t *testing.T) {
	parseErr(t, `program p; var x:string; main { } end`) // sólo int|float
}

func TestVars_FaltaColon(t *testing.T) {
	parseErr(t, `program p; var x int; main { } end`)
}

// 3) Statements: assign, if/else, while, print, call

func TestAssign_OK(t *testing.T) {
	parseOK(t, `program p; main { x = 42; } end`)
	parseOK(t, `program p; main { y = (1+2)*3; } end`)
}

func TestAssign_SinPuntoYComa(t *testing.T) {
	parseErr(t, `program p; main { x = 1 } end`)
}

func TestIf_Simple(t *testing.T) {
	parseOK(t, `program p; main { if (x > 0) { x = 1; }; } end`)
}

func TestIf_ConElse(t *testing.T) {
	parseOK(t, `program p; main { if (x < 0) { x = 1; } else { x = 2; }; } end`)
}

func TestIf_SinParentesis(t *testing.T) {
	parseErr(t, `program p; main { if x { x = 1; }; } end`)
}

func TestWhile_OK(t *testing.T) {
	parseOK(t, `program p; main { while (x != 0) do { x = x - 1; }; } end`)
}

func TestWhile_SinDo(t *testing.T) {
	parseErr(t, `program p; main { while (x) { x = 1; }; } end`)
}

func TestPrint_StringYExpr(t *testing.T) {
	parseOK(t, `program p; main { print("x=", x+1); } end`)
}

func TestPrint_VacioNoPermitido(t *testing.T) {
	parseErr(t, `program p; main { print(); } end`)
}

func TestBracket_Block(t *testing.T) {
	parseOK(t, `program p; main { [ x = 1; y = 2; ] } end`)
}

func TestCall_AsStatement(t *testing.T) {
	parseOK(t, `program p; main { foo(1,2,3); } end`)
}

func TestCall_AsFactor(t *testing.T) {
	parseOK(t, `program p; main { x = (foo(1,2*3)); } end`)
}

// 4) Funciones

func TestFunc_SinParams_SinLocals(t *testing.T) {
	parseOK(t, `program p; void f()[]{}; main { } end`)
}

func TestFunc_ConParamsYLocals(t *testing.T) {
	parseOK(t, `
		program p;
		void f(a:int, b:float)[ var x:int; ] { print("ok"); };
		main { f(1,2.0); }
		end`)
}

// 5) Expresiones (precedencia y relaciones)

func TestExpr_Precedencia(t *testing.T) {
	parseOK(t, `program p; main { x = 1 + 2 * 3; } end`)
	parseOK(t, `program p; main { x = (1 + 2) * 3; } end`)
}

func TestExpr_Relacionales(t *testing.T) {
	parseOK(t, `program p; main { x = 1; if (x > 0) { }; } end`)
	parseOK(t, `program p; main { if (x != 0) { }; } end`)
}

// 6) Casos borde

func TestVacio(t *testing.T) {
	parseErr(t, ``)
}

func TestSoloEspacios(t *testing.T) {
	parseErr(t, `   `)
}
