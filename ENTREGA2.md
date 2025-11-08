### Entrega 2: Semántica inicial

- **Cubo semántico (`semantic/cube.go`)**
  - Representé los operadores como `semantic.Operator` y los tipos válidos del lenguaje como `semantic.Type`.
  - `SemanticCube.Result(op, left, right)` entrega el tipo resultante o error si la combinación no es válida. Para operadores unarios uso `ResultUnary`.
  - El cubo por defecto (`DefaultSemanticCube`) contempla aritmética (`+,-,*,/`), relacionales (`>,<,!=,==`), asignaciones (con promoción `int→float`) y unarios (`+x`, `-x`).

- **Directorios y tablas (`semantic/directory.go`)**
  - `FunctionDirectory` concentra: nombre del programa, tabla global (variables) y mapa de funciones.
  - Cada función (`FunctionEntry`) mantiene dos tablas: parámetros y variables locales. Ambas usan `VariableTable`, que resguarda orden de declaración y detecta duplicados.
  - Los registros intermedios los aporta el parser como `VariableSpec` (nombre, tipo, posición del token) para producir errores con información precisa.
  - Errores semánticos explicitados en `semantic/errors.go`: variables duplicadas por scope, redefinición de funciones y de `program`.

- **Contexto semántico (`semantic/context.go`)**
  - El parser guarda un `*semantic.Context` en `parser.Parser.Context`, de modo que cada reducción puede llenar el directorio y consultar el cubo.
  - `pkg/parser.Adapter` expone `SemanticContext()` para inspeccionar lo que se construyó tras un parse.

- **Puntos neurálgicos (`parser/semantic_actions.go`)**
  - Sin modificar archivos generados, sobreescribo `ReduceFunc` de las producciones clave:
    - `Program`: fija el nombre del programa y vuelca las declaraciones globales.
    - `F_VAR`, `R_ID`, `VARS`: construyen `VariableSpec` para globals/locals.
    - `I_T`, `S_T`, `R_T`: generan la lista ordenada de parámetros.
    - `FUNCS`: registra cada función en el directorio y valida duplicados contra parámetros/locales.
  - Al devolver errores en estas acciones, el parser corta el análisis y reporta la causa semántica (p.ej. “símbolo duplicado”).

- **Pruebas (`patito_test/semantic_test.go`)**
  - Verifico que tras un parse exitoso obtengo un `*semantic.FunctionDirectory` con globals, params y locals en orden.
  - Casos negativos cubren variables duplicadas (global, local, parámetro), choques parámetro/local y redefinición de funciones; todos deben fallar con error semántico.


