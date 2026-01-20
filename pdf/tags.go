// Package pdf provides PDF export functionality including PDF/UA accessibility tagging.
package pdf

import (
	"github.com/boergens/gotypst/layout/pages"
)

// StructRole represents standard PDF structure roles for accessibility.
type StructRole string

// Standard PDF/UA structure roles.
const (
	// Document structure roles
	RoleDocument StructRole = "Document"
	RolePart     StructRole = "Part"
	RoleSect     StructRole = "Sect"
	RoleDiv      StructRole = "Div"

	// Block-level structure roles
	RoleP          StructRole = "P"
	RoleH          StructRole = "H"
	RoleH1         StructRole = "H1"
	RoleH2         StructRole = "H2"
	RoleH3         StructRole = "H3"
	RoleH4         StructRole = "H4"
	RoleH5         StructRole = "H5"
	RoleH6         StructRole = "H6"
	RoleBlockQuote StructRole = "BlockQuote"
	RoleCaption    StructRole = "Caption"

	// List roles
	RoleL     StructRole = "L"
	RoleLI    StructRole = "LI"
	RoleLbl   StructRole = "Lbl"
	RoleLBody StructRole = "LBody"

	// Table roles
	RoleTable StructRole = "Table"
	RoleTHead StructRole = "THead"
	RoleTBody StructRole = "TBody"
	RoleTFoot StructRole = "TFoot"
	RoleTR    StructRole = "TR"
	RoleTH    StructRole = "TH"
	RoleTD    StructRole = "TD"

	// Inline roles
	RoleSpan   StructRole = "Span"
	RoleLink   StructRole = "Link"
	RoleCode   StructRole = "Code"
	RoleQuote  StructRole = "Quote"
	RoleEm     StructRole = "Em"
	RoleStrong StructRole = "Strong"

	// Special content roles
	RoleFigure   StructRole = "Figure"
	RoleFormula  StructRole = "Formula"
	RoleForm     StructRole = "Form"
	RoleArtifact StructRole = "Artifact"

	// Navigation roles
	RoleTOC      StructRole = "TOC"
	RoleTOCI     StructRole = "TOCI"
	RoleIndex    StructRole = "Index"
	RoleNonStruct StructRole = "NonStruct"
)

// StructElem represents a structure element in the PDF tag tree.
type StructElem struct {
	// Role is the structure type (e.g., "P", "H1", "Figure").
	Role StructRole
	// Ref is the PDF object reference for this element.
	Ref Ref
	// Parent is the parent structure element reference.
	Parent Ref
	// Kids contains child elements or marked content IDs.
	Kids []StructKid
	// PageRef is the page containing this element's content.
	PageRef Ref
	// AltText is alternative text for accessibility.
	AltText string
	// ActualText is the actual text content.
	ActualText string
	// Lang is the language tag (e.g., "en-US").
	Lang string
}

// StructKid represents a child of a structure element.
// It can be either another StructElem reference or a marked content reference.
type StructKid interface {
	isStructKid()
}

// StructKidElem is a child structure element.
type StructKidElem struct {
	Ref Ref
}

func (StructKidElem) isStructKid() {}

// StructKidMCID is a marked content identifier.
type StructKidMCID struct {
	MCID int
}

func (StructKidMCID) isStructKid() {}

// MarkedContent represents a marked content region in a content stream.
type MarkedContent struct {
	// MCID is the marked content identifier.
	MCID int
	// Role is the structure role for this content.
	Role StructRole
	// PageIndex is the page containing this content.
	PageIndex int
	// Parent is the parent structure element.
	Parent *StructElem
}

// TagManager manages PDF accessibility tag state during rendering.
type TagManager struct {
	// nextMCID is the next available marked content ID.
	nextMCID int
	// structElems holds all structure elements.
	structElems []*StructElem
	// activeStack tracks the current tag nesting.
	activeStack []*StructElem
	// mcByPage maps page index to marked content regions on that page.
	mcByPage map[int][]MarkedContent
	// currentPageIndex is the current page being rendered.
	currentPageIndex int
	// rootElem is the document root structure element.
	rootElem *StructElem
	// pageRefs maps page index to page object reference.
	pageRefs map[int]Ref
	// customRoles maps custom role names to standard roles.
	customRoles map[string]StructRole
}

// NewTagManager creates a new tag manager.
func NewTagManager() *TagManager {
	tm := &TagManager{
		nextMCID:    0,
		mcByPage:    make(map[int][]MarkedContent),
		pageRefs:    make(map[int]Ref),
		customRoles: make(map[string]StructRole),
	}

	// Create document root element
	tm.rootElem = &StructElem{
		Role: RoleDocument,
	}
	tm.structElems = append(tm.structElems, tm.rootElem)
	tm.activeStack = append(tm.activeStack, tm.rootElem)

	return tm
}

// SetPageRef associates a page index with its PDF object reference.
func (tm *TagManager) SetPageRef(pageIndex int, ref Ref) {
	tm.pageRefs[pageIndex] = ref
}

// SetCurrentPage sets the current page being rendered.
func (tm *TagManager) SetCurrentPage(pageIndex int) {
	tm.currentPageIndex = pageIndex
}

// AllocMCID allocates a new marked content ID.
func (tm *TagManager) AllocMCID() int {
	mcid := tm.nextMCID
	tm.nextMCID++
	return mcid
}

// MapRole maps a custom role name to a standard PDF role.
func (tm *TagManager) MapRole(custom string, standard StructRole) {
	tm.customRoles[custom] = standard
}

// ResolveRole resolves a role name to its standard PDF role.
func (tm *TagManager) ResolveRole(role string) StructRole {
	if standard, ok := tm.customRoles[role]; ok {
		return standard
	}
	// Check if it's already a standard role
	switch StructRole(role) {
	case RoleDocument, RolePart, RoleSect, RoleDiv, RoleP,
		RoleH, RoleH1, RoleH2, RoleH3, RoleH4, RoleH5, RoleH6,
		RoleBlockQuote, RoleCaption, RoleL, RoleLI, RoleLbl, RoleLBody,
		RoleTable, RoleTHead, RoleTBody, RoleTFoot, RoleTR, RoleTH, RoleTD,
		RoleSpan, RoleLink, RoleCode, RoleQuote, RoleEm, RoleStrong,
		RoleFigure, RoleFormula, RoleForm, RoleArtifact,
		RoleTOC, RoleTOCI, RoleIndex, RoleNonStruct:
		return StructRole(role)
	}
	// Default to Span for unknown roles
	return RoleSpan
}

// BeginTag starts a new tagged region.
func (tm *TagManager) BeginTag(role StructRole) *StructElem {
	elem := &StructElem{
		Role: role,
	}

	// Set parent from active stack
	if len(tm.activeStack) > 0 {
		parent := tm.activeStack[len(tm.activeStack)-1]
		elem.Parent = parent.Ref
		parent.Kids = append(parent.Kids, StructKidElem{Ref: elem.Ref})
	}

	tm.structElems = append(tm.structElems, elem)
	tm.activeStack = append(tm.activeStack, elem)

	return elem
}

// EndTag ends the current tagged region.
func (tm *TagManager) EndTag() {
	if len(tm.activeStack) > 1 {
		tm.activeStack = tm.activeStack[:len(tm.activeStack)-1]
	}
}

// CurrentElement returns the current active structure element.
func (tm *TagManager) CurrentElement() *StructElem {
	if len(tm.activeStack) == 0 {
		return nil
	}
	return tm.activeStack[len(tm.activeStack)-1]
}

// AddMarkedContent adds a marked content region to the current page.
func (tm *TagManager) AddMarkedContent(mcid int, role StructRole) {
	mc := MarkedContent{
		MCID:      mcid,
		Role:      role,
		PageIndex: tm.currentPageIndex,
		Parent:    tm.CurrentElement(),
	}
	tm.mcByPage[tm.currentPageIndex] = append(tm.mcByPage[tm.currentPageIndex], mc)

	// Add MCID to current element's kids
	if elem := tm.CurrentElement(); elem != nil {
		elem.Kids = append(elem.Kids, StructKidMCID{MCID: mcid})
	}
}

// ProcessTag processes a tag item from the layout.
// Returns the role and MCID if this is a start tag that needs marked content.
func (tm *TagManager) ProcessTag(tag *pages.Tag) (role StructRole, mcid int, isStart bool) {
	if tag.Kind == pages.TagStart {
		// Determine role from tag element
		role = tm.determineRole(tag)
		elem := tm.BeginTag(role)
		mcid = tm.AllocMCID()
		tm.AddMarkedContent(mcid, role)
		elem.PageRef = tm.pageRefs[tm.currentPageIndex]
		return role, mcid, true
	}

	// End tag
	tm.EndTag()
	return "", 0, false
}

// determineRole determines the structure role from a tag element.
func (tm *TagManager) determineRole(tag *pages.Tag) StructRole {
	if tag.Elem == nil {
		return RoleSpan
	}

	// Check for counter update elements to determine role
	if counter, ok := tag.Elem.(*pages.CounterUpdateElem); ok {
		switch counter.Key {
		case pages.CounterKeyHeading:
			// Could determine heading level from counter state
			// For now, return generic heading
			return RoleH
		case pages.CounterKeyFigure:
			return RoleFigure
		case pages.CounterKeyTable:
			return RoleTable
		case pages.CounterKeyEquation:
			return RoleFormula
		}
	}

	return RoleSpan
}

// RootElement returns the document root structure element.
func (tm *TagManager) RootElement() *StructElem {
	return tm.rootElem
}

// StructElements returns all structure elements.
func (tm *TagManager) StructElements() []*StructElem {
	return tm.structElems
}

// MarkedContentForPage returns marked content regions for a page.
func (tm *TagManager) MarkedContentForPage(pageIndex int) []MarkedContent {
	return tm.mcByPage[pageIndex]
}

// CustomRoles returns the custom role mappings.
func (tm *TagManager) CustomRoles() map[string]StructRole {
	return tm.customRoles
}

// HasTags returns true if any tags were processed.
func (tm *TagManager) HasTags() bool {
	return len(tm.structElems) > 1 || tm.nextMCID > 0
}

// BuildParentTree builds the parent tree for reverse MCID lookup.
// Returns a map from MCID to parent structure element reference.
func (tm *TagManager) BuildParentTree() map[int]Ref {
	parentTree := make(map[int]Ref)
	for pageIndex, mcs := range tm.mcByPage {
		_ = pageIndex
		for _, mc := range mcs {
			if mc.Parent != nil {
				parentTree[mc.MCID] = mc.Parent.Ref
			}
		}
	}
	return parentTree
}

// AssignRefs assigns PDF object references to all structure elements.
func (tm *TagManager) AssignRefs(allocRef func() Ref) {
	for _, elem := range tm.structElems {
		elem.Ref = allocRef()
	}

	// Update parent references and kids after refs are assigned
	for _, elem := range tm.structElems {
		// Find parent and update reference
		for _, potential := range tm.structElems {
			for i, kid := range potential.Kids {
				if kidElem, ok := kid.(StructKidElem); ok {
					// Find the actual element this refers to
					for _, e := range tm.structElems {
						if e == elem && kidElem.Ref.ID == 0 {
							potential.Kids[i] = StructKidElem{Ref: elem.Ref}
							elem.Parent = potential.Ref
						}
					}
				}
			}
		}
	}
}
