# Patito - Compilador

Compilador para el lenguaje Patito desarrollado en Go.

## Descripción

Patito es un lenguaje de programación que incluye:
- Variables (int, float)
- Funciones con parámetros y variables locales
- Estructuras de control (if/else, while)
- Llamadas a funciones
- Sentencias de impresión y asignación

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
├── ast/                    # Estructura de datos del AST
├── errors/                 # Manejo de errores
├── patito_test/            # Pruebas unitarias
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

### Parsear un archivo

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
- [Entrega 2](ENTREGA2.md) - Semántica de Variables

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

