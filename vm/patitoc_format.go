package vm

import (
	"Patito/semantic"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	PATITOC_MAGIC   = 0x50415449 // "PATI" en ASCII
	PATITOC_VERSION = 1
)

type PatitocWriter struct {
	w io.Writer
}

type PatitocReader struct {
	r io.Reader
}

// Header del archivo
type PatitocHeader struct {
	Magic       uint32   // 0x50415449 ("PATI")
	Version     uint16   // Versión del formato
	QuadCount   uint32   // Número de cuádruplos
	ConstCount  uint32   // Número de constantes
	FuncCount   uint32   // Número de funciones
	GlobalCount uint32   // Número de variables globales
	Reserved    [16]byte // Para futuras extensiones
}

func SavePatitoc(ctx *semantic.Context, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := &PatitocWriter{w: file}
	fmt.Print("Writing .patitoc file")
	return writer.Write(ctx)
}

func (pw *PatitocWriter) Write(ctx *semantic.Context) error {
	quads := ctx.Quadruples.Get()
	constants := ctx.ConstantTable.Entries()

	// 1. Escribir header
	header := PatitocHeader{
		Magic:       PATITOC_MAGIC,
		Version:     PATITOC_VERSION,
		QuadCount:   uint32(len(quads)),
		ConstCount:  uint32(len(constants)),
		FuncCount:   uint32(len(ctx.Directory.Functions)),
		GlobalCount: uint32(len(ctx.Directory.Globals.Entries())),
	}

	if err := binary.Write(pw.w, binary.LittleEndian, &header); err != nil {
		return err
	}

	// 2. Escribir nombre del programa
	progName := []byte(ctx.Directory.ProgramName)
	if err := pw.writeString(progName); err != nil {
		return err
	}

	// 3. Escribir variables globales
	if err := pw.writeGlobals(ctx.Directory.Globals); err != nil {
		return err
	}

	// 4. Escribir funciones
	if err := pw.writeFunctions(ctx.Directory.Functions, ctx.FunctionStartQuads); err != nil {
		return err
	}

	// 5. Escribir constantes
	if err := pw.writeConstants(constants); err != nil {
		return err
	}

	// 6. Escribir cuádruplos
	if err := pw.writeQuadruples(quads); err != nil {
		return err
	}

	// 7. Escribir mapa de tipos (para temporales y validación)
	if err := pw.writeTypeMap(ctx); err != nil {
		return err
	}

	return nil
}

func (pw *PatitocWriter) writeString(s []byte) error {
	// Escribir longitud + datos
	length := uint16(len(s))
	if err := binary.Write(pw.w, binary.LittleEndian, length); err != nil {
		return err
	}
	if length > 0 {
		_, err := pw.w.Write(s)
		return err
	}
	return nil
}

func (pw *PatitocWriter) writeGlobals(globals *semantic.VariableTable) error {
	entries := globals.Entries()
	for _, entry := range entries {
		// Escribir: nombre, tipo, dirección
		if err := pw.writeString([]byte(entry.Name)); err != nil {
			return err
		}
		if err := binary.Write(pw.w, binary.LittleEndian, uint8(entry.Type)); err != nil {
			return err
		}
		if err := binary.Write(pw.w, binary.LittleEndian, uint32(entry.Address)); err != nil {
			return err
		}
	}
	return nil
}

func (pw *PatitocWriter) writeFunctions(functions map[string]*semantic.FunctionEntry, startQuads map[string]int) error {
	for name, fn := range functions {
		// Nombre de función
		if err := pw.writeString([]byte(name)); err != nil {
			return err
		}

		// Tipo de retorno
		if err := binary.Write(pw.w, binary.LittleEndian, uint8(fn.ReturnType)); err != nil {
			return err
		}

		// Índice de inicio
		startQuad := -1
		if sq, ok := startQuads[name]; ok {
			startQuad = sq
		}
		if err := binary.Write(pw.w, binary.LittleEndian, int32(startQuad)); err != nil {
			return err
		}

		// Parámetros
		params := fn.Params.Entries()
		if err := binary.Write(pw.w, binary.LittleEndian, uint16(len(params))); err != nil {
			return err
		}
		for _, param := range params {
			if err := pw.writeString([]byte(param.Name)); err != nil {
				return err
			}
			if err := binary.Write(pw.w, binary.LittleEndian, uint8(param.Type)); err != nil {
				return err
			}
			if err := binary.Write(pw.w, binary.LittleEndian, uint32(param.Address)); err != nil {
				return err
			}
		}

		// Locales
		locals := fn.Locals.Entries()
		if err := binary.Write(pw.w, binary.LittleEndian, uint16(len(locals))); err != nil {
			return err
		}
		for _, local := range locals {
			if err := pw.writeString([]byte(local.Name)); err != nil {
				return err
			}
			if err := binary.Write(pw.w, binary.LittleEndian, uint8(local.Type)); err != nil {
				return err
			}
			if err := binary.Write(pw.w, binary.LittleEndian, uint32(local.Address)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (pw *PatitocWriter) writeConstants(constants []*semantic.ConstantEntry) error {
	for _, entry := range constants {
		// Tipo
		if err := binary.Write(pw.w, binary.LittleEndian, uint8(entry.Type)); err != nil {
			return err
		}

		// Dirección
		if err := binary.Write(pw.w, binary.LittleEndian, uint32(entry.Address)); err != nil {
			return err
		}

		// Valor (como string, la VM lo parsea según el tipo)
		if err := pw.writeString([]byte(entry.Value)); err != nil {
			return err
		}
	}
	return nil
}

func (pw *PatitocWriter) writeQuadruples(quads []semantic.Quadruple) error {
	for _, quad := range quads {
		// Operator (string)
		if err := pw.writeString([]byte(quad.Operator)); err != nil {
			return err
		}

		// Operand1 (string, puede ser dirección o vacío)
		if err := pw.writeString([]byte(quad.Operand1)); err != nil {
			return err
		}

		// Operand2 (string)
		if err := pw.writeString([]byte(quad.Operand2)); err != nil {
			return err
		}

		// Result (string)
		if err := pw.writeString([]byte(quad.Result)); err != nil {
			return err
		}
	}
	return nil
}

func (pw *PatitocWriter) writeTypeMap(ctx *semantic.Context) error {
	// Construir mapa de tipos (dirección -> tipo)
	typeMap := make(map[uint32]uint8)

	// Globales
	for _, entry := range ctx.Directory.Globals.Entries() {
		typeMap[uint32(entry.Address)] = uint8(entry.Type)
	}

	// Funciones
	for _, fn := range ctx.Directory.Functions {
		for _, param := range fn.Params.Entries() {
			typeMap[uint32(param.Address)] = uint8(param.Type)
		}
		for _, local := range fn.Locals.Entries() {
			typeMap[uint32(local.Address)] = uint8(local.Type)
		}
	}

	// Constantes
	for _, entry := range ctx.ConstantTable.Entries() {
		typeMap[uint32(entry.Address)] = uint8(entry.Type)
	}

	// Escribir cantidad de entradas
	if err := binary.Write(pw.w, binary.LittleEndian, uint32(len(typeMap))); err != nil {
		return err
	}

	// Escribir entradas
	for addr, t := range typeMap {
		if err := binary.Write(pw.w, binary.LittleEndian, addr); err != nil {
			return err
		}
		if err := binary.Write(pw.w, binary.LittleEndian, t); err != nil {
			return err
		}
	}

	return nil
}
