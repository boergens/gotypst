package eval

import (
	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/library/layout"
	"github.com/boergens/gotypst/syntax"
)

// PadFunc creates the pad element function.
func PadFunc() *Func {
	name := "pad"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: padNative,
			Info: layout.PadDef.ToFuncInfo(),
		},
	}
}

// padNative implements the pad() function using the generic element parser.
func padNative(engine foundations.Engine, context foundations.Context, args *Args) (Value, error) {
	elem, err := foundations.ParseElement[layout.PadElement](layout.PadDef, args)
	if err != nil {
		return nil, err
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}
