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
		if err := ctx.Directory.AddGlobals(globals); err != nil {
			return nil, err
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

func reduceVarDeclaration(X []Attrib, _ interface{}) (Attrib, error) {
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
		specs = append(specs, &semantic.VariableSpec{
			Name: tok.IDValue(),
			Type: typeVal,
			Pos:  tok.Pos,
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

	if _, err := ctx.Directory.AddFunction(fnID.IDValue(), returnType, fnID.Pos, params, locals); err != nil {
		return nil, err
	}
	return nil, nil
}

func reduceParamSequence(X []Attrib, _ interface{}) (Attrib, error) {
	first, ok := X[0].(*semantic.VariableSpec)
	if !ok || first == nil {
		return nil, fmt.Errorf("esperaba VariableSpec, obtuvo %T", X[0])
	}
	rest := specsFromAttrib(X[1])
	out := make([]*semantic.VariableSpec, 0, 1+len(rest))
	out = append(out, first)
	out = append(out, rest...)
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
