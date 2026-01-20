package pages

// Item represents an item in page layout.
type Item interface {
	isItem()
}

// RunItem represents a page run containing content.
// All runs will be laid out in parallel.
type RunItem struct {
	Children []Pair
	Initial  StyleChain
	Locator  Locator
}

func (RunItem) isItem() {}

// TagsItem represents tags between pages.
// These will be prepended to the first start of the next page,
// or appended at the very end of the final page if there is no next page.
type TagsItem struct {
	Children []Pair
}

func (TagsItem) isItem() {}

// ParityItem represents an instruction to possibly add a page
// to bring the page number parity to the desired state.
// Can only be done at the end, sequentially, because it requires
// knowledge of the concrete page number.
type ParityItem struct {
	Parity  Parity
	Initial StyleChain
	Locator Locator
}

func (ParityItem) isItem() {}

// PagebreakElem represents a pagebreak element.
type PagebreakElem struct {
	// Weak indicates if this is a weak pagebreak.
	Weak bool
	// To specifies desired page parity after the break.
	To *Parity
	// Boundary indicates if this is a style boundary.
	Boundary bool
}

func (*PagebreakElem) isContentElement() {}

// isPagebreak checks if an element is a pagebreak.
func isPagebreak(elem ContentElement) (*PagebreakElem, bool) {
	if pb, ok := elem.(*PagebreakElem); ok {
		return pb, true
	}
	return nil, false
}

// TagElem represents a tag element.
type TagElem struct {
	Tag Tag
}

func (TagElem) isContentElement() {}

// isTag checks if an element is a tag.
func isTag(elem ContentElement) bool {
	_, ok := elem.(*TagElem)
	return ok
}

// Collect slices up children into logical parts, processing styles
// and handling things like tags and weak pagebreaks.
func Collect(children []Pair, locator *SplitLocator, initial StyleChain) []Item {
	var items []Item
	stagedEmptyPage := true
	idx := 0

	for idx < len(children) {
		elem := children[idx].Element
		styles := children[idx].Styles

		if pb, ok := isPagebreak(elem); ok {
			strong := !pb.Weak
			if strong && stagedEmptyPage {
				loc := locator.Next(nil)
				items = append(items, RunItem{
					Children: nil,
					Initial:  initial,
					Locator:  loc,
				})
			}

			if pb.To != nil {
				loc := locator.Next(nil)
				items = append(items, ParityItem{
					Parity:  *pb.To,
					Initial: styles,
					Locator: loc,
				})
			}

			if !pb.Boundary {
				initial = styles
			}

			if strong {
				stagedEmptyPage = true
			}
			idx++
		} else {
			// Find the end of the non-pagebreak group
			end := idx
			for end < len(children) {
				if _, ok := isPagebreak(children[end].Element); ok {
					break
				}
				end++
			}

			// Migrate unterminated tags
			end = migrateUnterminatedTags(children, idx, end)
			if end == idx {
				continue
			}

			group := children[idx:end]
			idx = end

			// Check if the group is all tags
			allTags := true
			for _, pair := range group {
				if !isTag(pair.Element) {
					allTags = false
					break
				}
			}

			// Check if remaining items are all boundary pagebreaks
			allBoundary := true
			for _, pair := range children[idx:] {
				if pb, ok := isPagebreak(pair.Element); ok {
					if !pb.Boundary {
						allBoundary = false
						break
					}
				} else {
					allBoundary = false
					break
				}
			}

			if allTags && !(stagedEmptyPage && allBoundary) {
				items = append(items, TagsItem{Children: group})
				continue
			}

			loc := locator.Next(nil)
			items = append(items, RunItem{
				Children: group,
				Initial:  initial,
				Locator:  loc,
			})
			stagedEmptyPage = false
		}
	}

	if stagedEmptyPage {
		items = append(items, RunItem{
			Children: nil,
			Initial:  initial,
			Locator:  locator.Next(nil),
		})
	}

	return items
}

// migrateUnterminatedTags migrates trailing start tags without accompanying
// end tags from before a pagebreak to after it.
//
// This function reorders elements in the children slice so that:
// 1. End tags and their matching start tags stay before pagebreaks
// 2. Pagebreaks remain in their relative position
// 3. Unterminated start tags move after the pagebreaks
//
// Returns the position right after the last tag that should stay before pagebreaks.
func migrateUnterminatedTags(children []Pair, start, mid int) int {
	if mid <= start {
		return mid
	}

	// Find trailing tags before mid
	tagStart := mid
	for tagStart > start {
		if !isTag(children[tagStart-1].Element) {
			break
		}
		tagStart--
	}

	if tagStart == mid {
		return mid // No trailing tags
	}

	// Find pagebreaks after mid
	pbEnd := mid
	for pbEnd < len(children) {
		if _, ok := isPagebreak(children[pbEnd].Element); !ok {
			break
		}
		pbEnd++
	}

	if pbEnd == mid {
		return mid // No pagebreaks to migrate across
	}

	// Collect locations that have end tags in this trailing group.
	// Tags with these locations are "terminated" and should stay.
	terminatedLocations := make(map[Location]bool)
	for i := tagStart; i < mid; i++ {
		if te, ok := children[i].Element.(*TagElem); ok {
			if te.Tag.Kind == TagEnd {
				terminatedLocations[te.Tag.Location] = true
			}
		}
	}

	// Partition trailing tags into those that stay and those that migrate.
	// We need to actually reorder the slice.
	var stayTags []Pair
	var migrateTags []Pair

	for i := tagStart; i < mid; i++ {
		pair := children[i]
		if te, ok := pair.Element.(*TagElem); ok {
			if te.Tag.Kind == TagEnd {
				// End tags always stay
				stayTags = append(stayTags, pair)
			} else if terminatedLocations[te.Tag.Location] {
				// Start tags with matching end tags stay
				stayTags = append(stayTags, pair)
			} else {
				// Unterminated start tags migrate
				migrateTags = append(migrateTags, pair)
			}
		}
	}

	// If nothing to migrate, return original mid
	if len(migrateTags) == 0 {
		return mid
	}

	// Collect the pagebreaks
	var pagebreaks []Pair
	for i := mid; i < pbEnd; i++ {
		pagebreaks = append(pagebreaks, children[i])
	}

	// Reorder: stayTags, then pagebreaks, then migrateTags
	// Write back to children slice starting at tagStart
	writeIdx := tagStart
	for _, p := range stayTags {
		children[writeIdx] = p
		writeIdx++
	}
	for _, p := range pagebreaks {
		children[writeIdx] = p
		writeIdx++
	}
	for _, p := range migrateTags {
		children[writeIdx] = p
		writeIdx++
	}

	// Return position after the tags that stayed (before pagebreaks)
	return tagStart + len(stayTags)
}
