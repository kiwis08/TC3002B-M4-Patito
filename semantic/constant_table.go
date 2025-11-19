package semantic

import "fmt"

// ConstantEntry representa una constante almacenada en la tabla de constantes
type ConstantEntry struct {
	Value   string // Valor de la constante (como string)
	Type    Type   // Tipo de la constante
	Address int    // Dirección virtual asignada
}

// ConstantTable almacena todas las constantes del programa con sus direcciones virtuales
type ConstantTable struct {
	constants map[string]*ConstantEntry // Mapa: valor -> entrada
	order     []*ConstantEntry          // Orden de inserción
}

// NewConstantTable crea una nueva tabla de constantes
func NewConstantTable() *ConstantTable {
	return &ConstantTable{
		constants: make(map[string]*ConstantEntry),
		order:     make([]*ConstantEntry, 0),
	}
}

// Add agrega una constante a la tabla. Si ya existe, retorna la entrada existente.
// Si no existe, crea una nueva entrada con la dirección virtual proporcionada.
func (ct *ConstantTable) Add(value string, constType Type, address int) *ConstantEntry {
	// Crear clave única basada en valor y tipo
	key := fmt.Sprintf("%s:%s", value, constType.String())

	// Si ya existe, retornar la existente
	if entry, ok := ct.constants[key]; ok {
		return entry
	}

	// Crear nueva entrada
	entry := &ConstantEntry{
		Value:   value,
		Type:    constType,
		Address: address,
	}

	ct.constants[key] = entry
	ct.order = append(ct.order, entry)
	return entry
}

// Get busca una constante por su valor y tipo
func (ct *ConstantTable) Get(value string, constType Type) (*ConstantEntry, bool) {
	key := fmt.Sprintf("%s:%s", value, constType.String())
	entry, ok := ct.constants[key]
	return entry, ok
}

// Entries devuelve todas las constantes en orden de inserción
func (ct *ConstantTable) Entries() []*ConstantEntry {
	result := make([]*ConstantEntry, len(ct.order))
	copy(result, ct.order)
	return result
}
