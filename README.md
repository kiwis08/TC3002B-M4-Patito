# Patito - Compilador

Compilador para el lenguaje Patito desarrollado en Go.

## Descripción

Patito es un lenguaje de programación que incluye:
- Variables (int, float)
- Funciones con parámetros y variables locales
- Estructuras de control (if/else, while)
- Llamadas a funciones
- Sentencias de impresión y asignación
- Generación de código intermedio mediante cuádruplos

## Requisitos

- Go 1.25 o superior
- gocc (generador de parser) instalado

### Instalación de gocc

```bash
go install github.com/goccmack/gocc
```

## Estructura del Proyecto

```
.
├── patito.bnf              # Definición léxica y gramatical del lenguaje
├── lexer/                  # Lexer generado por gocc
├── parser/                 # Parser generado por gocc
├── token/                  # Tokens generados por gocc
├── semantic/               # Análisis semántico y generación de cuádruplos
├── ast/                    # Estructura de datos del AST
├── errors/                 # Manejo de errores
├── patito_test/            # Pruebas unitarias
├── test_programs/          # Programas de prueba para cuádruplos
├── pkg/                    # Utilidades del parser
├── util/                   # Utilidades generales
└── main.go                 # Punto de entrada del programa
```

## Generación del Parser

Antes de ejecutar el código, es necesario generar el lexer y parser desde la gramática:

```bash
gocc patito.bnf
```

Este comando genera los archivos en los directorios `lexer/`, `parser/`, `token/` y `util/`.

## Ejecución

### Parsear un archivo y ver cuádruplos

```bash
go run . path/to/file.patito
```

### Parsear desde entrada estándar

```bash
cat programa.patito | go run .
```

### Salida esperada

Si el programa es válido:
```
OK: parsed Patito successfully

Fila de cuádruplos:
  0: (*, 3, 2, t1)
  1: (+, 5, t1, t2)
  2: (=, t2, , x)
  ...
```

Si hay errores, se mostrará un mensaje de error descriptivo.

## Pruebas

Ejecutar todas las pruebas:

```bash
go test ./...
```

Ejecutar pruebas en un directorio específico:

```bash
go test ./patito_test
```

## Documentación por Entrega

Cada entrega tiene su propio documento con información específica:

- [Entrega 1](ENTREGA1.md) - Implementación inicial del parser
- [Entrega 3](ENTREGA3.md) - Generación de código intermedio (cuádruplos)

---

# Documentación Consolidada

## Entrega 1: Implementación Inicial del Parser

### Herramientas: qué probé y qué uso

- **Primero intenté con Participle**
  - Probé `github.com/alecthomas/participle` porque permite definir el AST con etiquetas en Go.
  - ¿Por qué no me funcionó bien aquí?
    - La precedencia de operadores se volvía complicada y requería trabajo manual.
    - Resolver ambigüedades clásicas (como el `else` colgante) era menos directo.
    - Quería separar claramente el léxico de la gramática y ver conflictos LR.

- **Me quedé con gocc**
  - Con `gocc` definí léxico y gramática en `patito.bnf` y genero `lexer/`, `parser/`, `token/` y `util/`.
  - Ventajas:
    - Reporte claro de conflictos LR(1) al generar.
    - Separación limpia entre tokens (regex) y reglas gramaticales.
    - Precedencia de expresiones clara usando el patrón `EXP/TERMINO/FACTOR`.
  - Conflictos que arreglé:
    - `if/else`: separé en dos reglas (`if` con y sin `else`) para eliminar el `dangling-else`.
    - Llamadas vs identificadores: `id` puede tener un sufijo `(...)`; si está, es llamada; si no, es `id` normal.

### Cómo definí las reglas (léxico y gramática)

Todo está en `patito.bnf`. Hay dos partes:

- **Léxico (regex)**: define tokens como `id`, `cte_int`, `cte_float`, `cte_string` y lo que se ignora (espacios y comentarios):

```bnf
id          : ('a'-'z' | 'A'-'Z' | '_') { 'a'-'z' | 'A'-'Z' | '0'-'9' | '_' } ;
cte_int     : '0' | ('1'-'9' {'0'-'9'}) ;
cte_float   : {'0'-'9'} '.' {'0'-'9'} ;
cte_string  : '"' { ' '-'!' | '#'-'~' } '"' ;
!whitespace : ' ' | '\t' | '\n' | '\r' ;
```

- **Sintaxis (gramática)**: define la estructura del lenguaje. Ejemplos clave:

```bnf
Program  : "program" id ";" P_VAR P_FUNCS "main" BODY "end" ;

CONDITION
  : "if" "(" EXPRESSION ")" BODY ";"
  | "if" "(" EXPRESSION ")" BODY "else" BODY ";"
  ;

FACTOR
  : S_OP FACTOR_CORE
  ;
FACTOR_CORE
  : "(" EXPRESSION ")"
  | id FACTOR_SUFFIX
  | CTE
  ;
FACTOR_SUFFIX
  : "(" S_E ")"
  | empty
  ;
```

Con esto:
- `if/else` ya no es ambiguo.
- `id` y `id(...)` se distinguen por un sufijo opcional.

### Test-cases principales

Están en `patito_test/patito_test.go` usando `testify/assert`. Casos más importantes:

- **Estructura base**: `program p; main { } end` y errores por faltar `id`, `;` o `end`.
- **Variables**: `var x:int;`, `var x,y,z:float;` y `var x:string;` (debe fallar).
- **Statements**:
  - Asignación: `x = 42;`, `y = (1+2)*3;`
  - Condicionales: `if (x > 0) { x = 1; };` y `if (x < 0) { x = 1; } else { x = 2; };`
  - Ciclo: `while (x != 0) do { x = x - 1; };`
  - `print`: `print("x=", x+1);`
  - Bloques: `[ x = 1; y = 2; ]`
  - Llamadas: `foo(1,2,3);` y como factor `x = (foo(1,2*3));`
- **Funciones**: `void f()[]{};` y `void f(a:int, b:float)[ var x:int; ] { print("ok"); };`
- **Expresiones**: `1 + 2 * 3` vs `(1 + 2) * 3`, y `x > 0`, `x != 0`.
- **Borde**: entrada vacía o solo espacios debe fallar.

---

## Entrega 3: Generación de Código Intermedio (Cuádruplos)

Esta entrega documenta la implementación de la generación de código intermedio mediante cuádruplos para el lenguaje Patito.

### Estructuras de Datos Implementadas

#### 1. Fila de Cuádruplos (QuadrupleQueue)

La **Fila de Cuádruplos** es una estructura de datos tipo cola (FIFO) que almacena todos los cuádruplos generados durante la traducción del código fuente.

**Estructura:**
- `quadruples []Quadruple`: Arreglo que almacena los cuádruplos en orden de generación

**Operaciones principales:**
- `Enqueue(quad Quadruple)`: Agrega un cuádruplo al final de la fila
- `Get() []Quadruple`: Devuelve todos los cuádruplos en orden
- `Size() int`: Devuelve el número de cuádruplos
- `NextIndex() int`: Devuelve el siguiente índice disponible (útil para GOTO)
- `GetAt(index int) *Quadruple`: Obtiene un cuádruplo en un índice específico
- `UpdateAt(index int, quad Quadruple)`: Actualiza un cuádruplo existente (útil para completar GOTO)

**Formato de Cuádruplo:**
Cada cuádruplo tiene la estructura `(operador, operando1, operando2, resultado)`:
- **Operador**: Operación a realizar (+, -, *, /, >, <, !=, ==, =, GOTO, GOTOF, PRINT, etc.)
- **Operando1**: Primer operando (puede ser vacío para operaciones unarias)
- **Operando2**: Segundo operando (puede ser vacío para operaciones unarias o asignaciones)
- **Resultado**: Variable destino o temporal donde se almacena el resultado

#### 2. Pila de Operadores (OperatorStack)

La **Pila de Operadores** almacena los operadores pendientes de procesar durante la evaluación de expresiones, siguiendo el algoritmo de precedencia de operadores.

**Estructura:**
- `operators []string`: Arreglo que funciona como pila (LIFO)

**Operaciones principales:**
- `Push(op string)`: Apila un operador
- `Pop() (string, bool)`: Desapila y devuelve el operador del tope
- `Top() (string, bool)`: Devuelve el operador del tope sin desapilar
- `IsEmpty() bool`: Verifica si la pila está vacía

**Precedencia de operadores:**
- Nivel 3 (mayor): `*`, `/`
- Nivel 2: `+`, `-`
- Nivel 1: `>`, `<`, `!=`, `==`
- Nivel 0 (menor): `=`

#### 3. Pila de Operandos (OperandStack)

La **Pila de Operandos** almacena las direcciones (variables, constantes o temporales) que participan en las operaciones.

**Estructura:**
- `operands []string`: Arreglo que funciona como pila (LIFO)

**Operaciones principales:**
- `Push(operand string)`: Apila un operando
- `Pop() (string, bool)`: Desapila y devuelve el operando del tope
- `Top() (string, bool)`: Devuelve el operando del tope sin desapilar
- `IsEmpty() bool`: Verifica si la pila está vacía

#### 4. Pila de Tipos (TypeStack)

La **Pila de Tipos** mantiene la correspondencia entre operandos y sus tipos semánticos, permitiendo validar operaciones mediante el cubo semántico.

**Estructura:**
- `types []Type`: Arreglo que funciona como pila (LIFO), almacena tipos semánticos

**Operaciones principales:**
- `Push(t Type)`: Apila un tipo
- `Pop() (Type, bool)`: Desapila y devuelve el tipo del tope
- `Top() (Type, bool)`: Devuelve el tipo del tope sin desapilar
- `IsEmpty() bool`: Verifica si la pila está vacía

#### 5. Pila de Saltos (JumpStack)

La **Pila de Saltos** almacena índices de cuádruplos pendientes de completar, utilizada para manejar estructuras de control (if, else, while) que requieren backpatching.

**Estructura:**
- `jumps []int`: Arreglo que funciona como pila (LIFO), almacena índices de cuádruplos

**Operaciones principales:**
- `Push(index int)`: Apila un índice de salto
- `Pop() (int, bool)`: Desapila y devuelve el índice del tope
- `Top() (int, bool)`: Devuelve el índice del tope sin desapilar
- `IsEmpty() bool`: Verifica si la pila está vacía

**Uso:**
- Para `if`: Guarda el índice del GOTOF que se completará al final del bloque
- Para `if-else`: Guarda primero el índice del GOTOF, luego el índice del GOTO del else
- Para `while`: Guarda el índice de inicio del ciclo y el índice del GOTOF de la condición

#### 6. Contador de Temporales (TempCounter)

El **Contador de Temporales** genera nombres únicos para variables temporales creadas durante la evaluación de expresiones.

**Estructura:**
- `counter int`: Contador interno que se incrementa con cada temporal generado

**Operaciones principales:**
- `Next() string`: Genera el siguiente nombre de temporal (t1, t2, t3, ...)
- `Reset()`: Reinicia el contador

### Algoritmos de Traducción Implementados

#### Expresiones Aritméticas

El algoritmo sigue el método de pila de operadores con precedencia:

1. **Al encontrar un operando** (constante o variable):
   - Se apila el operando en la pila de operandos
   - Se apila su tipo en la pila de tipos

2. **Al encontrar un operador**:
   - Mientras haya operadores en la pila con mayor o igual precedencia:
     - Desapilar operador, operandos y tipos
     - Validar operación con cubo semántico
     - Generar cuádruplo
     - Apilar resultado (temporal) y su tipo
   - Apilar el nuevo operador

3. **Al final de la expresión**:
   - Procesar todos los operadores pendientes

**Ejemplo:** `x = 5 + 3 * 2`
```
Cuádruplos generados:
  0: (*, 3, 2, t1)
  1: (+, 5, t1, t2)
  2: (=, t2, , x)
```

#### Expresiones Relacionales

Las expresiones relacionales se procesan de manera similar, pero el operador relacional tiene menor precedencia que los aritméticos:

1. Procesar la expresión aritmética izquierda
2. Apilar el operador relacional
3. Procesar la expresión aritmética derecha
4. Al final, generar el cuádruplo relacional que produce un resultado booleano

**Ejemplo:** `a > b + 1`
```
Cuádruplos generados:
  0: (+, b, 1, t1)
  1: (>, a, t1, t2)
```

#### Estatutos Lineales

##### Asignación

1. Procesar la expresión del lado derecho
2. Obtener el resultado de la pila de operandos
3. Validar compatibilidad de tipos con cubo semántico
4. Generar cuádruplo de asignación: `(=, resultado, , variable)`

##### Print

1. Si es una expresión: procesarla y obtener el resultado
2. Si es un string: usar el valor directamente
3. Generar cuádruplo: `(PRINT, valor, , )`

##### If (sin else)

1. Procesar la expresión condicional
2. Generar GOTOF: `(GOTOF, condición, , )` (índice pendiente)
3. Al final del bloque: completar el GOTOF con el índice actual

##### If-Else

1. Procesar la expresión condicional
2. Generar GOTOF: `(GOTOF, condición, , )` (índice pendiente)
3. Al final del primer bloque: completar GOTOF y generar GOTO incondicional
4. Al final del bloque else: completar el GOTO

##### While

1. Guardar índice de inicio del ciclo
2. Procesar la expresión condicional
3. Generar GOTOF: `(GOTOF, condición, , )` (índice pendiente)
4. Al final del bloque: generar GOTO al inicio y completar GOTOF

### Programas de Prueba

Se han creado varios programas de prueba en el directorio `test_programs/`:

1. **test1_arithmetic.patito**: Expresiones aritméticas básicas
2. **test2_relational.patito**: Expresiones relacionales e if-else
3. **test3_while.patito**: Ciclo while básico
4. **test4_print.patito**: Instrucciones print
5. **test5_complex.patito**: Programa complejo con todas las características

### Ejecución y Visualización

Para ejecutar un programa y ver los cuádruplos generados:

```bash
go run . test_programs/test1_arithmetic.patito
```

La salida mostrará:
1. Confirmación de parseo exitoso
2. Lista completa de cuádruplos generados con sus índices

### Diagramas del Lenguaje Patito

**ESPACIO RESERVADO PARA DIAGRAMAS**

A continuación se incluirán diagramas del lenguaje Patito con los puntos neurálgicos claramente señalados:


### Descripción de Acciones en Puntos Neurálgicos

**ESPACIO RESERVADO PARA DESCRIPCIONES DETALLADAS**

Se documentarán las acciones específicas que se realizan en cada punto neurálgico identificado en los diagramas:


### Notas de Implementación

- Los cuádruplos se generan en orden secuencial
- Los índices de GOTO se completan mediante backpatching al final de los bloques
- Las variables temporales se generan automáticamente (t1, t2, t3, ...)
- La validación semántica se realiza en tiempo de generación mediante el cubo semántico
- Los tipos se mantienen sincronizados con los operandos mediante la pila de tipos

---

## Desarrollo

### Agregar nuevas características

1. Modificar `patito.bnf` con las nuevas reglas léxicas o gramaticales
2. Regenerar el parser con `gocc patito.bnf`
3. Agregar pruebas en `patito_test/patito_test.go`
4. Actualizar la documentación correspondiente

### Manejo de conflictos

Si aparecen conflictos LR(1) durante la generación:
- Revisar la gramática en `patito.bnf`
- Considerar refactorizar reglas ambiguas
- Consultar `ENTREGA1.md` para ejemplos de resolución de conflictos comunes

## Herramientas Utilizadas

- **gocc**: Generador de parser LR(1) para Go
- **testify**: Framework de pruebas para Go
- **Go**: Lenguaje de programación base

## Estado del Proyecto

Este proyecto está en desarrollo activo. Cada entrega documenta el progreso y las características implementadas.
