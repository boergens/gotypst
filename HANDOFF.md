# gotypst Handoff Document

**Repo:** https://github.com/boergens/gotypst
**Date:** 2026-01-21
**Status:** Active development, Phase 4 (PDF output) in progress

## Project Overview

gotypst is a Go implementation of Typst (the typesetting system). Development has been using Gas Town for automated task dispatch, but this document provides everything needed to continue without it.

## Current State

### What's Merged to Main

Recent work merged to `main` branch:
- Full standard library wired up in CLI
- Set rule evaluation fixes
- `image()` element for embedding images
- `link()` element for hyperlinks
- `table()` element for grid-based tables
- `lorem()` function for placeholder text
- Font subsetting and CID font handling for PDF
- Headers and footers with page numbering
- MathFragment types for math layout
- `box()`, `block()`, `pad()` layout elements
- `columns()` layout element
- `list()` and `enum()` elements with nesting
- `heading()` element and type constructors

### Branches Ready to Merge (6 total)

These branches have **completed work** ready for merge. Review and merge at your discretion:

| Branch | Task | Description |
|--------|------|-------------|
| `polecat/quartz/go-vwv2c@mkndznss` | P4-B | PDF page.rs - page layout to PDF |
| `polecat/obsidian/go-4okim@mkndz53k` | P4-A | PDF lib.rs - main entry point |
| `polecat/quartz/go-dt1qs@mknbl78r` | Review | table() against Typst Rust source |
| `polecat/obsidian/go-s9jzg@mknbkjaj` | Review | lorem() against Typst Rust source |
| `polecat/opal/go-cdrox@mknbnamt` | Review | set rule evaluation against Typst source |
| `polecat/jasper/go-ljzs6@mknblv08` | Review | link() against Typst Rust source |

**To merge a branch:**
```bash
git checkout main
git merge polecat/quartz/go-vwv2c@mkndznss
# Resolve any conflicts
git push origin main
```

### Phase 4 Tasks (PDF Output) - NOT STARTED

These tasks are defined but no work has been done on them yet. They're assigned to "polecats" (workers) that don't exist without Gas Town.

| Task ID | Module | Description |
|---------|--------|-------------|
| go-cf5rw | P4-C | PDF text.rs - text rendering and typography |
| go-v57bk | P4-D | PDF paint.rs - graphics rendering |
| go-uu8lu | P4-E | PDF shape.rs - geometric shapes |
| go-sl2hc | P4-F | PDF image.rs - image embedding |
| go-yik4j | P4-G | PDF link.rs - hyperlinks |
| go-m9vzl | P4-H | PDF outline.rs - bookmarks |
| go-0lho9 | P4-I | PDF metadata.rs - document metadata |
| go-usfop | P4-J | PDF convert.rs - format conversion |
| go-ts56q | P4-K | PDF attach.rs - file attachments |
| go-iadmd | P4-L | PDF util.rs - utilities |

**Source reference:** These translate from the Typst Rust source at `typst-pdf/src/*.rs`

**Approach:** Stay close to Rust structure, translate to idiomatic Go while preserving naming and logic.

## Branch Cleanup

There are ~250 polecat branches in the repo. Most are from completed work that's either:
- Already merged to main
- Superseded by later work
- Abandoned experiments

**Safe to delete** (after merging the 6 ready branches above):
```bash
# Delete all local polecat branches
git branch | grep polecat | xargs git branch -D

# Delete all remote polecat branches (careful!)
git branch -r | grep 'origin/polecat' | sed 's/origin\///' | xargs -I {} git push origin --delete {}
```

**Or keep them** for archaeology if you want to see how work evolved.

## Key Files

- `cmd/gotypst/` - CLI entry point
- `pkg/` - Core packages
  - `pkg/eval/` - Evaluation/runtime
  - `pkg/layout/` - Layout engine
  - `pkg/pdf/` - PDF generation (Phase 4 target)
  - `pkg/stdlib/` - Standard library functions
  - `pkg/syntax/` - Parser/AST

## What You'll Miss Without Gas Town

1. **Automated task dispatch** - You'll manually assign work
2. **Merge conflict resolution** - Refinery auto-handled conflicts
3. **Progress tracking** - Beads tracked issue status
4. **Parallel workers** - Multiple polecats worked simultaneously

## Recommended Next Steps

1. **Merge the 6 ready branches** (P4-A, P4-B, and 4 review tasks)
2. **Decide on Phase 4** - Continue PDF work or focus elsewhere
3. **Clean up branches** - Delete polecat branches after review
4. **Use GitHub Issues** - For tracking if desired

## Questions?

The original work was coordinated by Gas Town mayor. Reach out to the repo owner for context on specific decisions.
