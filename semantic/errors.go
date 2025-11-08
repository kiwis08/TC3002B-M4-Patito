package semantic

import (
	"fmt"

	"Patito/token"
)

// DuplicateSymbolError se dispara al intentar declarar dos veces el mismo identificador.
type DuplicateSymbolError struct {
	Name      string
	Scope     ScopeKind
	FirstPos  token.Pos
	SecondPos token.Pos
}

func (e *DuplicateSymbolError) Error() string {
	return fmt.Sprintf("símbolo %q ya declarado en scope %s (%s); redeclarado en %s",
		e.Name, e.Scope, e.FirstPos, e.SecondPos)
}

// FunctionRedefinitionError indica que una función se declaró dos veces.
type FunctionRedefinitionError struct {
	Name         string
	ExistingPos  token.Pos
	RedeclaredAt token.Pos
}

func (e *FunctionRedefinitionError) Error() string {
	return fmt.Sprintf("función %q ya se había declarado en %s; redeclarada en %s",
		e.Name, e.ExistingPos, e.RedeclaredAt)
}

// ProgramRedefinitionError protege el encabezado `program`.
type ProgramRedefinitionError struct {
	Name         string
	Existing     string
	ExistingPos  token.Pos
	RedeclaredAt token.Pos
}

func (e *ProgramRedefinitionError) Error() string {
	return fmt.Sprintf("programa ya nombrado %q en %s; intento de renombrarlo a %q en %s",
		e.Existing, e.ExistingPos, e.Name, e.RedeclaredAt)
}
