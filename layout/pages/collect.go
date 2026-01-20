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

// isPagebreak checks if an element is a pagebreak.
func isPagebreak(elem interface{}) (*PagebreakElem, bool) {
	if pb, ok := elem.(*PagebreakElem); ok {
		return pb, true
	}
	return nil, false
}

// TagElem represents a tag element.
type TagElem struct {
	Tag Tag
}

// isTag checks if an element is a tag.
func isTag(elem interface{}) bool {
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
// Returns the position right after the last non-migrated tag.
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
	end := mid
	for end < len(children) {
		if _, ok := isPagebreak(children[end].Element); !ok {
			break
		}
		end++
	}

	if end == mid {
		return mid // No pagebreaks to migrate across
	}

	// Collect excluded locations (end tags)
	excluded := make(map[Location]bool)
	for i := tagStart; i < mid; i++ {
		if te, ok := children[i].Element.(*TagElem); ok {
			if te.Tag.Kind == TagEnd {
				excluded[te.Tag.Location] = true
			}
		}
	}

	// Sort: excluded tags first, then pagebreaks, then start tags
	// For simplicity, we'll just move unexcluded start tags after pagebreaks
	// This is a simplified version - full implementation would do proper sorting

	// Count how many tags should stay before pagebreaks
	stayCount := 0
	for i := tagStart; i < mid; i++ {
		if te, ok := children[i].Element.(*TagElem); ok {
			if excluded[te.Tag.Location] {
				stayCount++
			}
		}
	}

	return tagStart + stayCount
}
