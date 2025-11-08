package semantic

import "fmt"

// Operator representa las operaciones binarias del lenguaje que necesitan
// validación semántica (aritméticas, relacionales y asignaciones).
type Operator string

const (
	OpAdd      Operator = "+"
	OpSub      Operator = "-"
	OpMul      Operator = "*"
	OpDiv      Operator = "/"
	OpGt       Operator = ">"
	OpLt       Operator = "<"
	OpNeq      Operator = "!="
	OpEq       Operator = "=="
	OpAssign   Operator = "="
	OpUnaryPos Operator = "u+"
	OpUnaryNeg Operator = "u-"
)

// SemanticCube codifica la tabla de compatibilidades (cubo semántico).
type SemanticCube struct {
	table map[Operator]map[Type]map[Type]Type
}

// NewSemanticCube crea un cubo vacío listo para ser llenado.
func NewSemanticCube() *SemanticCube {
	return &SemanticCube{table: make(map[Operator]map[Type]map[Type]Type)}
}

// Set registra el tipo resultante de aplicar `op` sobre (left, right).
func (c *SemanticCube) Set(op Operator, left, right, result Type) {
	if _, ok := c.table[op]; !ok {
		c.table[op] = make(map[Type]map[Type]Type)
	}
	if _, ok := c.table[op][left]; !ok {
		c.table[op][left] = make(map[Type]Type)
	}
	c.table[op][left][right] = result
}

// Result devuelve el tipo resultante de op(left, right) o un error si es inválido.
func (c *SemanticCube) Result(op Operator, left, right Type) (Type, error) {
	if opTable, ok := c.table[op]; ok {
		if rightTable, ok := opTable[left]; ok {
			if result, ok := rightTable[right]; ok {
				return result, nil
			}
		}
	}
	return TypeInvalid, fmt.Errorf("operación inválida: %s (%s, %s)", op, left, right)
}

// ResultUnary evalúa operaciones unarias (+x, -x). Internamente modelamos la
// operación como si el segundo operando fuera TypeInvalid.
func (c *SemanticCube) ResultUnary(op Operator, operand Type) (Type, error) {
	return c.Result(op, operand, TypeInvalid)
}

// DefaultSemanticCube expone el cubo configurado para la gramática actual de Patito.
var DefaultSemanticCube = func() *SemanticCube {
	cube := NewSemanticCube()

	// --- Aritmética binaria ---
	arithOps := []Operator{OpAdd, OpSub, OpMul, OpDiv}
	for _, op := range arithOps {
		cube.Set(op, TypeInt, TypeInt, TypeInt)
		cube.Set(op, TypeInt, TypeFloat, TypeFloat)
		cube.Set(op, TypeFloat, TypeInt, TypeFloat)
		cube.Set(op, TypeFloat, TypeFloat, TypeFloat)
	}

	// --- Aritmética unaria ---
	cube.Set(OpUnaryPos, TypeInt, TypeInvalid, TypeInt)
	cube.Set(OpUnaryPos, TypeFloat, TypeInvalid, TypeFloat)
	cube.Set(OpUnaryNeg, TypeInt, TypeInvalid, TypeInt)
	cube.Set(OpUnaryNeg, TypeFloat, TypeInvalid, TypeFloat)

	// --- Relacionales (> < != ==) ---
	relOps := []Operator{OpGt, OpLt, OpNeq, OpEq}
	for _, op := range relOps {
		cube.Set(op, TypeInt, TypeInt, TypeBool)
		cube.Set(op, TypeFloat, TypeFloat, TypeBool)
		cube.Set(op, TypeInt, TypeFloat, TypeBool)
		cube.Set(op, TypeFloat, TypeInt, TypeBool)
	}

	// --- Asignaciones ---
	// Se permite asignar int->int, float->float, int->float (promoción) y bool->bool.
	cube.Set(OpAssign, TypeInt, TypeInt, TypeInt)
	cube.Set(OpAssign, TypeFloat, TypeFloat, TypeFloat)
	cube.Set(OpAssign, TypeFloat, TypeInt, TypeFloat) // promoción
	cube.Set(OpAssign, TypeBool, TypeBool, TypeBool)

	return cube
}()
