package semantic

import "fmt"

// Type encodes the semantic type system supported by Patito at this stage.
// The values match the classic Patito types introduced en la fase semántica:
// enteros, flotantes, void (para funciones), booleanos (resultado de relacionales)
// y strings (para literales impresos).
type Type int

const (
	TypeInvalid Type = iota
	TypeInt
	TypeFloat
	TypeVoid
	TypeBool
	TypeString
)

func (t Type) String() string {
	switch t {
	case TypeInt:
		return "int"
	case TypeFloat:
		return "float"
	case TypeVoid:
		return "void"
	case TypeBool:
		return "bool"
	case TypeString:
		return "string"
	default:
		return "invalid"
	}
}

// TypeFromKeyword maps un literal de la gramática ("int", "float", "void")
// a nuestro enumerado interno. Retorna un error si recibe un literal desconocido.
func TypeFromKeyword(keyword string) (Type, error) {
	switch keyword {
	case "int":
		return TypeInt, nil
	case "float":
		return TypeFloat, nil
	case "void":
		return TypeVoid, nil
	default:
		return TypeInvalid, fmt.Errorf("tipo semántico desconocido: %q", keyword)
	}
}
