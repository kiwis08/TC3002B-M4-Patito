package semantic

import (
	"fmt"

	"Patito/token"
)

// ScopeKind ayuda a describir el contexto en el que se declara un símbolo.
type ScopeKind int

const (
	ScopeGlobal ScopeKind = iota
	ScopeLocal
	ScopeParam
)

func (s ScopeKind) String() string {
	switch s {
	case ScopeGlobal:
		return "global"
	case ScopeLocal:
		return "local"
	case ScopeParam:
		return "parámetros"
	default:
		return "desconocido"
	}
}

// VariableSpec es el resultado intermedio que producen las acciones del parser.
// Agrupa nombre, tipo y posición para luego insertarlo en la tabla definitiva.
type VariableSpec struct {
	Name    string
	Type    Type
	Pos     token.Pos
	Address int // Dirección virtual asignada
}

// VariableEntry es el elemento almacenado en una tabla después de validar duplicados.
type VariableEntry struct {
	Name       string
	Type       Type
	Scope      ScopeKind
	DeclaredAt token.Pos
	Address    int // Dirección virtual asignada
}

// VariableTable mantiene las variables de un determinado scope.
type VariableTable struct {
	kind    ScopeKind
	entries map[string]*VariableEntry
	order   []*VariableEntry
}

func NewVariableTable(kind ScopeKind) *VariableTable {
	return &VariableTable{
		kind:    kind,
		entries: make(map[string]*VariableEntry),
		order:   make([]*VariableEntry, 0),
	}
}

func (vt *VariableTable) Add(spec *VariableSpec) error {
	if spec == nil {
		return fmt.Errorf("variable spec nil")
	}
	if existing, ok := vt.entries[spec.Name]; ok {
		return &DuplicateSymbolError{
			Name:      spec.Name,
			Scope:     vt.kind,
			FirstPos:  existing.DeclaredAt,
			SecondPos: spec.Pos,
		}
	}
	entry := &VariableEntry{
		Name:       spec.Name,
		Type:       spec.Type,
		Scope:      vt.kind,
		DeclaredAt: spec.Pos,
		Address:    spec.Address, // Dirección virtual asignada
	}
	vt.entries[spec.Name] = entry
	vt.order = append(vt.order, entry)
	return nil
}

func (vt *VariableTable) AddMany(specs []*VariableSpec) error {
	for _, spec := range specs {
		if err := vt.Add(spec); err != nil {
			return err
		}
	}
	return nil
}

func (vt *VariableTable) Has(name string) bool {
	_, ok := vt.entries[name]
	return ok
}

func (vt *VariableTable) Entries() []*VariableEntry {
	result := make([]*VariableEntry, len(vt.order))
	copy(result, vt.order)
	return result
}

// FunctionEntry representa una función registrada en el directorio.
type FunctionEntry struct {
	Name       string
	ReturnType Type
	DeclaredAt token.Pos

	Params *VariableTable
	Locals *VariableTable
}

// FunctionDirectory centraliza la tabla de funciones y la tabla global.
type FunctionDirectory struct {
	ProgramName string
	ProgramPos  token.Pos

	Globals   *VariableTable
	Functions map[string]*FunctionEntry
}

func NewFunctionDirectory() *FunctionDirectory {
	return &FunctionDirectory{
		Globals:   NewVariableTable(ScopeGlobal),
		Functions: make(map[string]*FunctionEntry),
	}
}

func (fd *FunctionDirectory) SetProgram(name string, pos token.Pos) error {
	if fd.ProgramName == "" {
		fd.ProgramName = name
		fd.ProgramPos = pos
		return nil
	}
	return &ProgramRedefinitionError{
		Name:         name,
		Existing:     fd.ProgramName,
		ExistingPos:  fd.ProgramPos,
		RedeclaredAt: pos,
	}
}

func (fd *FunctionDirectory) AddGlobals(specs []*VariableSpec, addressManager *VirtualAddressManager) error {
	if len(specs) == 0 {
		return nil
	}
	// Asignar direcciones virtuales a variables globales
	for _, spec := range specs {
		if spec.Address == 0 { // Si no tiene dirección asignada
			spec.Address = addressManager.NextGlobal()
		}
	}
	return fd.Globals.AddMany(specs)
}

func (fd *FunctionDirectory) AddFunction(name string, returnType Type, pos token.Pos, params []*VariableSpec, locals []*VariableSpec, addressManager *VirtualAddressManager) (*FunctionEntry, error) {
	if existing, ok := fd.Functions[name]; ok {
		return nil, &FunctionRedefinitionError{
			Name:         name,
			ExistingPos:  existing.DeclaredAt,
			RedeclaredAt: pos,
		}
	}

	// Reiniciar contador de locales para esta función
	addressManager.ResetLocals()

	// Asignar direcciones virtuales a parámetros
	for _, spec := range params {
		if spec.Address == 0 { // Si no tiene dirección asignada
			spec.Address = addressManager.NextLocal()
		}
	}

	paramTable := NewVariableTable(ScopeParam)
	if err := paramTable.AddMany(params); err != nil {
		return nil, err
	}

	localTable := NewVariableTable(ScopeLocal)
	// Asignar direcciones virtuales a variables locales
	for _, spec := range locals {
		if spec.Address == 0 { // Si no tiene dirección asignada
			spec.Address = addressManager.NextLocal()
		}
		if paramTable.Has(spec.Name) {
			return nil, &DuplicateSymbolError{
				Name:      spec.Name,
				Scope:     ScopeLocal,
				FirstPos:  paramTable.entries[spec.Name].DeclaredAt,
				SecondPos: spec.Pos,
			}
		}
		if err := localTable.Add(spec); err != nil {
			return nil, err
		}
	}

	fn := &FunctionEntry{
		Name:       name,
		ReturnType: returnType,
		DeclaredAt: pos,
		Params:     paramTable,
		Locals:     localTable,
	}

	fd.Functions[name] = fn
	return fn, nil
}

func (fd *FunctionDirectory) GetFunction(name string) (*FunctionEntry, bool) {
	fn, ok := fd.Functions[name]
	return fn, ok
}

// GetVariableType busca una variable en el directorio y devuelve su tipo
// Busca primero en globales, luego en función main (si existe)
// Nota: En el cuerpo de main, las variables globales están disponibles
func (fd *FunctionDirectory) GetVariableType(name string) (Type, error) {
	// Buscar en globales primero (disponibles en todo el programa)
	if entry, ok := fd.Globals.entries[name]; ok {
		return entry.Type, nil
	}

	// Buscar en función main (si existe) - locales y parámetros
	if mainFn, ok := fd.Functions["main"]; ok {
		if entry, ok := mainFn.Locals.entries[name]; ok {
			return entry.Type, nil
		}
		if entry, ok := mainFn.Params.entries[name]; ok {
			return entry.Type, nil
		}
	}

	return TypeInvalid, fmt.Errorf("variable '%s' no declarada", name)
}

// GetVariableTypeFromContext busca una variable primero en el contexto (tipos temporales)
// y luego en el directorio
func GetVariableTypeFromContext(ctx *Context, name string) (Type, error) {
	// Buscar primero en VariableTypes (tipos temporales durante parsing)
	if varType, ok := ctx.VariableTypes[name]; ok {
		return varType, nil
	}

	// Buscar en el directorio
	return ctx.Directory.GetVariableType(name)
}

// GetVariableAddressFromContext busca la dirección virtual de una variable
func GetVariableAddressFromContext(ctx *Context, name string) (int, error) {
	// Buscar primero en VariableAddresses (direcciones durante parsing)
	if addr, ok := ctx.VariableAddresses[name]; ok && addr != 0 {
		return addr, nil
	}

	// Si no está en VariableAddresses, verificar si la variable existe
	// Si existe pero no tiene dirección asignada todavía, asignarla ahora
	if _, ok := ctx.VariableTypes[name]; ok {
		// Variable existe pero dirección no asignada todavía
		// Asignar dirección temporalmente (será reemplazada cuando se agregue al directorio)
		// Esto puede pasar si la variable se usa antes de que reduceProgram termine
		// Por ahora, asignar una dirección global temporal
		addr := ctx.AddressManager.NextGlobal()
		ctx.VariableAddresses[name] = addr
		return addr, nil
	}

	// Si no está en VariableAddresses ni en VariableTypes, buscar en el directorio
	return ctx.Directory.GetVariableAddress(name)
}

// GetVariableAddress busca una variable en el directorio y devuelve su dirección virtual
func (fd *FunctionDirectory) GetVariableAddress(name string) (int, error) {
	// Buscar en globales primero
	if entry, ok := fd.Globals.entries[name]; ok {
		if entry.Address == 0 {
			return 0, fmt.Errorf("variable '%s' tiene dirección 0 (no asignada)", name)
		}
		return entry.Address, nil
	}

	// Debug: verificar si la variable está en el directorio pero con otro nombre
	// (esto no debería ser necesario, pero ayuda a debuggear)

	// Buscar en función main (si existe)
	if mainFn, ok := fd.Functions["main"]; ok {
		if entry, ok := mainFn.Locals.entries[name]; ok {
			if entry.Address == 0 {
				return 0, fmt.Errorf("variable '%s' tiene dirección 0 (no asignada)", name)
			}
			return entry.Address, nil
		}
		if entry, ok := mainFn.Params.entries[name]; ok {
			if entry.Address == 0 {
				return 0, fmt.Errorf("variable '%s' tiene dirección 0 (no asignada)", name)
			}
			return entry.Address, nil
		}
	}

	return 0, fmt.Errorf("variable '%s' no declarada", name)
}
