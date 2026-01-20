package pdf

import (
	"strings"
	"testing"

	"github.com/boergens/gotypst/layout/pages"
)

func TestTagManager_AllocMCID(t *testing.T) {
	tm := NewTagManager()

	// First MCID should be 0
	mcid1 := tm.AllocMCID()
	if mcid1 != 0 {
		t.Errorf("First MCID = %d, want 0", mcid1)
	}

	// Second MCID should be 1
	mcid2 := tm.AllocMCID()
	if mcid2 != 1 {
		t.Errorf("Second MCID = %d, want 1", mcid2)
	}
}

func TestTagManager_BeginEndTag(t *testing.T) {
	tm := NewTagManager()

	// Root element should already be active
	if tm.CurrentElement() == nil {
		t.Error("Expected root element to be active")
	}
	if tm.CurrentElement().Role != RoleDocument {
		t.Errorf("Root element role = %s, want Document", tm.CurrentElement().Role)
	}

	// Begin a paragraph tag
	elem := tm.BeginTag(RoleP)
	if elem.Role != RoleP {
		t.Errorf("Element role = %s, want P", elem.Role)
	}
	if tm.CurrentElement() != elem {
		t.Error("Current element should be the new paragraph")
	}

	// End the tag
	tm.EndTag()
	if tm.CurrentElement().Role != RoleDocument {
		t.Errorf("After EndTag, current element role = %s, want Document", tm.CurrentElement().Role)
	}
}

func TestTagManager_ResolveRole(t *testing.T) {
	tm := NewTagManager()

	// Standard roles should pass through
	if r := tm.ResolveRole("P"); r != RoleP {
		t.Errorf("ResolveRole(P) = %s, want P", r)
	}
	if r := tm.ResolveRole("H1"); r != RoleH1 {
		t.Errorf("ResolveRole(H1) = %s, want H1", r)
	}

	// Unknown roles default to Span
	if r := tm.ResolveRole("Unknown"); r != RoleSpan {
		t.Errorf("ResolveRole(Unknown) = %s, want Span", r)
	}

	// Custom role mapping
	tm.MapRole("CustomPara", RoleP)
	if r := tm.ResolveRole("CustomPara"); r != RoleP {
		t.Errorf("ResolveRole(CustomPara) = %s, want P", r)
	}
}

func TestTagManager_HasTags(t *testing.T) {
	tm := NewTagManager()

	// Fresh manager only has root element
	if tm.HasTags() {
		t.Error("Fresh TagManager should not have tags (only root)")
	}

	// After beginning a tag, HasTags should be true
	tm.BeginTag(RoleP)
	if !tm.HasTags() {
		t.Error("After BeginTag, HasTags should be true")
	}
}

func TestContentStream_MarkedContent(t *testing.T) {
	cs := NewContentStream()

	cs.BeginMarkedContentWithProps("P", 0)
	cs.ShowText("Hello")
	cs.EndMarkedContent()

	content := cs.String()

	// Should contain BDC with role and MCID
	if !strings.Contains(content, "/P <</MCID 0>> BDC") {
		t.Errorf("Content stream missing BDC operator: %s", content)
	}

	// Should contain EMC
	if !strings.Contains(content, "EMC") {
		t.Errorf("Content stream missing EMC operator: %s", content)
	}
}

func TestContentStream_Artifact(t *testing.T) {
	cs := NewContentStream()

	cs.BeginArtifact("Pagination")
	cs.ShowText("Page 1")
	cs.EndArtifact()

	content := cs.String()

	// Should contain artifact marker
	if !strings.Contains(content, "/Artifact") {
		t.Errorf("Content stream missing Artifact marker: %s", content)
	}
	if !strings.Contains(content, "Pagination") {
		t.Errorf("Content stream missing artifact type: %s", content)
	}
}

func TestTaggedWriter(t *testing.T) {
	w := NewTaggedWriter()

	if !w.tagged {
		t.Error("NewTaggedWriter should set tagged = true")
	}
	if w.tagManager == nil {
		t.Error("NewTaggedWriter should initialize tagManager")
	}
}

func TestWriter_EnableTagging(t *testing.T) {
	w := NewWriter()

	if w.tagged {
		t.Error("NewWriter should not be tagged by default")
	}

	w.EnableTagging()

	if !w.tagged {
		t.Error("After EnableTagging, tagged should be true")
	}
	if w.tagManager == nil {
		t.Error("After EnableTagging, tagManager should be initialized")
	}
}

func TestTagManager_ProcessTag(t *testing.T) {
	tm := NewTagManager()
	tm.SetCurrentPage(0)
	tm.SetPageRef(0, Ref{ID: 5, Gen: 0})

	// Process a start tag
	tag := &pages.Tag{
		Kind: pages.TagStart,
		Elem: &pages.CounterUpdateElem{
			Key:    pages.CounterKeyHeading,
			Update: pages.CounterUpdateStep{},
		},
	}

	role, mcid, isStart := tm.ProcessTag(tag)
	if !isStart {
		t.Error("Start tag should return isStart=true")
	}
	if role != RoleH {
		t.Errorf("Heading tag role = %s, want H", role)
	}
	if mcid != 0 {
		t.Errorf("First MCID = %d, want 0", mcid)
	}

	// Process an end tag
	endTag := &pages.Tag{
		Kind: pages.TagEnd,
	}
	_, _, isStart = tm.ProcessTag(endTag)
	if isStart {
		t.Error("End tag should return isStart=false")
	}
}

func TestTagManager_BuildParentTree(t *testing.T) {
	tm := NewTagManager()
	tm.SetCurrentPage(0)
	tm.SetPageRef(0, Ref{ID: 5, Gen: 0})

	// Add some marked content
	elem := tm.BeginTag(RoleP)
	elem.Ref = Ref{ID: 10, Gen: 0}
	mcid := tm.AllocMCID()
	tm.AddMarkedContent(mcid, RoleP)

	parentTree := tm.BuildParentTree()
	if len(parentTree) == 0 {
		t.Error("Parent tree should not be empty")
	}
	if _, ok := parentTree[mcid]; !ok {
		t.Errorf("Parent tree should contain MCID %d", mcid)
	}
}
