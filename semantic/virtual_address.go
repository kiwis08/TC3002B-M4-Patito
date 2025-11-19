package semantic

import "fmt"

// VirtualAddressManager gestiona la asignación de direcciones virtuales
// según los rangos estándar de Patito
type VirtualAddressManager struct {
	// Contadores para cada tipo de dirección
	globalCounter   int // 1000-9999
	localCounter    int // 10000-19999
	temporalCounter int // 20000-29999
	constantCounter int // 30000-39999

	// Rangos base
	GlobalBase   int
	LocalBase    int
	TemporalBase int
	ConstantBase int
}

// NewVirtualAddressManager crea un nuevo gestor de direcciones virtuales
func NewVirtualAddressManager() *VirtualAddressManager {
	return &VirtualAddressManager{
		GlobalBase:      1000,
		LocalBase:       10000,
		TemporalBase:    20000,
		ConstantBase:    30000,
		globalCounter:   1000,
		localCounter:    10000,
		temporalCounter: 20000,
		constantCounter: 30000,
	}
}

// NextGlobal asigna la siguiente dirección global
func (vam *VirtualAddressManager) NextGlobal() int {
	addr := vam.globalCounter
	vam.globalCounter++
	if vam.globalCounter > 9999 {
		panic("se excedió el rango de direcciones globales")
	}
	return addr
}

// NextLocal asigna la siguiente dirección local
func (vam *VirtualAddressManager) NextLocal() int {
	addr := vam.localCounter
	vam.localCounter++
	if vam.localCounter > 19999 {
		panic("se excedió el rango de direcciones locales")
	}
	return addr
}

// NextTemporal asigna la siguiente dirección temporal
func (vam *VirtualAddressManager) NextTemporal() int {
	addr := vam.temporalCounter
	vam.temporalCounter++
	if vam.temporalCounter > 29999 {
		panic("se excedió el rango de direcciones temporales")
	}
	return addr
}

// NextConstant asigna la siguiente dirección de constante
func (vam *VirtualAddressManager) NextConstant() int {
	addr := vam.constantCounter
	vam.constantCounter++
	if vam.constantCounter > 39999 {
		panic("se excedió el rango de direcciones de constantes")
	}
	return addr
}

// ResetLocals reinicia el contador de locales (al entrar a una nueva función)
func (vam *VirtualAddressManager) ResetLocals() {
	vam.localCounter = vam.LocalBase
}

// ResetTemporals reinicia el contador de temporales (opcional, según necesidad)
func (vam *VirtualAddressManager) ResetTemporals() {
	vam.temporalCounter = vam.TemporalBase
}

// AddressToString convierte una dirección virtual a string
func AddressToString(addr int) string {
	return fmt.Sprintf("%d", addr)
}
