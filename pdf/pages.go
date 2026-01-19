package pdf

// PageTree builds a PDF page tree structure.
// The page tree is a balanced tree that allows efficient access to pages.
type PageTree struct {
	writer *Writer
	root   Ref
	pages  []Ref
}

// NewPageTree creates a new page tree builder.
func NewPageTree(w *Writer) *PageTree {
	root := w.Alloc()
	return &PageTree{
		writer: w,
		root:   root,
	}
}

// Root returns the reference to the page tree root.
func (pt *PageTree) Root() Ref {
	return pt.root
}

// PageCount returns the number of pages.
func (pt *PageTree) PageCount() int {
	return len(pt.pages)
}

// AddPage adds a new page to the tree and returns its reference.
// The page dimensions are specified as width and height in points.
func (pt *PageTree) AddPage(width, height float64) *PageBuilder {
	ref := pt.writer.Alloc()
	pt.pages = append(pt.pages, ref)
	return &PageBuilder{
		writer: pt.writer,
		tree:   pt,
		ref:    ref,
		width:  width,
		height: height,
	}
}

// Finish writes the page tree to the PDF.
// This must be called after all pages have been added.
func (pt *PageTree) Finish() {
	// Build Kids array.
	kids := make(Array, len(pt.pages))
	for i, ref := range pt.pages {
		kids[i] = ref
	}

	// Write Pages dictionary.
	dict := Dict{
		Name("Type"):  Name("Pages"),
		Name("Kids"):  kids,
		Name("Count"): Int(len(pt.pages)),
	}
	pt.writer.Write(pt.root, dict)
}

// PageBuilder builds an individual PDF page.
type PageBuilder struct {
	writer   *Writer
	tree     *PageTree
	ref      Ref
	width    float64
	height   float64
	contents Ref
	resources *Resources
}

// Ref returns the page reference.
func (pb *PageBuilder) Ref() Ref {
	return pb.ref
}

// SetContents sets the page content stream reference.
func (pb *PageBuilder) SetContents(ref Ref) {
	pb.contents = ref
}

// Resources returns the page resources, creating them if needed.
func (pb *PageBuilder) Resources() *Resources {
	if pb.resources == nil {
		pb.resources = NewResources()
	}
	return pb.resources
}

// Finish writes the page object to the PDF.
func (pb *PageBuilder) Finish() {
	dict := Dict{
		Name("Type"):   Name("Page"),
		Name("Parent"): pb.tree.root,
		Name("MediaBox"): Array{
			Real(0),
			Real(0),
			Real(pb.width),
			Real(pb.height),
		},
	}

	if !pb.contents.IsZero() {
		dict[Name("Contents")] = pb.contents
	}

	if pb.resources != nil {
		dict[Name("Resources")] = pb.resources.ToDict()
	} else {
		// Empty resources dictionary.
		dict[Name("Resources")] = Dict{}
	}

	pb.writer.Write(pb.ref, dict)
}

// Resources manages PDF page resources (fonts, images, etc.).
type Resources struct {
	fonts      map[Name]Ref
	xObjects   map[Name]Ref
	extGStates map[Name]Ref
	colorSpaces map[Name]Object
	patterns   map[Name]Ref
	shadings   map[Name]Ref
	procSets   []Name
}

// NewResources creates a new resources manager.
func NewResources() *Resources {
	return &Resources{
		fonts:      make(map[Name]Ref),
		xObjects:   make(map[Name]Ref),
		extGStates: make(map[Name]Ref),
		colorSpaces: make(map[Name]Object),
		patterns:   make(map[Name]Ref),
		shadings:   make(map[Name]Ref),
	}
}

// AddFont adds a font resource.
func (r *Resources) AddFont(name Name, ref Ref) {
	r.fonts[name] = ref
}

// AddXObject adds an XObject (image, form) resource.
func (r *Resources) AddXObject(name Name, ref Ref) {
	r.xObjects[name] = ref
}

// AddExtGState adds an extended graphics state.
func (r *Resources) AddExtGState(name Name, ref Ref) {
	r.extGStates[name] = ref
}

// AddColorSpace adds a color space.
func (r *Resources) AddColorSpace(name Name, obj Object) {
	r.colorSpaces[name] = obj
}

// AddPattern adds a pattern resource.
func (r *Resources) AddPattern(name Name, ref Ref) {
	r.patterns[name] = ref
}

// AddShading adds a shading resource.
func (r *Resources) AddShading(name Name, ref Ref) {
	r.shadings[name] = ref
}

// AddProcSet adds a procedure set name.
func (r *Resources) AddProcSet(name Name) {
	r.procSets = append(r.procSets, name)
}

// ToDict converts resources to a PDF dictionary.
func (r *Resources) ToDict() Dict {
	dict := make(Dict)

	if len(r.fonts) > 0 {
		fonts := make(Dict)
		for name, ref := range r.fonts {
			fonts[name] = ref
		}
		dict[Name("Font")] = fonts
	}

	if len(r.xObjects) > 0 {
		xobjs := make(Dict)
		for name, ref := range r.xObjects {
			xobjs[name] = ref
		}
		dict[Name("XObject")] = xobjs
	}

	if len(r.extGStates) > 0 {
		gstates := make(Dict)
		for name, ref := range r.extGStates {
			gstates[name] = ref
		}
		dict[Name("ExtGState")] = gstates
	}

	if len(r.colorSpaces) > 0 {
		cspaces := make(Dict)
		for name, obj := range r.colorSpaces {
			cspaces[name] = obj
		}
		dict[Name("ColorSpace")] = cspaces
	}

	if len(r.patterns) > 0 {
		patterns := make(Dict)
		for name, ref := range r.patterns {
			patterns[name] = ref
		}
		dict[Name("Pattern")] = patterns
	}

	if len(r.shadings) > 0 {
		shadings := make(Dict)
		for name, ref := range r.shadings {
			shadings[name] = ref
		}
		dict[Name("Shading")] = shadings
	}

	if len(r.procSets) > 0 {
		arr := make(Array, len(r.procSets))
		for i, name := range r.procSets {
			arr[i] = name
		}
		dict[Name("ProcSet")] = arr
	}

	return dict
}
