## Patito – Documento de la entrega

Aquí explico qué hice, cómo definí las reglas del lenguaje y qué pruebas uso. Lo iré ampliando en próximas entregas.

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

### Generación y ejecución

- Generar lexer/parser desde la gramática:

```bash
gocc patito.bnf
```

- Correr pruebas (en mi máquina):

```bash
go test ./...
```

- Ejecutar el parser:

```bash
# Desde archivo
go run . path/al/archivo.patito

# Desde stdin
cat programa.patito | go run .
```

Salida esperada si todo va bien:

```
OK: parsed Patito successfully
```

### Estructura del proyecto (parcial)

- `patito.bnf`: reglas léxicas y gramaticales.
- `lexer/`, `parser/`, `token/`, `util`: generado por `gocc`.
- `patito_test/`: pruebas del lenguaje.
- `main.go`: CLI de parseo.
- `errors/`, `ast/`, `pkg/`: módulos de soporte.

### Siguientes pasos

- Voy a mantener la gramática sin conflictos LR(1) tras cambios.
- Agregaré nuevas features manteniendo la precedencia y claridad del AST.
- Documentaré nuevos casos y decisiones en este README conforme avance el proyecto.

