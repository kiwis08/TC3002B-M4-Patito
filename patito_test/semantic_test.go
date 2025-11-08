package parser_test

import (
	"testing"

	pwrap "Patito/pkg/parser"
	"Patito/semantic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSemanticDirectoryBuilt(t *testing.T) {
	parser := pwrap.MustBuildParser()
	res, err := pwrap.ParseString(parser, "", `
		program demo;
		var x:int;
		void foo(a:int)[ var y:float; ] { y = a; };
		main { foo(1); }
		end`)
	require.NoError(t, err)

	dir, ok := res.(*semantic.FunctionDirectory)
	require.True(t, ok, "Parse debe regresar *semantic.FunctionDirectory")
	assert.Equal(t, "demo", dir.ProgramName)
	assert.True(t, dir.Globals.Has("x"))

	fn, exists := dir.GetFunction("foo")
	require.True(t, exists, "función foo debe registrarse")
	assert.Equal(t, semantic.TypeVoid, fn.ReturnType)

	params := fn.Params.Entries()
	require.Len(t, params, 1)
	assert.Equal(t, "a", params[0].Name)
	assert.Equal(t, semantic.TypeInt, params[0].Type)

	locals := fn.Locals.Entries()
	require.Len(t, locals, 1)
	assert.Equal(t, "y", locals[0].Name)
	assert.Equal(t, semantic.TypeFloat, locals[0].Type)
}

func TestSemanticDuplicateGlobalVariable(t *testing.T) {
	parser := pwrap.MustBuildParser()
	_, err := pwrap.ParseString(parser, "", `
		program demo;
		var x:int;
		x:float;
		main { }
		end`)
	assert.Error(t, err, "debe detectar variable global duplicada")
}

func TestSemanticDuplicateParameter(t *testing.T) {
	parser := pwrap.MustBuildParser()
	_, err := pwrap.ParseString(parser, "", `
		program demo;
		void foo(a:int, a:float)[] { };
		main { }
		end`)
	assert.Error(t, err, "debe detectar parámetro duplicado")
}

func TestSemanticDuplicateLocalVariable(t *testing.T) {
	parser := pwrap.MustBuildParser()
	_, err := pwrap.ParseString(parser, "", `
		program demo;
		void foo(a:int)[ var x:int; x:float; ] { };
		main { }
		end`)
	assert.Error(t, err, "debe detectar variable local duplicada")
}

func TestSemanticParamLocalClash(t *testing.T) {
	parser := pwrap.MustBuildParser()
	_, err := pwrap.ParseString(parser, "", `
		program demo;
		void foo(a:int)[ var a:float; ] { };
		main { }
		end`)
	assert.Error(t, err, "debe detectar choque entre parámetro y local")
}

func TestSemanticFunctionRedefinition(t *testing.T) {
	parser := pwrap.MustBuildParser()
	_, err := pwrap.ParseString(parser, "", `
		program demo;
		void foo()[] { };
		void foo()[] { };
		main { }
		end`)
	assert.Error(t, err, "debe detectar redefinición de función")
}
