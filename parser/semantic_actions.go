package parser

import (
	"fmt"

	"Patito/semantic"
	"Patito/token"
)

type reduceFunc = func([]Attrib, interface{}) (Attrib, error)

func init() {
	setReduceFunc(1, reduceProgram)
	setReduceFunc(2, passThrough)
	setReduceFunc(3, returnEmptySpecs)
	setReduceFunc(4, takeSecond)
	setReduceFunc(5, concatSpecSlices)
	setReduceFunc(6, returnEmptySpecs)
	setReduceFunc(7, reduceVarDeclaration)
	setReduceFunc(8, prependIDToken)
	setReduceFunc(9, returnEmptyIDTokens)
	setReduceFunc(10, reduceTypeInt)
	setReduceFunc(11, reduceTypeFloat)
	setReduceFunc(14, passThrough)
	setReduceFunc(15, reduceTypeVoid)
	setReduceFunc(16, reduceFunction)
	setReduceFunc(17, takeSecond)
	setReduceFunc(18, returnEmptySpecs)
	setReduceFunc(19, reduceParamSequence)
	setReduceFunc(20, returnEmptySpecs)
	setReduceFunc(21, reduceParamTail)
	setReduceFunc(22, returnEmptySpecs)
	setReduceFunc(23, passThrough)
	setReduceFunc(24, returnEmptySpecs)
	setReduceFunc(25, reduceParam)
	// Expresiones y estatutos
	setReduceFunc(37, passThrough)            // PRINT_P : E_PRINT R_PRINT
	setReduceFunc(38, reduceEPrintExpression) // E_PRINT : EXPRESSION
	setReduceFunc(39, reduceEPrintString)     // E_PRINT : cte_string
	setReduceFunc(40, reduceRPrint)           // R_PRINT : "," PRINT_P
	setReduceFunc(41, returnEmptySpecs)       // R_PRINT : empty
	setReduceFunc(42, reduceAssign)           // ASSIGN : id "=" EXPRESSION ";"
	setReduceFunc(43, reduceCycle)            // CYCLE : "while" "(" EXPRESSION ")" "do" BODY ";"
	setReduceFunc(44, reduceCondition)        // CONDITION : "if" "(" EXPRESSION ")" BODY ";"
	setReduceFunc(45, reduceConditionElse)    // CONDITION : "if" "(" EXPRESSION ")" BODY "else" BODY ";"
	setReduceFunc(46, reduceReturn)           // RETURN : "return" EXPRESSION ";"
	setReduceFunc(47, reduceReturnVoid)       // RETURN : "return" ";"
	setReduceFunc(48, reduceExpression)       // EXPRESSION : EXP REL_TAIL
	setReduceFunc(49, reduceRelTail)          // REL_TAIL : REL_OP EXP
	setReduceFunc(50, returnEmptySpecs)       // REL_TAIL : empty
	setReduceFunc(51, reduceRelOpGt)          // REL_OP : ">"
	setReduceFunc(52, reduceRelOpLt)          // REL_OP : "<"
	setReduceFunc(53, reduceRelOpNeq)         // REL_OP : "!="
	setReduceFunc(54, reduceRelOpEq)          // REL_OP : "=="
	setReduceFunc(55, reduceExp)              // EXP : TERMINO EXP_P
	setReduceFunc(56, reduceExpPAdd)          // EXP_P : "+" TERMINO EXP_P
	setReduceFunc(57, reduceExpPSub)          // EXP_P : "-" TERMINO EXP_P
	setReduceFunc(58, returnEmptySpecs)       // EXP_P : empty
	setReduceFunc(59, reduceTermino)          // TERMINO : FACTOR TERMINO_P
	setReduceFunc(60, reduceTerminoPMul)      // TERMINO_P : "*" FACTOR TERMINO_P
	setReduceFunc(61, reduceTerminoPDiv)      // TERMINO_P : "/" FACTOR TERMINO_P
	setReduceFunc(62, returnEmptySpecs)       // TERMINO_P : empty
	setReduceFunc(63, reduceFactor)           // FACTOR : S_OP FACTOR_CORE
	setReduceFunc(64, reduceFactorCoreParen)  // FACTOR_CORE : "(" EXPRESSION ")"
	setReduceFunc(65, reduceFactorCoreId)     // FACTOR_CORE : id FACTOR_SUFFIX
	setReduceFunc(66, reduceFactorCoreCte)    // FACTOR_CORE : CTE
	setReduceFunc(67, reduceFactorSuffixCall)
	setReduceFunc(68, returnEmptySpecs)
}

func setReduceFunc(index int, fn reduceFunc) {
	for i := range productionsTable {
		if productionsTable[i].Index == index {
			productionsTable[i].ReduceFunc = fn
			return
		}
	}
	panic(fmt.Sprintf("no se encontró producción con index %d", index))
}

func semanticCtx(C interface{}) (*semantic.Context, error) {
	ctx, ok := C.(*semantic.Context)
	if !ok || ctx == nil {
		return nil, fmt.Errorf("parser.Context no inicializado con *semantic.Context")
	}
	return ctx, nil
}

func tokenFromAttrib(a Attrib) (*token.Token, error) {
	tok, ok := a.(*token.Token)
	if !ok || tok == nil {
		return nil, fmt.Errorf("se esperaba token, obtuvo %T", a)
	}
	return tok, nil
}

func specsFromAttrib(a Attrib) []*semantic.VariableSpec {
	if a == nil {
		return nil
	}
	switch v := a.(type) {
	case []*semantic.VariableSpec:
		return v
	default:
		return nil
	}
}

func idTokensFromAttrib(a Attrib) []*token.Token {
	if a == nil {
		return nil
	}
	switch v := a.(type) {
	case []*token.Token:
		return v
	default:
		return nil
	}
}

// --- Reduce functions ---

func reduceProgram(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	programID, err := tokenFromAttrib(X[1])
	if err != nil {
		return nil, err
	}
	if err := ctx.Directory.SetProgram(programID.IDValue(), programID.Pos); err != nil {
		return nil, err
	}
	if globals := specsFromAttrib(X[3]); len(globals) > 0 {
		if err := ctx.Directory.AddGlobals(globals, ctx.AddressManager); err != nil {
			return nil, err
		}
		// Almacenar direcciones en VariableAddresses para uso inmediato
		for _, spec := range globals {
			ctx.VariableAddresses[spec.Name] = spec.Address
		}
	}
	return ctx.Directory, nil
}

func passThrough(X []Attrib, _ interface{}) (Attrib, error) {
	if len(X) == 0 {
		return nil, nil
	}
	return X[0], nil
}

func takeSecond(X []Attrib, _ interface{}) (Attrib, error) {
	if len(X) < 2 {
		return nil, nil
	}
	return X[1], nil
}

func returnEmptySpecs(_ []Attrib, _ interface{}) (Attrib, error) {
	return []*semantic.VariableSpec{}, nil
}

func returnEmptyIDTokens(_ []Attrib, _ interface{}) (Attrib, error) {
	return []*token.Token{}, nil
}

func concatSpecSlices(X []Attrib, _ interface{}) (Attrib, error) {
	left := specsFromAttrib(X[0])
	right := specsFromAttrib(X[1])
	if len(left) == 0 {
		return right, nil
	}
	if len(right) == 0 {
		return left, nil
	}
	out := make([]*semantic.VariableSpec, 0, len(left)+len(right))
	out = append(out, left...)
	out = append(out, right...)
	return out, nil
}

func reduceVarDeclaration(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	firstID, err := tokenFromAttrib(X[0])
	if err != nil {
		return nil, err
	}
	otherIDs := idTokensFromAttrib(X[1])
	typeVal, ok := X[3].(semantic.Type)
	if !ok {
		return nil, fmt.Errorf("esperaba semantic.Type, obtuvo %T", X[3])
	}

	idTokens := append([]*token.Token{firstID}, otherIDs...)
	specs := make([]*semantic.VariableSpec, 0, len(idTokens))
	for _, tok := range idTokens {
		varName := tok.IDValue()
		// Almacenar tipo en contexto para uso inmediato durante parsing
		ctx.VariableTypes[varName] = typeVal
		// Asignar dirección virtual inmediatamente para que esté disponible durante parsing
		// Esto es necesario porque las variables pueden usarse en el main body antes de que
		// reduceProgram agregue las variables al directorio
		addr := ctx.AddressManager.NextGlobal()
		ctx.VariableAddresses[varName] = addr
		specs = append(specs, &semantic.VariableSpec{
			Name:    varName,
			Type:    typeVal,
			Pos:     tok.Pos,
			Address: addr,
		})
	}
	return specs, nil
}

func prependIDToken(X []Attrib, _ interface{}) (Attrib, error) {
	idTok, err := tokenFromAttrib(X[1])
	if err != nil {
		return nil, err
	}
	rest := idTokensFromAttrib(X[2])
	out := make([]*token.Token, 0, 1+len(rest))
	out = append(out, idTok)
	out = append(out, rest...)
	return out, nil
}

func reduceTypeInt(_ []Attrib, _ interface{}) (Attrib, error) {
	return semantic.TypeInt, nil
}

func reduceTypeFloat(_ []Attrib, _ interface{}) (Attrib, error) {
	return semantic.TypeFloat, nil
}

func reduceTypeVoid(_ []Attrib, _ interface{}) (Attrib, error) {
	return semantic.TypeVoid, nil
}

func reduceFunction(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	returnType, ok := X[0].(semantic.Type)
	if !ok {
		return nil, fmt.Errorf("esperaba semantic.Type para retorno, obtuvo %T", X[0])
	}
	fnID, err := tokenFromAttrib(X[1])
	if err != nil {
		return nil, err
	}
	params := specsFromAttrib(X[3])
	locals := specsFromAttrib(X[5])

	fnEntry, err := ctx.Directory.AddFunction(fnID.IDValue(), returnType, fnID.Pos, params, locals, ctx.AddressManager)
	if err != nil {
		return nil, err
	}

	// Validate any pending returns for this function
	fnName := fnID.IDValue()
	hasReturnForThisFunction := false
	for i := len(ctx.PendingReturns) - 1; i >= 0; i-- {
		pendingReturn := ctx.PendingReturns[i]
		// Check if this return belongs to this function
		// (If Function is empty or matches this function name)
		if pendingReturn.Function == "" || pendingReturn.Function == fnName {
			// Validate return type

			hasReturnForThisFunction = true
			if pendingReturn.Type != returnType {
				return nil, fmt.Errorf("%s: tipo de retorno %s no coincide con tipo de función %s",
					pendingReturn.Pos, pendingReturn.Type, returnType)
			}
			// Remove from pending list
			ctx.PendingReturns = append(ctx.PendingReturns[:i], ctx.PendingReturns[i+1:]...)
		}
	}
	// Store function name for return statements to use
	previousFunctionName := ctx.CurrentFunctionName
	previousPendingName := ctx.PendingFunctionName
	ctx.CurrentFunctionName = fnName
	ctx.PendingFunctionName = fnName
	ctx.PushFunction(fnEntry)
	ctx.HasReturn = false

	// Almacenar direcciones de parámetros y locales en VariableAddresses para uso inmediato
	for _, spec := range params {
		ctx.VariableAddresses[spec.Name] = spec.Address
	}
	for _, spec := range locals {
		ctx.VariableAddresses[spec.Name] = spec.Address
	}

	// Validate that function has at least one return statement
	if !hasReturnForThisFunction {
		return nil, fmt.Errorf("error: función %s debe tener al menos un return statement", fnID.IDValue())
	}

	// Restore previous function name and pop function from stack
	ctx.CurrentFunctionName = previousFunctionName
	ctx.PendingFunctionName = previousPendingName
	ctx.PopFunction()

	return nil, nil
}

func reduceParamSequence(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	first, ok := X[0].(*semantic.VariableSpec)
	if !ok || first == nil {
		return nil, fmt.Errorf("esperaba VariableSpec, obtuvo %T", X[0])
	}
	rest := specsFromAttrib(X[1])
	out := make([]*semantic.VariableSpec, 0, 1+len(rest))
	out = append(out, first)
	out = append(out, rest...)

	// Add parameters to context so they are available when body is parsed
	for _, spec := range out {
		ctx.VariableTypes[spec.Name] = spec.Type
		// Assign virtual addresses to parameters (this address will be updated in reduceFunction)
		tempAddr := ctx.AddressManager.NextLocal()
		ctx.VariableAddresses[spec.Name] = tempAddr
		spec.Address = tempAddr
	}

	return out, nil
}

func reduceParamTail(X []Attrib, _ interface{}) (Attrib, error) {
	first, ok := X[1].(*semantic.VariableSpec)
	if !ok || first == nil {
		return nil, fmt.Errorf("esperaba VariableSpec en cola de parámetros, obtuvo %T", X[1])
	}
	rest := specsFromAttrib(X[2])
	out := make([]*semantic.VariableSpec, 0, 1+len(rest))
	out = append(out, first)
	out = append(out, rest...)
	return out, nil
}

func reduceParam(X []Attrib, _ interface{}) (Attrib, error) {
	idTok, err := tokenFromAttrib(X[0])
	if err != nil {
		return nil, err
	}
	typ, ok := X[2].(semantic.Type)
	if !ok {
		return nil, fmt.Errorf("esperaba semantic.Type en parámetro, obtuvo %T", X[2])
	}
	return &semantic.VariableSpec{
		Name: idTok.IDValue(),
		Type: typ,
		Pos:  idTok.Pos,
	}, nil
}

// --- Funciones de reducción para generación de cuádruplos ---

// reduceFactorCoreCte: CTE -> cte_int | cte_float
func reduceFactorCoreCte(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	tok, err := tokenFromAttrib(X[0])
	if err != nil {
		return nil, err
	}
	if err := semantic.PushConstant(ctx, tok); err != nil {
		return nil, err
	}
	return tok, nil
}

// reduceFactorCoreId: FACTOR_CORE -> id FACTOR_SUFFIX
func reduceFactorCoreId(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	idTok, err := tokenFromAttrib(X[0])
	if err != nil {
		return nil, err
	}

	// Check if FACTOR_SUFFIX is a function call (not empty)
	// If X[1] is not nil, it means FACTOR_SUFFIX was "(" S_E ")" (a function call)
	// If X[1] is nil, it means FACTOR_SUFFIX was empty (a variable)

	if X[1] != nil {
		if X[1] == "FUNC_CALL" {
			return processFunctionCall(ctx, idTok, nil)
		}
	}

	if err := semantic.PushVariable(ctx, idTok.IDValue(), idTok.Pos); err != nil {
		return nil, err
	}
	return idTok, nil
}

func processFunctionCall(ctx *semantic.Context, fnID *token.Token, argsAttrib Attrib) (Attrib, error) {
	fnName := fnID.IDValue()

	//Get the function from the directory
	fnEntry, ok := ctx.Directory.GetFunction(fnName)
	if !ok {
		return nil, fmt.Errorf("%s: función '%s' no declarada", fnID.Pos, fnName)
	}

	expectedParamCount := len(fnEntry.Params.Entries())

	// Get arguments from operand stack
	argValues := make([]string, 0, expectedParamCount)
	argTypes := make([]semantic.Type, 0, expectedParamCount)

	// Pop arguments from stack

	for i := 0; i < expectedParamCount; i++ {
		if ctx.OperandStack.IsEmpty() {
			return nil, fmt.Errorf("%s: función '%s' esperaba %d argumentos, pero se proporcionaron menos", fnID.Pos, fnName, expectedParamCount)
		}
		argValue, _ := ctx.OperandStack.Pop()
		argType, _ := ctx.TypeStack.Pop()
		argValues = append([]string{argValue}, argValues...)     // Prepend to maintain the order
		argTypes = append([]semantic.Type{argType}, argTypes...) // Prepend to maintain the order
	}

	// Validate the argument count
	if len(argValues) != expectedParamCount {
		return nil, fmt.Errorf("%s: función '%s' esperaba %d argumentos, pero se proporcionaron %d",
			fnID.Pos, fnName, expectedParamCount, len(argValues))
	}

	// Validate argument types
	params := fnEntry.Params.Entries()
	for i, param := range params {
		if argTypes[i] != param.Type {
			return nil, fmt.Errorf("%s: tipo de argumento %d en llama a '%s': esperaba %s, obtuvo %s", fnID.Pos, i+1, fnName, param.Type, argTypes[i])
		}
	}

	semantic.GenerateQuadruple(ctx, "ERA", fnName, "", "")

	// Generate PARAM quadruples for each argument
	for _, argValue := range argValues {
		semantic.GenerateQuadruple(ctx, "PARAM", argValue, "", "")
	}

	semantic.GenerateQuadruple(ctx, "GOSUB", fnName, "", "")

	if fnEntry.ReturnType != semantic.TypeVoid {
		temp := ctx.TempCounter.NextString()

		// Push the return value location onto the operand stack
		semantic.PushOperand(ctx, temp, fnEntry.ReturnType)
	}

	return fnID, nil
}

// reduceFactorCoreParen: FACTOR_CORE -> "(" EXPRESSION ")"
func reduceFactorCoreParen(X []Attrib, C interface{}) (Attrib, error) {
	// La expresión ya procesó todo, solo pasamos
	return X[1], nil
}

// reduceFactor: FACTOR -> S_OP FACTOR_CORE
func reduceFactor(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	// Si hay operador unario (S_OP)
	if len(X) >= 2 && X[0] != nil {
		if opTok, ok := X[0].(*token.Token); ok {
			op := string(opTok.Lit)
			if op == "+" {
				// + unario no hace nada, solo retornar el factor
				return X[1], nil
			} else if op == "-" {
				// - unario: procesar después de que FACTOR_CORE haya apilado el operando
				// El operando ya está en la pila, solo aplicar el operador unario
				if err := semantic.ProcessUnaryOperator(ctx, "u-"); err != nil {
					return nil, err
				}
			}
		}
		// Si X[0] es nil, S_OP era empty, no hay operador unario
	}
	return X[1], nil
}

// reduceTermino: TERMINO -> FACTOR TERMINO_P
func reduceTermino(X []Attrib, C interface{}) (Attrib, error) {
	// Solo pasamos, TERMINO_P ya procesó los operadores
	return X[0], nil
}

// reduceTerminoPMul: TERMINO_P -> "*" FACTOR TERMINO_P
func reduceTerminoPMul(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	if err := semantic.ProcessOperator(ctx, "*"); err != nil {
		return nil, err
	}
	return X[2], nil
}

// reduceTerminoPDiv: TERMINO_P -> "/" FACTOR TERMINO_P
func reduceTerminoPDiv(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	if err := semantic.ProcessOperator(ctx, "/"); err != nil {
		return nil, err
	}
	return X[2], nil
}

// reduceExp: EXP -> TERMINO EXP_P
func reduceExp(X []Attrib, C interface{}) (Attrib, error) {
	// Solo pasamos, EXP_P ya procesó los operadores
	return X[0], nil
}

// reduceExpPAdd: EXP_P -> "+" TERMINO EXP_P
func reduceExpPAdd(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	if err := semantic.ProcessOperator(ctx, "+"); err != nil {
		return nil, err
	}
	return X[2], nil
}

// reduceExpPSub: EXP_P -> "-" TERMINO EXP_P
func reduceExpPSub(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	if err := semantic.ProcessOperator(ctx, "-"); err != nil {
		return nil, err
	}
	return X[2], nil
}

// reduceRelOpGt: REL_OP -> ">"
func reduceRelOpGt(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	ctx.OpStack.Push(">")
	return X[0], nil
}

// reduceRelOpLt: REL_OP -> "<"
func reduceRelOpLt(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	ctx.OpStack.Push("<")
	return X[0], nil
}

// reduceRelOpNeq: REL_OP -> "!="
func reduceRelOpNeq(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	ctx.OpStack.Push("!=")
	return X[0], nil
}

// reduceRelOpEq: REL_OP -> "=="
func reduceRelOpEq(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	ctx.OpStack.Push("==")
	return X[0], nil
}

// reduceRelTail: REL_TAIL -> REL_OP EXP
func reduceRelTail(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	// Procesar la expresión relacional completa
	if err := semantic.ProcessRelationalExpression(ctx); err != nil {
		return nil, err
	}
	return X[1], nil
}

// reduceExpression: EXPRESSION -> EXP REL_TAIL
func reduceExpression(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	// Si no hay REL_TAIL, solo procesar EXP
	if X[1] == nil {
		if err := semantic.ProcessExpressionEnd(ctx); err != nil {
			return nil, err
		}
	}
	// Si hay REL_TAIL, ya se procesó en reduceRelTail
	return X[0], nil
}

// reduceAssign: ASSIGN -> id "=" EXPRESSION ";"
func reduceAssign(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	idTok, err := tokenFromAttrib(X[0])
	if err != nil {
		return nil, err
	}
	if err := semantic.ProcessAssignment(ctx, idTok.IDValue(), idTok.Pos); err != nil {
		return nil, err
	}
	return nil, nil
}

// reduceEPrintExpression: E_PRINT -> EXPRESSION
func reduceEPrintExpression(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	// Procesar expresión si hay operadores pendientes
	if err := semantic.ProcessExpressionEnd(ctx); err != nil {
		return nil, err
	}
	// Obtener resultado
	result, ok := ctx.OperandStack.Pop()
	if !ok {
		return nil, fmt.Errorf("error: no hay expresión para print")
	}
	ctx.TypeStack.Pop() // Remover tipo
	// Generar cuádruplo de print
	semantic.ProcessPrint(ctx, result, false)
	return X[0], nil
}

// reduceEPrintString: E_PRINT -> cte_string
func reduceEPrintString(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	tok, err := tokenFromAttrib(X[0])
	if err != nil {
		return nil, err
	}
	// Generar cuádruplo de print para string
	semantic.ProcessPrint(ctx, tok.StringValue(), true)
	return X[0], nil
}

// reduceRPrint: R_PRINT -> "," PRINT_P
func reduceRPrint(X []Attrib, C interface{}) (Attrib, error) {
	// Solo pasamos, PRINT_P ya procesó
	return X[1], nil
}

// reduceCondition: CONDITION -> "if" "(" EXPRESSION ")" BODY ";"
func reduceCondition(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	// La EXPRESSION (X[2]) ya fue procesada, ahora procesamos el if
	if _, err := semantic.ProcessIf(ctx); err != nil {
		return nil, err
	}
	// Al final del BODY, completar el if
	if err := semantic.ProcessIfEnd(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

// reduceConditionElse: CONDITION -> "if" "(" EXPRESSION ")" BODY "else" BODY ";"
func reduceConditionElse(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	// La EXPRESSION (X[2]) ya fue procesada, ahora procesamos el if
	if _, err := semantic.ProcessIf(ctx); err != nil {
		return nil, err
	}
	// Después del primer BODY, procesar else
	if _, err := semantic.ProcessElse(ctx); err != nil {
		return nil, err
	}
	// Al final del segundo BODY, completar el if-else
	if err := semantic.ProcessIfElseEnd(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

// reduceCycle: CYCLE -> "while" "(" EXPRESSION ")" "do" BODY ";"
func reduceCycle(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}
	// La EXPRESSION (X[2]) ya fue procesada
	// Procesar la condición: esto genera el GOTOF y guarda los índices necesarios
	if err := semantic.ProcessWhileCondition(ctx); err != nil {
		return nil, err
	}
	// Al final del BODY, completar el while
	if err := semantic.ProcessWhileEnd(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

// reduceReturn: RETURN -> "return" EXPRESSION ";"
func reduceReturn(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}

	// The EXPRESSION (X[1]) has already been processed, so we need to
	// process any pending operators and get the result
	if err := semantic.ProcessExpressionEnd(ctx); err != nil {
		return nil, err
	}

	// Get the expression result and type from the stacks
	exprValue, ok := ctx.OperandStack.Pop()
	if !ok {
		return nil, fmt.Errorf("error: no hay expresión para return")
	}
	exprType, ok := ctx.TypeStack.Pop()
	if !ok {
		return nil, fmt.Errorf("error: no hay tipo para expresión de return")
	}

	// Process the return statement
	if err := semantic.ProcessReturn(ctx, exprValue, exprType); err != nil {
		return nil, err
	}

	return nil, nil
}

// reduceReturnVoid: RETURN -> "return" ";"
func reduceReturnVoid(X []Attrib, C interface{}) (Attrib, error) {
	ctx, err := semanticCtx(C)
	if err != nil {
		return nil, err
	}

	// Process void return statement
	if err := semantic.ProcessReturnVoid(ctx); err != nil {
		return nil, err
	}

	return nil, nil
}

func reduceFactorSuffixCall(X []Attrib, C interface{}) (Attrib, error) {
	return "FUNC_CALL", nil
}
