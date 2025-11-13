package semantic

import (
	"fmt"
	"Patito/token"
)

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
			
			// Generar temporal
			temp := ctx.TempCounter.Next()
			
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
	
	// Generar temporal
	temp := ctx.TempCounter.Next()
	
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
func PushConstant(ctx *Context, tok *token.Token) error {
	var operandType Type
	var operand string
	
	switch tok.Type {
	case token.TokMap.Type("cte_int"):
		operandType = TypeInt
		operand = string(tok.Lit)
	case token.TokMap.Type("cte_float"):
		operandType = TypeFloat
		operand = string(tok.Lit)
	default:
		return fmt.Errorf("tipo de constante no soportado: %s", tok.Type)
	}
	
	PushOperand(ctx, operand, operandType)
	return nil
}

// PushVariable apila una variable y busca su tipo en el directorio
func PushVariable(ctx *Context, varName string, pos token.Pos) error {
	// Buscar variable primero en contexto, luego en directorio
	varType, err := GetVariableTypeFromContext(ctx, varName)
	if err != nil {
		return err
	}
	
	PushOperand(ctx, varName, varType)
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
		
		// Generar temporal
		temp := ctx.TempCounter.Next()
		
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
	
	// Generar cuádruplo de asignación
	generateQuadruple(ctx, "=", result, "", varName)
	
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
	
	// Generar temporal para el resultado booleano
	temp := ctx.TempCounter.Next()
	
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
// Retorna el índice donde comienza el ciclo (donde está la condición)
// Nota: La condición ya fue evaluada. El índice de inicio es donde está la condición,
// que es el último cuádruplo generado antes de este punto (antes del cuerpo del while)
func ProcessWhileCondition(ctx *Context) (int, error) {
	// Obtener resultado de la condición (ya procesada)
	condition, ok := ctx.OperandStack.Pop()
	if !ok {
		return -1, fmt.Errorf("error: no hay condición para while")
	}
	
	ctx.TypeStack.Pop() // Remover tipo de la condición
	
	// El índice de inicio del ciclo es donde está la condición
	// Como la condición ya fue evaluada antes del cuerpo, necesitamos encontrar
	// el índice del último cuádruplo antes del cuerpo.
	// Por simplicidad, asumimos que la condición es el penúltimo cuádruplo
	// (el último es algo del cuerpo, el penúltimo es la condición)
	// Pero esto no es confiable. Mejor: el índice de inicio es donde generaremos el GOTOF - 1
	// (asumiendo que la condición está justo antes)
	currentIndex := ctx.Quadruples.NextIndex()
	// El índice de inicio es el índice actual (donde se generará el GOTOF)
	// pero la condición está antes. Por ahora, usamos currentIndex - 1 como aproximación
	// Esto asume que la condición es el cuádruplo inmediatamente anterior
	startIndex := currentIndex - 1
	if startIndex < 0 {
		startIndex = 0
	}
	
	// Generar GOTOF (salto si falso, salir del ciclo)
	gotoIndex := ctx.Quadruples.NextIndex()
	generateQuadruple(ctx, "GOTOF", condition, "", "")
	
	// Guardar índice del GOTOF para completar después
	ctx.JumpStack.Push(gotoIndex)
	
	// Retornar el índice de inicio (aproximación: condición está en startIndex)
	return startIndex, nil
}

// ProcessWhileEnd completa el while
func ProcessWhileEnd(ctx *Context) error {
	// Obtener el índice del GOTOF
	gotoIndex, ok := ctx.JumpStack.Pop()
	if !ok {
		return fmt.Errorf("error: no hay salto pendiente para while")
	}
	
	// Obtener el índice de inicio del ciclo (donde está la condición)
	startIndex, ok := ctx.JumpStack.Pop()
	if !ok {
		return fmt.Errorf("error: no hay índice de inicio para while")
	}
	
	// Generar GOTO al inicio del ciclo (donde está la condición)
	generateQuadruple(ctx, "GOTO", "", "", fmt.Sprintf("%d", startIndex))
	
	// Actualizar el GOTOF con el índice después del ciclo (en Result)
	quad := ctx.Quadruples.GetAt(gotoIndex)
	if quad != nil {
		quad.Result = fmt.Sprintf("%d", ctx.Quadruples.NextIndex())
		ctx.Quadruples.UpdateAt(gotoIndex, *quad)
	}
	
	return nil
}

