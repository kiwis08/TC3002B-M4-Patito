package semantic

// Context es el objeto que asignamos a parser.Context para compartir estado
// entre las acciones semánticas.
type Context struct {
	Directory    *FunctionDirectory
	Cube         *SemanticCube
	Quadruples   *QuadrupleQueue
	OpStack      *OperatorStack
	OperandStack *OperandStack
	TypeStack    *TypeStack
	JumpStack    *JumpStack
	TempCounter  *TempCounter
	// VariableTypes almacena tipos de variables mientras se procesan (antes de agregar al directorio)
	VariableTypes map[string]Type
}

func NewContext() *Context {
	return &Context{
		Directory:     NewFunctionDirectory(),
		Cube:          DefaultSemanticCube,
		Quadruples:    NewQuadrupleQueue(),
		OpStack:       NewOperatorStack(),
		OperandStack:  NewOperandStack(),
		TypeStack:     NewTypeStack(),
		JumpStack:     NewJumpStack(),
		TempCounter:   NewTempCounter(),
		VariableTypes: make(map[string]Type),
	}
}

// DirectorySnapshot expone el directorio final tras el parseo.
func (c *Context) DirectorySnapshot() *FunctionDirectory {
	return c.Directory
}
