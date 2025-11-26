package semantic

import (
	"Patito/token"
	"fmt"
	"os"
)

// GenerateQuadruple es un wrapper de generateQuadruple
func GenerateQuadruple(ctx *Context, op, op1, op2, result string) {
	generateQuadruple(ctx, op, op1, op2, result)
}

// getOperatorPrecedence devuelve la precedencia de un operador
// Mayor número = mayor precedencia
func getOperatorPrecedence(op string) int {
	switch op {
	case "*", "/":
		return 3
	case "+", "-":
		return 2
	case ">", "<", "!=", "==":
		return 1
	case "=":
		return 0
	default:
		return -1
	}
}

// generateQuadruple genera un cuádruplo y lo agrega a la fila
func generateQuadruple(ctx *Context, op, op1, op2, result string) {
	quad := Quadruple{
		Operator: op,
		Operand1: op1,
		Operand2: op2,
		Result:   result,
	}
	ctx.Quadruples.Enqueue(quad)
	index := ctx.Quadruples.Size() - 1
	fmt.Fprintf(os.Stderr, "[DEBUG] Quad %d: %s\n", index, quad.String())
}

// ProcessOperator procesa un operador según el algoritmo de traducción
func ProcessOperator(ctx *Context, op string) error {
	// Mientras haya operadores en la pila con mayor o igual precedencia
	for !ctx.OpStack.IsEmpty() {
		topOp, _ := ctx.OpStack.Top()
		if topOp == "(" {
			break // Paréntesis de apertura, no procesar
		}
		if getOperatorPrecedence(topOp) >= getOperatorPrecedence(op) {
			// Generar cuádruplo para el operador del tope
			ctx.OpStack.Pop()

			// Obtener operandos y tipos
			right, ok1 := ctx.OperandStack.Pop()
			rightType, _ := ctx.TypeStack.Pop()
			left, ok2 := ctx.OperandStack.Pop()
			leftType, _ := ctx.TypeStack.Pop()

			if !ok1 || !ok2 {
				return fmt.Errorf("error: operandos insuficientes para operador %s", topOp)
			}

			// Validar con cubo semántico
			operator := Operator(topOp)
			resultType, err := ctx.Cube.Result(operator, leftType, rightType)
			if err != nil {
				return err
			}

			// Generar temporal (dirección virtual)
			temp := ctx.TempCounter.NextString()

			// Generar cuádruplo
			generateQuadruple(ctx, topOp, left, right, temp)

			// Apilar resultado
			ctx.OperandStack.Push(temp)
			ctx.TypeStack.Push(resultType)
		} else {
			break
		}
	}

	// Apilar el nuevo operador
	ctx.OpStack.Push(op)
	return nil
}

// ProcessUnaryOperator procesa un operador unario
func ProcessUnaryOperator(ctx *Context, op string) error {
	// Para operadores unarios, solo necesitamos el operando
	operand, ok := ctx.OperandStack.Pop()
	if !ok {
		return fmt.Errorf("error: operando insuficiente para operador unario %s", op)
	}

	operandType, _ := ctx.TypeStack.Pop()

	// Validar con cubo semántico
	operator := Operator(op)
	resultType, err := ctx.Cube.ResultUnary(operator, operandType)
	if err != nil {
		return err
	}

	// Generar temporal (dirección virtual)
	temp := ctx.TempCounter.NextString()

	// Generar cuádruplo (operador unario, operando, vacío, resultado)
	generateQuadruple(ctx, op, operand, "", temp)

	// Apilar resultado
	PushOperand(ctx, temp, resultType)

	return nil
}

// PushOperand apila un operando y su tipo
func PushOperand(ctx *Context, operand string, operandType Type) {
	ctx.OperandStack.Push(operand)
	ctx.TypeStack.Push(operandType)
}

// PushConstant apila una constante y determina su tipo
// Asigna una dirección virtual a la constante si no existe
func PushConstant(ctx *Context, tok *token.Token) error {
	var operandType Type
	var value string

	switch tok.Type {
	case token.TokMap.Type("cte_int"):
		operandType = TypeInt
		value = string(tok.Lit)
	case token.TokMap.Type("cte_float"):
		operandType = TypeFloat
		value = string(tok.Lit)
	default:
		return fmt.Errorf("tipo de constante no soportado: %v", tok.Type)
	}

	// Buscar o crear entrada en tabla de constantes
	entry, exists := ctx.ConstantTable.Get(value, operandType)
	if !exists {
		// Crear nueva constante con dirección virtual
		address := ctx.AddressManager.NextConstant()
		entry = ctx.ConstantTable.Add(value, operandType, address)
	}

	// Apilar la dirección virtual como string
	operand := AddressToString(entry.Address)
	PushOperand(ctx, operand, operandType)
	return nil
}

// PushVariable apila una variable y busca su tipo y dirección virtual en el directorio
func PushVariable(ctx *Context, varName string, pos token.Pos) error {
	// Buscar variable primero en contexto, luego en directorio
	varType, err := GetVariableTypeFromContext(ctx, varName)
	if err != nil {
		return err
	}

	// Obtener dirección virtual
	address, err := GetVariableAddressFromContext(ctx, varName)
	if err != nil {
		return err
	}

	// Apilar la dirección virtual como string
	operand := AddressToString(address)
	PushOperand(ctx, operand, varType)
	return nil
}

// ProcessExpressionEnd procesa los operadores restantes al final de una expresión
// No procesa operadores relacionales (>, <, !=, ==) - estos se procesan en ProcessRelationalExpression
func ProcessExpressionEnd(ctx *Context) error {
	for !ctx.OpStack.IsEmpty() {
		op, _ := ctx.OpStack.Top()
		// No procesar operadores relacionales aquí
		if op == ">" || op == "<" || op == "!=" || op == "==" {
			break
		}
		if op == "(" {
			return fmt.Errorf("error: paréntesis no balanceado")
		}

		// Desapilar el operador
		ctx.OpStack.Pop()

		// Obtener operandos
		right, ok1 := ctx.OperandStack.Pop()
		rightType, _ := ctx.TypeStack.Pop()
		left, ok2 := ctx.OperandStack.Pop()
		leftType, _ := ctx.TypeStack.Pop()

		if !ok1 || !ok2 {
			return fmt.Errorf("error: operandos insuficientes para operador %s", op)
		}

		// Validar con cubo semántico
		operator := Operator(op)
		resultType, err := ctx.Cube.Result(operator, leftType, rightType)
		if err != nil {
			return err
		}

		// Generar temporal (dirección virtual)
		temp := ctx.TempCounter.NextString()

		// Generar cuádruplo
		generateQuadruple(ctx, op, left, right, temp)

		// Apilar resultado
		PushOperand(ctx, temp, resultType)
	}
	return nil
}

// ProcessAssignment procesa una asignación
func ProcessAssignment(ctx *Context, varName string, pos token.Pos) error {
	// Procesar operadores pendientes
	if err := ProcessExpressionEnd(ctx); err != nil {
		return err
	}

	// Obtener el resultado de la expresión
	result, ok := ctx.OperandStack.Pop()
	if !ok {
		return fmt.Errorf("error: no hay resultado de expresión para asignar")
	}

	resultType, _ := ctx.TypeStack.Pop()

	// Verificar tipo de variable
	varType, err := GetVariableTypeFromContext(ctx, varName)
	if err != nil {
		return err
	}

	// Validar asignación con cubo semántico
	_, err = ctx.Cube.Result(OpAssign, varType, resultType)
	if err != nil {
		return err
	}

	// Obtener dirección virtual de la variable
	varAddress, err := GetVariableAddressFromContext(ctx, varName)
	if err != nil {
		return err
	}

	// Generar cuádruplo de asignación (usar dirección virtual)
	generateQuadruple(ctx, "=", result, "", AddressToString(varAddress))

	return nil
}

// ProcessRelationalOperator procesa un operador relacional
func ProcessRelationalOperator(ctx *Context, op string) error {
	// Procesar la expresión izquierda primero
	if err := ProcessExpressionEnd(ctx); err != nil {
		return err
	}

	// Apilar el operador relacional
	ctx.OpStack.Push(op)

	return nil
}

// ProcessRelationalExpression procesa el final de una expresión relacional
func ProcessRelationalExpression(ctx *Context) error {
	// Procesar operadores aritméticos pendientes
	if err := ProcessExpressionEnd(ctx); err != nil {
		return err
	}

	// Debe haber un operador relacional en la pila
	relOp, ok := ctx.OpStack.Pop()
	if !ok {
		return fmt.Errorf("error: se esperaba operador relacional")
	}

	// Obtener operandos
	right, ok1 := ctx.OperandStack.Pop()
	rightType, _ := ctx.TypeStack.Pop()
	left, ok2 := ctx.OperandStack.Pop()
	leftType, _ := ctx.TypeStack.Pop()

	if !ok1 || !ok2 {
		return fmt.Errorf("error: operandos insuficientes para operador relacional %s", relOp)
	}

	// Validar con cubo semántico
	operator := Operator(relOp)
	resultType, err := ctx.Cube.Result(operator, leftType, rightType)
	if err != nil {
		return err
	}

	// Generar temporal para el resultado booleano (dirección virtual)
	temp := ctx.TempCounter.NextString()

	// Generar cuádruplo
	generateQuadruple(ctx, relOp, left, right, temp)

	// Apilar resultado
	PushOperand(ctx, temp, resultType)

	return nil
}

// ProcessPrint procesa una instrucción print
func ProcessPrint(ctx *Context, value string, isString bool) {
	if isString {
		// Para strings, generar cuádruplo de print directo
		generateQuadruple(ctx, "PRINT", value, "", "")
	} else {
		// Para expresiones, el valor ya está en la pila de operandos
		generateQuadruple(ctx, "PRINT", value, "", "")
	}
}

// ProcessIf procesa el inicio de un if
// Asume que la expresión condicional ya fue procesada y el resultado está en la pila
func ProcessIf(ctx *Context) (int, error) {
	// Obtener resultado de la condición (ya procesada)
	condition, ok := ctx.OperandStack.Pop()
	if !ok {
		return -1, fmt.Errorf("error: no hay condición para if")
	}

	ctx.TypeStack.Pop() // Remover tipo de la condición

	// Generar GOTOF (salto si falso)
	gotoIndex := ctx.Quadruples.NextIndex()
	generateQuadruple(ctx, "GOTOF", condition, "", "")

	// Guardar índice para completar después
	ctx.JumpStack.Push(gotoIndex)

	return gotoIndex, nil
}

// ProcessIfEnd completa el if sin else
func ProcessIfEnd(ctx *Context) error {
	// Completar el GOTOF con el índice actual
	jumpIndex, ok := ctx.JumpStack.Pop()
	if !ok {
		return fmt.Errorf("error: no hay salto pendiente para if")
	}

	// Actualizar el cuádruplo con el índice correcto (en Result)
	quad := ctx.Quadruples.GetAt(jumpIndex)
	if quad != nil {
		quad.Result = fmt.Sprintf("%d", ctx.Quadruples.NextIndex())
		ctx.Quadruples.UpdateAt(jumpIndex, *quad)
	}

	return nil
}

// ProcessElse procesa el else
func ProcessElse(ctx *Context) (int, error) {
	// Completar el GOTOF del if con el inicio del else
	jumpIndex, ok := ctx.JumpStack.Pop()
	if !ok {
		return -1, fmt.Errorf("error: no hay salto pendiente para else")
	}

	// Generar GOTO incondicional para saltar el else
	gotoIndex := ctx.Quadruples.NextIndex()
	generateQuadruple(ctx, "GOTO", "", "", "")

	// Actualizar el GOTOF (en Result)
	quad := ctx.Quadruples.GetAt(jumpIndex)
	if quad != nil {
		quad.Result = fmt.Sprintf("%d", ctx.Quadruples.NextIndex())
		ctx.Quadruples.UpdateAt(jumpIndex, *quad)
	}

	// Guardar el GOTO para completar después del else
	ctx.JumpStack.Push(gotoIndex)

	return gotoIndex, nil
}

// ProcessIfElseEnd completa el if-else
func ProcessIfElseEnd(ctx *Context) error {
	// Completar el GOTO del else
	jumpIndex, ok := ctx.JumpStack.Pop()
	if !ok {
		return fmt.Errorf("error: no hay salto pendiente para else")
	}

	// Actualizar el cuádruplo con el índice correcto (en Result)
	quad := ctx.Quadruples.GetAt(jumpIndex)
	if quad != nil {
		quad.Result = fmt.Sprintf("%d", ctx.Quadruples.NextIndex())
		ctx.Quadruples.UpdateAt(jumpIndex, *quad)
	}

	return nil
}

// ProcessWhileStart procesa el inicio de un while
func ProcessWhileStart(ctx *Context) int {
	// Guardar el índice de inicio del ciclo
	startIndex := ctx.Quadruples.NextIndex()
	ctx.JumpStack.Push(startIndex)
	return startIndex
}

// ProcessWhileCondition procesa la condición del while
// Asume que la expresión condicional ya fue procesada y el resultado está en la pila
// Guarda el índice donde comienza el ciclo (donde se evalúa la condición)
func ProcessWhileCondition(ctx *Context) error {
	// Obtener resultado de la condición (ya procesada)
	condition, ok := ctx.OperandStack.Pop()
	if !ok {
		return fmt.Errorf("error: no hay condición para while")
	}

	ctx.TypeStack.Pop() // Remover tipo de la condición

	// El índice de inicio del ciclo es donde se evalúa la condición
	// Como la condición ya fue evaluada, el último cuádruplo generado es el de la condición
	// (el que produce el resultado booleano). Ese es el índice donde comienza el ciclo.
	// Usamos Size() en lugar de NextIndex() porque NextIndex() ya incluye el espacio para el GOTOF
	quadSize := ctx.Quadruples.Size()
	startIndex := quadSize - 1
	if startIndex < 0 {
		startIndex = 0
	}

	// Guardar el índice de inicio primero (para que sea el último en la pila)
	ctx.JumpStack.Push(startIndex)

	// Generar GOTOF (salto si falso, salir del ciclo)
	gotoIndex := ctx.Quadruples.NextIndex()
	generateQuadruple(ctx, "GOTOF", condition, "", "")

	// Guardar índice del GOTOF para completar después (será el primero en la pila)
	ctx.JumpStack.Push(gotoIndex)

	return nil
}

// ProcessWhileEnd completa el while
func ProcessWhileEnd(ctx *Context) error {
	// Obtener el índice del GOTOF (último que se pusó)
	gotoIndex, ok := ctx.JumpStack.Pop()
	if !ok {
		return fmt.Errorf("error: no hay salto pendiente para while")
	}

	// Obtener el índice de inicio del ciclo (penúltimo que se pusó)
	startIndex, ok := ctx.JumpStack.Pop()
	if !ok {
		return fmt.Errorf("error: no hay índice de inicio para while")
	}

	// Generar GOTO al inicio del ciclo (donde se evalúa la condición)
	generateQuadruple(ctx, "GOTO", "", "", fmt.Sprintf("%d", startIndex))

	// Actualizar el GOTOF con el índice después del ciclo (en Result)
	quad := ctx.Quadruples.GetAt(gotoIndex)
	if quad != nil {
		quad.Result = fmt.Sprintf("%d", ctx.Quadruples.NextIndex())
		ctx.Quadruples.UpdateAt(gotoIndex, *quad)
	}

	return nil
}

// ProcessReturn processes a return statement with an expression (non-void functions)
func ProcessReturn(ctx *Context, exprValue string, exprType Type) error {
	var currentFn *FunctionEntry
	// Try to get from stack first
	currentFn = ctx.CurrentFunction()

	// If not in stack, try to look up by name
	if currentFn == nil && ctx.CurrentFunctionName != "" {
		var ok bool
		currentFn, ok = ctx.Directory.GetFunction(ctx.CurrentFunctionName)
		if !ok {
			ctx.PendingReturns = append(ctx.PendingReturns, PendingReturn{
				Value:    exprValue,
				Type:     exprType,
				Function: ctx.CurrentFunctionName,
			})
			generateQuadruple(ctx, "RETURN", exprValue, "", "")
			ctx.HasReturn = true
			return nil // Validamos despues
		}
	}

	// If still not found, try using PendingFunctionName as a workaround
	if currentFn == nil && ctx.PendingFunctionName != "" {
		var ok bool
		currentFn, ok = ctx.Directory.GetFunction(ctx.PendingFunctionName)
		if !ok {
			ctx.PendingReturns = append(ctx.PendingReturns, PendingReturn{
				Value:    exprValue,
				Type:     exprType,
				Function: ctx.CurrentFunctionName,
			})
			generateQuadruple(ctx, "RETURN", exprValue, "", "")
			ctx.HasReturn = true
			return nil // Validamos despues
		}
	}

	if currentFn == nil && ctx.Directory.LastAddedFunction != "" {
		var ok bool
		currentFn, ok = ctx.Directory.GetFunction(ctx.Directory.LastAddedFunction)
		if !ok {
			// Function not found yet - store for later validation
			ctx.PendingReturns = append(ctx.PendingReturns, PendingReturn{
				Value:    exprValue,
				Type:     exprType,
				Function: ctx.Directory.LastAddedFunction,
			})
			generateQuadruple(ctx, "RETURN", exprValue, "", "")
			ctx.HasReturn = true
			return nil
		}
	}

	// If we found the function, validate immediately
	if currentFn != nil {
		// Validate return type matches function return type
		if currentFn.ReturnType != exprType {
			return fmt.Errorf("error: tipo de retorno %s no coincide con tipo de función %s",
				exprType, currentFn.ReturnType)
		}
		// Mark that function has at least one return
		ctx.HasReturn = true
		// Generate RETURN quadruple
		generateQuadruple(ctx, "RETURN", exprValue, "", "")
		return nil
	}

	// If we still can't find the function, store for later
	// (This handles the case where we're in a function but can't identify it yet)
	ctx.PendingReturns = append(ctx.PendingReturns, PendingReturn{
		Value:    exprValue,
		Type:     exprType,
		Function: "", // Will be set in reduceFunction
	})
	generateQuadruple(ctx, "RETURN", exprValue, "", "")
	ctx.HasReturn = true
	return nil
}

// ProcessReturnVoid processes a return statement without expression (void functions)
func ProcessReturnVoid(ctx *Context) error {
	var currentFn *FunctionEntry
	// Try to get from stack first
	currentFn = ctx.CurrentFunction()

	// If not in stack, try to look up by name
	if currentFn == nil && ctx.CurrentFunctionName != "" {
		var ok bool
		currentFn, ok = ctx.Directory.GetFunction(ctx.CurrentFunctionName)
		if !ok {
			// Function not found yet - store for later validation
			ctx.PendingReturns = append(ctx.PendingReturns, PendingReturn{
				Value:    "",
				Type:     TypeVoid,
				Function: ctx.CurrentFunctionName,
			})
			// Still generate the return quadruple
			generateQuadruple(ctx, "RETURN", "", "", "")
			ctx.HasReturn = true // Mark that a return exists
			return nil           // Don't fail - we'll validate later
		}
	}

	// If still not found, try using PendingFunctionName
	if currentFn == nil && ctx.PendingFunctionName != "" {
		var ok bool
		currentFn, ok = ctx.Directory.GetFunction(ctx.PendingFunctionName)
		if !ok {
			// Function not found yet - store for later validation
			ctx.PendingReturns = append(ctx.PendingReturns, PendingReturn{
				Value:    "",
				Type:     TypeVoid,
				Function: ctx.PendingFunctionName,
			})
			generateQuadruple(ctx, "RETURN", "", "", "")
			ctx.HasReturn = true
			return nil
		}
	}

	// If still not found, try LastAddedFunction
	if currentFn == nil && ctx.Directory.LastAddedFunction != "" {
		var ok bool
		currentFn, ok = ctx.Directory.GetFunction(ctx.Directory.LastAddedFunction)
		if !ok {
			// Function not found yet - store for later validation
			ctx.PendingReturns = append(ctx.PendingReturns, PendingReturn{
				Value:    "",
				Type:     TypeVoid,
				Function: ctx.Directory.LastAddedFunction,
			})
			generateQuadruple(ctx, "RETURN", "", "", "")
			ctx.HasReturn = true
			return nil
		}
	}

	// If we found the function, validate immediately
	if currentFn != nil {
		// Validate return type matches function return type
		if currentFn.ReturnType != TypeVoid {
			return fmt.Errorf("error: tipo de retorno %s no coincide con tipo de función %s",
				TypeVoid, currentFn.ReturnType)
		}
		// Mark that function has at least one return
		ctx.HasReturn = true
		// Generate RETURN quadruple
		generateQuadruple(ctx, "RETURN", "", "", "")
		return nil
	}

	// If we still can't find the function, store for later
	// (This handles the case where we're in a function but can't identify it yet)
	ctx.PendingReturns = append(ctx.PendingReturns, PendingReturn{
		Value:    "",
		Type:     TypeVoid,
		Function: "", // Will be set in reduceFunction
	})
	generateQuadruple(ctx, "RETURN", "", "", "")
	ctx.HasReturn = true
	return nil
}
