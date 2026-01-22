// Module type for Typst.
// Translated from foundations/module.rs

package foundations

// ModuleValue represents an evaluated module.
type ModuleValue struct {
	Module *Module
}

func (ModuleValue) Type() Type         { return TypeModule }
func (v ModuleValue) Display() Content { return Content{} }
func (v ModuleValue) Clone() Value     { return v }
func (ModuleValue) isValue()           {}

// Module represents an evaluated Typst module.
type Module struct {
	// Name is the module name.
	Name string
	// Scope contains the module's exported bindings.
	Scope *Scope
	// Content is the module's evaluated content.
	Content Content
}
