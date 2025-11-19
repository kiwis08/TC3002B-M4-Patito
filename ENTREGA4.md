# Patito – Entrega 4: Direcciones Virtuales y Cuádruplos para Estatutos de Control

Esta entrega documenta la implementación de la traducción a direcciones virtuales y la generación de cuádruplos para estatutos condicionales y cíclicos en el lenguaje Patito.

## Sistema de Direcciones Virtuales

### Rangos de Direcciones Virtuales

El compilador asigna direcciones virtuales según los siguientes rangos estándar de Patito:

- **Variables globales**: 1000-9999
- **Variables locales**: 10000-19999
- **Temporales**: 20000-29999
- **Constantes**: 30000-39999

### VirtualAddressManager

El `VirtualAddressManager` gestiona la asignación de direcciones virtuales para cada tipo de entidad:

```go
type VirtualAddressManager struct {
    GlobalBase    int // 1000
    LocalBase     int // 10000
    TemporalBase  int // 20000
    ConstantBase  int // 30000
    globalCounter    int
    localCounter     int
    temporalCounter  int
    constantCounter  int
}
```

**Operaciones principales:**
- `NextGlobal()`: Asigna la siguiente dirección global
- `NextLocal()`: Asigna la siguiente dirección local
- `NextTemporal()`: Asigna la siguiente dirección temporal
- `NextConstant()`: Asigna la siguiente dirección de constante
- `ResetLocals()`: Reinicia el contador de locales (al entrar a una nueva función)

### Asignación de Direcciones a Variables

Las direcciones virtuales se asignan automáticamente cuando las variables se declaran:

1. **Variables globales**: Se asignan direcciones en el rango 1000-9999 cuando se procesan en `AddGlobals`.
2. **Variables locales**: Se asignan direcciones en el rango 10000-19999 cuando se procesan en `AddFunction` (parámetros y locales).
3. **Variables usadas antes de declaración**: Si una variable se usa antes de que se agregue al directorio (durante el parsing), se asigna una dirección temporalmente para permitir su uso inmediato.

**Ejemplo:**
```go
// En reduceVarDeclaration
specs := make([]*semantic.VariableSpec, 0, len(idTokens))
for _, tok := range idTokens {
    varName := tok.IDValue()
    ctx.VariableTypes[varName] = typeVal
    specs = append(specs, &semantic.VariableSpec{
        Name: varName,
        Type: typeVal,
        Pos:  tok.Pos,
    })
}

// En AddGlobals
for _, spec := range specs {
    if spec.Address == 0 {
        spec.Address = addressManager.NextGlobal()
    }
}
```

### Tabla de Constantes

La `ConstantTable` almacena todas las constantes del programa con sus direcciones virtuales:

```go
type ConstantEntry struct {
    Value   string // Valor de la constante (como string)
    Type    Type   // Tipo de la constante
    Address int    // Dirección virtual asignada
}
```

Las constantes se almacenan de forma única: si la misma constante (mismo valor y tipo) se usa múltiples veces, se reutiliza la misma dirección virtual.

**Ejemplo:**
```go
// En PushConstant
entry, exists := ctx.ConstantTable.Get(value, operandType)
if !exists {
    address := ctx.AddressManager.NextConstant()
    entry = ctx.ConstantTable.Add(value, operandType, address)
}
```

### Temporales

Los temporales se generan automáticamente durante la evaluación de expresiones y se asignan direcciones en el rango 20000-29999:

```go
// En ProcessOperator
temp := ctx.TempCounter.NextString() // Genera dirección virtual como string
generateQuadruple(ctx, topOp, left, right, temp)
```

## Traducción de Cuádruplos a Direcciones Virtuales

Todos los cuádruplos ahora usan direcciones virtuales en lugar de nombres de variables o constantes:

### Variables

Cuando se usa una variable, se busca su dirección virtual y se usa en el cuádruplo:

```go
// En PushVariable
address, err := GetVariableAddressFromContext(ctx, varName)
operand := AddressToString(address)
PushOperand(ctx, operand, varType)
```

### Constantes

Cuando se usa una constante, se busca o crea su entrada en la tabla de constantes y se usa su dirección virtual:

```go
// En PushConstant
entry, exists := ctx.ConstantTable.Get(value, operandType)
if !exists {
    address := ctx.AddressManager.NextConstant()
    entry = ctx.ConstantTable.Add(value, operandType, address)
}
operand := AddressToString(entry.Address)
```

### Temporales

Los temporales se generan automáticamente con direcciones virtuales:

```go
// En ProcessOperator
temp := ctx.TempCounter.NextString() // Genera dirección virtual como string
generateQuadruple(ctx, topOp, left, right, temp)
```

## Generación de Cuádruplos para Estatutos Condicionales

### If (sin else)

El algoritmo para generar cuádruplos de un `if` sin `else` es:

1. Procesar la expresión condicional (produce un resultado booleano en un temporal)
2. Generar `GOTOF` con el resultado de la condición (índice de salto pendiente)
3. Procesar el cuerpo del `if`
4. Al final del cuerpo, completar el `GOTOF` con el índice actual

**Ejemplo:** `if (x > 0) { x = 1; };`

```
Cuádruplos generados:
  0: (>, x_addr, 0, t1)      // Evaluar condición
  1: (GOTOF, t1, , 3)        // Salto si falso (índice pendiente)
  2: (=, 1, , x_addr)        // Cuerpo del if
  3: (continuación...)        // Completar GOTOF apunta aquí
```

**Implementación:**
```go
func ProcessIf(ctx *Context) (int, error) {
    condition, ok := ctx.OperandStack.Pop()
    ctx.TypeStack.Pop()
    
    gotoIndex := ctx.Quadruples.NextIndex()
    generateQuadruple(ctx, "GOTOF", condition, "", "")
    ctx.JumpStack.Push(gotoIndex)
    
    return gotoIndex, nil
}

func ProcessIfEnd(ctx *Context) error {
    jumpIndex, ok := ctx.JumpStack.Pop()
    quad := ctx.Quadruples.GetAt(jumpIndex)
    if quad != nil {
        quad.Result = fmt.Sprintf("%d", ctx.Quadruples.NextIndex())
        ctx.Quadruples.UpdateAt(jumpIndex, *quad)
    }
    return nil
}
```

### If-Else

El algoritmo para generar cuádruplos de un `if-else` es:

1. Procesar la expresión condicional
2. Generar `GOTOF` con el resultado de la condición (índice de salto pendiente)
3. Procesar el cuerpo del `if`
4. Al final del primer cuerpo: completar `GOTOF` y generar `GOTO` incondicional (índice pendiente)
5. Procesar el cuerpo del `else`
6. Al final del segundo cuerpo: completar el `GOTO`

**Ejemplo:** `if (x > 0) { x = 1; } else { x = 2; };`

```
Cuádruplos generados:
  0: (>, x_addr, 0, t1)      // Evaluar condición
  1: (GOTOF, t1, , 3)        // Salto si falso (índice pendiente)
  2: (=, 1, , x_addr)        // Cuerpo del if
  3: (GOTO, , , 4)           // Salto incondicional (índice pendiente)
  4: (=, 2, , x_addr)        // Cuerpo del else
  5: (continuación...)        // Completar GOTO apunta aquí
```

**Implementación:**
```go
func ProcessElse(ctx *Context) (int, error) {
    jumpIndex, ok := ctx.JumpStack.Pop() // GOTOF del if
    gotoIndex := ctx.Quadruples.NextIndex()
    generateQuadruple(ctx, "GOTO", "", "", "")
    
    // Actualizar el GOTOF
    quad := ctx.Quadruples.GetAt(jumpIndex)
    if quad != nil {
        quad.Result = fmt.Sprintf("%d", ctx.Quadruples.NextIndex())
        ctx.Quadruples.UpdateAt(jumpIndex, *quad)
    }
    
    ctx.JumpStack.Push(gotoIndex)
    return gotoIndex, nil
}
```

## Generación de Cuádruplos para Estatutos Cíclicos

### While

El algoritmo para generar cuádruplos de un `while` es:

1. Guardar el índice donde comienza el ciclo (donde se evalúa la condición)
2. Procesar la expresión condicional (produce un resultado booleano)
3. Generar `GOTOF` con el resultado de la condición (índice de salto pendiente)
4. Procesar el cuerpo del `while`
5. Al final del cuerpo: generar `GOTO` al inicio del ciclo y completar `GOTOF`

**Ejemplo:** `while (i < 10) do { i = i + 1; };`

```
Cuádruplos generados:
  0: (=, 0, , i_addr)        // Inicialización
  1: (<, i_addr, 10, t1)     // Evaluar condición (índice de inicio)
  2: (+, i_addr, 1, t2)      // Cuerpo del while
  3: (=, t2, , i_addr)
  4: (GOTOF, t1, , 6)        // Salto si falso (índice pendiente)
  5: (GOTO, , , 1)            // Volver al inicio
  6: (continuación...)        // Completar GOTOF apunta aquí
```

**Implementación:**
```go
func ProcessWhileCondition(ctx *Context) error {
    condition, ok := ctx.OperandStack.Pop()
    ctx.TypeStack.Pop()
    
    // Guardar índice de inicio (donde se evalúa la condición)
    quadSize := ctx.Quadruples.Size()
    startIndex := quadSize - 1
    ctx.JumpStack.Push(startIndex)
    
    // Generar GOTOF
    gotoIndex := ctx.Quadruples.NextIndex()
    generateQuadruple(ctx, "GOTOF", condition, "", "")
    ctx.JumpStack.Push(gotoIndex)
    
    return nil
}

func ProcessWhileEnd(ctx *Context) error {
    gotoIndex, _ := ctx.JumpStack.Pop() // GOTOF
    startIndex, _ := ctx.JumpStack.Pop() // Índice de inicio
    
    // Generar GOTO al inicio
    generateQuadruple(ctx, "GOTO", "", "", fmt.Sprintf("%d", startIndex))
    
    // Completar GOTOF
    quad := ctx.Quadruples.GetAt(gotoIndex)
    if quad != nil {
        quad.Result = fmt.Sprintf("%d", ctx.Quadruples.NextIndex())
        ctx.Quadruples.UpdateAt(gotoIndex, *quad)
    }
    
    return nil
}
```

## Ejemplos de Salida

### Ejemplo 1: Expresiones Aritméticas

**Programa:**
```patito
program test1;
var x, y, z: int;
main {
    x = 5 + 3 * 2;
    y = (10 - 4) / 2;
    z = x + y;
}
end
```

**Cuádruplos generados:**
```
  0: (*, 30001, 30002, 20000)  // 3 * 2
  1: (+, 30000, 20000, 20001)  // 5 + t1
  2: (=, 20001, , 1000)         // x = t2
  3: (-, 30003, 30004, 20002)   // 10 - 4
  4: (/, 20002, 30002, 20003)   // t2 / 2
  5: (=, 20003, , 1001)         // y = t3
  6: (+, 1000, 1001, 20004)     // x + y
  7: (=, 20004, , 1002)         // z = t4
```

**Direcciones:**
- Variables: x=1000, y=1001, z=1002
- Constantes: 5=30000, 3=30001, 2=30002, 10=30003, 4=30004
- Temporales: t1=20000, t2=20001, t3=20002, t4=20003, t5=20004

### Ejemplo 2: Condicionales

**Programa:**
```patito
program test2;
var a, b: int;
main {
    a = 5;
    b = 3;
    if (a > b) {
        a = 1;
    } else {
        b = 2;
    };
}
end
```

**Cuádruplos generados:**
```
  0: (=, 30000, , 1000)         // a = 5
  1: (=, 30001, , 1001)         // b = 3
  2: (>, 1000, 1001, 20000)     // a > b
  3: (GOTOF, 20000, , 5)        // Salto si falso
  4: (=, 30002, , 1000)         // a = 1
  5: (GOTO, , , 6)              // Salto incondicional
  6: (=, 30003, , 1001)         // b = 2
  7: (continuación...)
```

### Ejemplo 3: Ciclo While

**Programa:**
```patito
program test3;
var i: int;
main {
    i = 0;
    while (i < 10) do {
        i = i + 1;
    };
}
end
```

**Cuádruplos generados:**
```
  0: (=, 30000, , 1000)         // i = 0
  1: (<, 1000, 30001, 20000)    // i < 10 (inicio del ciclo)
  2: (+, 1000, 30002, 20001)    // i + 1
  3: (=, 20001, , 1000)         // i = t2
  4: (GOTOF, 20000, , 6)        // Salto si falso
  5: (GOTO, , , 1)              // Volver al inicio
  6: (continuación...)
```

## Notas de Implementación

1. **Direcciones Virtuales**: Todas las variables, constantes y temporales ahora usan direcciones virtuales en lugar de nombres en los cuádruplos.

2. **Backpatching**: Los índices de salto en `GOTO` y `GOTOF` se completan mediante backpatching al final de los bloques.

3. **Tabla de Constantes**: Las constantes se almacenan de forma única para evitar duplicados.

4. **Gestión de Ámbito**: El contador de direcciones locales se reinicia al entrar a una nueva función.

5. **Orden de Procesamiento**: Las direcciones se asignan durante el parsing, permitiendo su uso inmediato en expresiones y estatutos.

## Estructuras de Datos Adicionales

### ConstantTable

Mantiene un registro de todas las constantes del programa:

```go
type ConstantTable struct {
    constants map[string]*ConstantEntry
    order     []*ConstantEntry
}
```

### VariableEntry (actualizado)

Ahora incluye la dirección virtual:

```go
type VariableEntry struct {
    Name       string
    Type       Type
    Scope      ScopeKind
    DeclaredAt token.Pos
    Address    int // Dirección virtual asignada
}
```

### Context (actualizado)

Incluye el gestor de direcciones virtuales y la tabla de constantes:

```go
type Context struct {
    Directory         *FunctionDirectory
    Cube              *SemanticCube
    Quadruples        *QuadrupleQueue
    OpStack           *OperatorStack
    OperandStack      *OperandStack
    TypeStack         *TypeStack
    JumpStack         *JumpStack
    TempCounter       *TempCounter
    AddressManager    *VirtualAddressManager
    ConstantTable     *ConstantTable
    VariableTypes     map[string]Type
    VariableAddresses map[string]int
}
```

