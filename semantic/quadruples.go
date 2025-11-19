package semantic

import "fmt"

// Quadruple representa un cuádruplo en el código intermedio
// Formato: (operador, operando1, operando2, resultado)
type Quadruple struct {
	Operator string // Operador: +, -, *, /, >, <, !=, ==, =, GOTO, GOTOF, GOTOV, etc.
	Operand1 string // Primer operando (puede ser vacío para operaciones unarias)
	Operand2 string // Segundo operando (puede ser vacío para operaciones unarias)
	Result   string // Resultado (dirección temporal o variable)
}

// String devuelve una representación legible del cuádruplo
func (q Quadruple) String() string {
	return fmt.Sprintf("(%s, %s, %s, %s)", q.Operator, q.Operand1, q.Operand2, q.Result)
}

// QuadrupleQueue es una fila (cola) para almacenar cuádruplos
type QuadrupleQueue struct {
	quadruples []Quadruple
}

// NewQuadrupleQueue crea una nueva fila de cuádruplos
func NewQuadrupleQueue() *QuadrupleQueue {
	return &QuadrupleQueue{
		quadruples: make([]Quadruple, 0),
	}
}

// Enqueue agrega un cuádruplo al final de la fila
func (q *QuadrupleQueue) Enqueue(quad Quadruple) {
	q.quadruples = append(q.quadruples, quad)
}

// Get devuelve todos los cuádruplos en orden
func (q *QuadrupleQueue) Get() []Quadruple {
	return q.quadruples
}

// Size devuelve el número de cuádruplos en la fila
func (q *QuadrupleQueue) Size() int {
	return len(q.quadruples)
}

// NextIndex devuelve el siguiente índice disponible (para GOTO)
func (q *QuadrupleQueue) NextIndex() int {
	return len(q.quadruples)
}

// GetAt devuelve el cuádruplo en el índice especificado
func (q *QuadrupleQueue) GetAt(index int) *Quadruple {
	if index < 0 || index >= len(q.quadruples) {
		return nil
	}
	return &q.quadruples[index]
}

// UpdateAt actualiza el cuádruplo en el índice especificado
func (q *QuadrupleQueue) UpdateAt(index int, quad Quadruple) {
	if index >= 0 && index < len(q.quadruples) {
		q.quadruples[index] = quad
	}
}

// String devuelve una representación legible de todos los cuádruplos
func (q *QuadrupleQueue) String() string {
	if len(q.quadruples) == 0 {
		return "Fila de cuádruplos vacía\n"
	}
	result := "Fila de cuádruplos:\n"
	for i, quad := range q.quadruples {
		result += fmt.Sprintf("  %d: %s\n", i, quad.String())
	}
	return result
}

// OperatorStack es una pila para operadores
type OperatorStack struct {
	operators []string
}

// NewOperatorStack crea una nueva pila de operadores
func NewOperatorStack() *OperatorStack {
	return &OperatorStack{
		operators: make([]string, 0),
	}
}

// Push agrega un operador a la pila
func (s *OperatorStack) Push(op string) {
	s.operators = append(s.operators, op)
}

// Pop elimina y devuelve el operador del tope de la pila
func (s *OperatorStack) Pop() (string, bool) {
	if len(s.operators) == 0 {
		return "", false
	}
	top := s.operators[len(s.operators)-1]
	s.operators = s.operators[:len(s.operators)-1]
	return top, true
}

// Top devuelve el operador del tope sin eliminarlo
func (s *OperatorStack) Top() (string, bool) {
	if len(s.operators) == 0 {
		return "", false
	}
	return s.operators[len(s.operators)-1], true
}

// IsEmpty verifica si la pila está vacía
func (s *OperatorStack) IsEmpty() bool {
	return len(s.operators) == 0
}

// OperandStack es una pila para operandos (direcciones/variables)
type OperandStack struct {
	operands []string
}

// NewOperandStack crea una nueva pila de operandos
func NewOperandStack() *OperandStack {
	return &OperandStack{
		operands: make([]string, 0),
	}
}

// Push agrega un operando a la pila
func (s *OperandStack) Push(operand string) {
	s.operands = append(s.operands, operand)
}

// Pop elimina y devuelve el operando del tope de la pila
func (s *OperandStack) Pop() (string, bool) {
	if len(s.operands) == 0 {
		return "", false
	}
	top := s.operands[len(s.operands)-1]
	s.operands = s.operands[:len(s.operands)-1]
	return top, true
}

// Top devuelve el operando del tope sin eliminarlo
func (s *OperandStack) Top() (string, bool) {
	if len(s.operands) == 0 {
		return "", false
	}
	return s.operands[len(s.operands)-1], true
}

// IsEmpty verifica si la pila está vacía
func (s *OperandStack) IsEmpty() bool {
	return len(s.operands) == 0
}

// TypeStack es una pila para tipos semánticos
type TypeStack struct {
	types []Type
}

// NewTypeStack crea una nueva pila de tipos
func NewTypeStack() *TypeStack {
	return &TypeStack{
		types: make([]Type, 0),
	}
}

// Push agrega un tipo a la pila
func (s *TypeStack) Push(t Type) {
	s.types = append(s.types, t)
}

// Pop elimina y devuelve el tipo del tope de la pila
func (s *TypeStack) Pop() (Type, bool) {
	if len(s.types) == 0 {
		return TypeInvalid, false
	}
	top := s.types[len(s.types)-1]
	s.types = s.types[:len(s.types)-1]
	return top, true
}

// Top devuelve el tipo del tope sin eliminarlo
func (s *TypeStack) Top() (Type, bool) {
	if len(s.types) == 0 {
		return TypeInvalid, false
	}
	return s.types[len(s.types)-1], true
}

// IsEmpty verifica si la pila está vacía
func (s *TypeStack) IsEmpty() bool {
	return len(s.types) == 0
}

// JumpStack es una pila para almacenar índices de cuádruplos (para GOTO pendientes)
type JumpStack struct {
	jumps []int
}

// NewJumpStack crea una nueva pila de saltos
func NewJumpStack() *JumpStack {
	return &JumpStack{
		jumps: make([]int, 0),
	}
}

// Push agrega un índice de salto a la pila
func (s *JumpStack) Push(index int) {
	s.jumps = append(s.jumps, index)
}

// Pop elimina y devuelve el índice del tope de la pila
func (s *JumpStack) Pop() (int, bool) {
	if len(s.jumps) == 0 {
		return -1, false
	}
	top := s.jumps[len(s.jumps)-1]
	s.jumps = s.jumps[:len(s.jumps)-1]
	return top, true
}

// Top devuelve el índice del tope sin eliminarlo
func (s *JumpStack) Top() (int, bool) {
	if len(s.jumps) == 0 {
		return -1, false
	}
	return s.jumps[len(s.jumps)-1], true
}

// IsEmpty verifica si la pila está vacía
func (s *JumpStack) IsEmpty() bool {
	return len(s.jumps) == 0
}

// TempCounter mantiene un contador para generar direcciones temporales únicas
// Ahora usa direcciones virtuales en lugar de nombres
type TempCounter struct {
	addressManager *VirtualAddressManager
}

// NewTempCounter crea un nuevo contador de temporales
func NewTempCounter() *TempCounter {
	return &TempCounter{}
}

// SetAddressManager establece el gestor de direcciones virtuales
func (tc *TempCounter) SetAddressManager(vam *VirtualAddressManager) {
	tc.addressManager = vam
}

// Next genera la siguiente dirección virtual temporal
func (tc *TempCounter) Next() int {
	if tc.addressManager == nil {
		panic("TempCounter: addressManager no inicializado")
	}
	return tc.addressManager.NextTemporal()
}

// NextString genera la siguiente dirección temporal como string
func (tc *TempCounter) NextString() string {
	return AddressToString(tc.Next())
}

// Reset reinicia el contador (reinicia el contador de temporales en el address manager)
func (tc *TempCounter) Reset() {
	if tc.addressManager != nil {
		tc.addressManager.ResetTemporals()
	}
}
