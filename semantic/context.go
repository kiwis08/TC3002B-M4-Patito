package semantic

// Context es el objeto que asignamos a parser.Context para compartir estado
// entre las acciones semánticas.
type Context struct {
	Directory         *FunctionDirectory
	Cube              *SemanticCube
	Quadruples        *QuadrupleQueue
	OpStack           *OperatorStack
	OperandStack      *OperandStack
	TypeStack         *TypeStack
	JumpStack         *JumpStack
	TempCounter       *TempCounter
	AddressManager    *VirtualAddressManager
	ConstantTable     *ConstantTable
	// VariableTypes almacena tipos de variables mientras se procesan (antes de agregar al directorio)
	VariableTypes map[string]Type
	// VariableAddresses almacena direcciones virtuales de variables mientras se procesan
	VariableAddresses map[string]int
}

func NewContext() *Context {
	addressManager := NewVirtualAddressManager()
	tempCounter := NewTempCounter()
	tempCounter.SetAddressManager(addressManager)
	
	return &Context{
		Directory:         NewFunctionDirectory(),
		Cube:              DefaultSemanticCube,
		Quadruples:        NewQuadrupleQueue(),
		OpStack:           NewOperatorStack(),
		OperandStack:      NewOperandStack(),
		TypeStack:         NewTypeStack(),
		JumpStack:         NewJumpStack(),
		TempCounter:       tempCounter,
		AddressManager:    addressManager,
		ConstantTable:     NewConstantTable(),
		VariableTypes:      make(map[string]Type),
		VariableAddresses: make(map[string]int),
	}
}

// DirectorySnapshot expone el directorio final tras el parseo.
func (c *Context) DirectorySnapshot() *FunctionDirectory {
	return c.Directory
}
