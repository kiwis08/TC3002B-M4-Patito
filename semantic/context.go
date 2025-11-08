package semantic

// Context es el objeto que asignamos a parser.Context para compartir estado
// entre las acciones semánticas.
type Context struct {
	Directory *FunctionDirectory
	Cube      *SemanticCube
}

func NewContext() *Context {
	return &Context{
		Directory: NewFunctionDirectory(),
		Cube:      DefaultSemanticCube,
	}
}

// DirectorySnapshot expone el directorio final tras el parseo.
func (c *Context) DirectorySnapshot() *FunctionDirectory {
	return c.Directory
}
