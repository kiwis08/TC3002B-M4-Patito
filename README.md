# Patito - Compilador

Editor con binario de compilador y VM: [Disponible para macOS 26+](https://github.com/kiwis08/Patito/releases/download/submission-part6/PatitoEditor.1.0.dmg)

Compilador para el lenguaje educativo Patito escrito íntegramente en Go. La cadena completa cubre análisis léxico/sintáctico, verificación semántica, generación de cuádruplos y ejecución sobre una máquina virtual ligera.

## Características clave

- Gramática LR(1) definida en `patito.bnf`, procesada con `gocc`.
- Directorio de funciones, tabla de constantes y manejador de tipos para semántica.
- Generación de código intermedio mediante cuádruplos y motor de ejecución en `vm/`.
- Programas de prueba y paquete de tests unitarios (`patito_test/`) para validar cada entrega.
- Documentación consolidada con diagramas, decisiones y guías en [DOCUMENTATION.md](DOCUMENTATION.md).

## Arquitectura general

```
.
├── main.go                 # Punto de entrada del compilador
├── patito.bnf              # Definición léxica y gramatical
├── ast/                    # Estructuras de nodos del AST
├── lexer/, parser/, token/ # Componentes generados por gocc
├── semantic/               # Directorios, cubo semántico y cuádruplos
├── vm/                     # Máquina virtual Patito y formato .patitoc
├── patito_test/            # Suite de pruebas en Go
├── test_programs/          # Casos de uso completos (.patito y .patitoc)
├── DOCUMENTATION.md        # Documentación centralizada
└── ENTREGA*.md             # Reportes por entrega
```

## Requisitos

- Go 1.25 o superior.
- `gocc` en `$GOBIN` (`go install github.com/goccmack/gocc@latest`).
## Instalación rápida

1. Clona el repositorio y entra al directorio `Patito`.
2. Ejecuta `go mod tidy` para descargar dependencias.
3. Genera lexer, parser y tokens desde la gramática (ver sección siguiente).

## Generar lexer y parser

```bash
gocc patito.bnf
```

El comando volverá a crear `lexer/`, `parser/`, `token/` y utilidades auxiliares; ejecútalo cada vez que cambies la gramática.

## Uso básico

### Analizar un archivo y mostrar cuádruplos

```bash
go run . test_programs/test1_arithmetic.patito
```

### Leer desde entrada estándar

```bash
cat programa.patito | go run .
```

### Salida típica

```
OK: parsed Patito successfully

Fila de cuádruplos:
  0: (*, 3, 2, t1)
  1: (+, 5, t1, t2)
  2: (=, t2, , x)
```

En caso de error se imprime un mensaje con el token y la posición involucrada.

## Pruebas y programas de ejemplo

- Ejecuta toda la suite con `go test ./...`.
- Ejecuta únicamente los casos semánticos con `go test ./patito_test`.
- Los archivos dentro de `test_programs/` cubren aritmética, relacionales, ciclos, funciones y recursión; cada `.patito` tiene su `.patitoc` resultante para referencia.

## Documentación y entregas

- Documento principal (arquitectura, diagramas, decisiones, checklist de pruebas): [DOCUMENTATION.md](DOCUMENTATION.md).
- Reportes por entrega:
  - [ENTREGA1.md](ENTREGA1.md) – Parser inicial y resolución de ambigüedades.
  - [ENTREGA2.md](ENTREGA2.md)
  - [ENTREGA3.md](ENTREGA3.md) – Generación de cuádruplos y estructuras auxiliares.
  - [ENTREGA4.md](ENTREGA4.md) – Direcciones virtuales y control de flujo.

## Desarrollo y mantenimiento

- Cuando modifiques `patito.bnf`, regenera los artefactos con `gocc` y actualiza las pruebas relevantes.
- Conserva sincronizadas las tablas de tipos y direcciones en `semantic/` con los bloques de memoria definidos en la VM.
- Añade nuevos programas a `test_programs/` para cubrir casos límite antes de modificar la VM o el generador.
