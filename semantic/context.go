package semantic

import "Patito/token"

// Context es el objeto que asignamos a parser.Context para compartir estado
// entre las acciones semánticas.
type Context struct {
	Directory      *FunctionDirectory
	Cube           *SemanticCube
	Quadruples     *QuadrupleQueue
	OpStack        *OperatorStack
	OperandStack   *OperandStack
	TypeStack      *TypeStack
	JumpStack      *JumpStack
	TempCounter    *TempCounter
	AddressManager *VirtualAddressManager
	ConstantTable  *ConstantTable
	// VariableTypes almacena tipos de variables mientras se procesan (antes de agregar al directorio)
	VariableTypes map[string]Type
	// VariableAddresses almacena direcciones virtuales de variables mientras se procesan
	VariableAddresses map[string]int
	// FunctionStack tracks the function currently being processed for return statement validation
	// We use a stack to handle nested scopes (if needed in the future)
	FunctionStack []*FunctionEntry
	// HasReturn tracks if the current function has at least one return statement
	HasReturn bool
	// CurrentFunctionName tracks the name of the function currently being processed
	// This is set when we start processing a function and used by return statements
	// CurrentFunctionName string
	// PendingFunctionName tracks the function name that will be processed
	// This is used as a workaround for bottom-up parsing where body is processed before reduceFunction
	PendingFunctionName string
	PendingReturns      []PendingReturn
	FunctionStartQuads  map[string]int
	// ProgramStartGotoIndex stores the index of the GOTO quadruple at program start
	// This will be filled when the main function body starts
	ProgramStartGotoIndex int
	// MainStartIndex tracks when main body actually starts (for filling GOTO)
	MainStartIndex       int
	LastFunctionEndIndex int
}

type PendingReturn struct {
	Value    string
	Type     Type
	Pos      token.Pos
	Function string
}

// // CurrentFunction returns the function currently being processed
// func (c *Context) CurrentFunction() *FunctionEntry {
// 	if len(c.FunctionStack) == 0 {
// 		return nil
// 	}
// 	return c.FunctionStack[len(c.FunctionStack)-1]
// }

// PushFunction pushes a function onto the function stack
func (c *Context) PushFunction(fn *FunctionEntry) {
	c.FunctionStack = append(c.FunctionStack, fn)
	c.HasReturn = false
}

// // PopFunction pops a function from the function stack
// func (c *Context) PopFunction() *FunctionEntry {
// 	if len(c.FunctionStack) == 0 {
// 		return nil
// 	}
// 	fn := c.FunctionStack[len(c.FunctionStack)-1]
// 	c.FunctionStack = c.FunctionStack[:len(c.FunctionStack)-1]
// 	return fn
// }

func NewContext() *Context {
	addressManager := NewVirtualAddressManager()
	tempCounter := NewTempCounter()
	tempCounter.SetAddressManager(addressManager)

	ctx := &Context{
		Directory:             NewFunctionDirectory(),
		Cube:                  DefaultSemanticCube,
		Quadruples:            NewQuadrupleQueue(),
		OpStack:               NewOperatorStack(),
		OperandStack:          NewOperandStack(),
		TypeStack:             NewTypeStack(),
		JumpStack:             NewJumpStack(),
		TempCounter:           tempCounter,
		AddressManager:        addressManager,
		ConstantTable:         NewConstantTable(),
		VariableTypes:         make(map[string]Type),
		VariableAddresses:     make(map[string]int),
		PendingReturns:        make([]PendingReturn, 0),
		FunctionStartQuads:    make(map[string]int),
		ProgramStartGotoIndex: -1,
		MainStartIndex:        -1,
	}

	return ctx
}

// DirectorySnapshot expone el directorio final tras el parseo.
func (c *Context) DirectorySnapshot() *FunctionDirectory {
	return c.Directory
}
