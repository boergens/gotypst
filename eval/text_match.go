package eval

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/boergens/gotypst/syntax"
)

// TextMatch represents a match of a pattern in text content.
type TextMatch struct {
	// Start is the byte offset of the match start in the collected text.
	Start int
	// End is the byte offset of the match end in the collected text.
	End int
	// Text is the matched text.
	Text string
	// Captures are the captured groups (for regex matches).
	Captures []string
}

// TextSpan represents a span of text content with source mapping.
type TextSpan struct {
	// Text is the text content.
	Text string
	// Start is the byte offset in the collected text.
	Start int
	// End is the byte offset in the collected text.
	End int
	// Element is the source content element.
	Element ContentElement
	// Index is the index of the element in the parent content.
	Index int
}

// CollectedText holds the result of collecting text from content.
type CollectedText struct {
	// Text is the merged text content.
	Text string
	// Spans maps regions of the text back to source elements.
	Spans []TextSpan
}

// CollectText collects all text from content elements, merging adjacent
// text elements and collapsing spaces according to Typst rules.
//
// Space collapsing rules:
// - Multiple consecutive spaces collapse to a single space
// - Spaces at the start/end of content are preserved
// - Linebreaks and parbreaks do NOT contribute to the collected text
//   (they are structural, not textual)
func CollectText(content Content) *CollectedText {
	var builder strings.Builder
	var spans []TextSpan
	offset := 0

	for i, elem := range content.Elements {
		text := extractText(elem)
		if text == "" {
			continue
		}

		// Track the span
		startOffset := offset
		builder.WriteString(text)
		offset += len(text)

		spans = append(spans, TextSpan{
			Text:    text,
			Start:   startOffset,
			End:     offset,
			Element: elem,
			Index:   i,
		})
	}

	return &CollectedText{
		Text:  builder.String(),
		Spans: spans,
	}
}

// CollectTextWithSpaceCollapsing collects text while collapsing whitespace.
// This is used for matching patterns where whitespace should be normalized.
func CollectTextWithSpaceCollapsing(content Content) *CollectedText {
	var builder strings.Builder
	var spans []TextSpan
	offset := 0
	lastWasSpace := false

	for i, elem := range content.Elements {
		text := extractText(elem)
		if text == "" {
			continue
		}

		// Collapse spaces
		var collapsedText strings.Builder
		for _, r := range text {
			if unicode.IsSpace(r) {
				if !lastWasSpace {
					collapsedText.WriteRune(' ')
					lastWasSpace = true
				}
			} else {
				collapsedText.WriteRune(r)
				lastWasSpace = false
			}
		}

		collapsed := collapsedText.String()
		if collapsed == "" {
			continue
		}

		startOffset := offset
		builder.WriteString(collapsed)
		offset += len(collapsed)

		spans = append(spans, TextSpan{
			Text:    collapsed,
			Start:   startOffset,
			End:     offset,
			Element: elem,
			Index:   i,
		})
	}

	return &CollectedText{
		Text:  builder.String(),
		Spans: spans,
	}
}

// extractText extracts text content from a content element.
func extractText(elem ContentElement) string {
	switch e := elem.(type) {
	case *TextElement:
		return e.Text
	case *StrongElement:
		return CollectText(e.Content).Text
	case *EmphElement:
		return CollectText(e.Content).Text
	case *RawElement:
		return e.Text
	case *HeadingElement:
		return CollectText(e.Content).Text
	case *ParagraphElement:
		return CollectText(e.Body).Text
	default:
		// Other elements don't contribute text
		return ""
	}
}

// MatchTextSelector finds all matches of a TextSelector in content.
// Returns the matches sorted by position.
func MatchTextSelector(selector TextSelector, content Content) ([]TextMatch, error) {
	collected := CollectText(content)
	return matchInText(selector, collected.Text)
}

// MatchTextSelectorWithCollapsing finds matches with space collapsing.
func MatchTextSelectorWithCollapsing(selector TextSelector, content Content) ([]TextMatch, error) {
	collected := CollectTextWithSpaceCollapsing(content)
	return matchInText(selector, collected.Text)
}

// matchInText finds all matches of a pattern in a string.
func matchInText(selector TextSelector, text string) ([]TextMatch, error) {
	if selector.IsRegex {
		return matchRegex(selector.Text, text)
	}
	return matchLiteral(selector.Text, text), nil
}

// matchLiteral finds all literal string matches.
func matchLiteral(pattern, text string) []TextMatch {
	var matches []TextMatch

	if pattern == "" {
		return matches
	}

	idx := 0
	for {
		pos := strings.Index(text[idx:], pattern)
		if pos == -1 {
			break
		}
		absPos := idx + pos
		matches = append(matches, TextMatch{
			Start:    absPos,
			End:      absPos + len(pattern),
			Text:     pattern,
			Captures: nil,
		})
		idx = absPos + len(pattern)
	}

	return matches
}

// matchRegex finds all regex pattern matches.
func matchRegex(pattern, text string) ([]TextMatch, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, &RegexError{
			Pattern: pattern,
			Message: err.Error(),
		}
	}

	allMatches := re.FindAllStringSubmatchIndex(text, -1)
	var matches []TextMatch

	for _, m := range allMatches {
		if len(m) < 2 {
			continue
		}
		start, end := m[0], m[1]
		matchText := text[start:end]

		// Extract captures (groups beyond the full match)
		var captures []string
		for i := 2; i < len(m); i += 2 {
			if m[i] >= 0 && m[i+1] >= 0 {
				captures = append(captures, text[m[i]:m[i+1]])
			} else {
				captures = append(captures, "")
			}
		}

		matches = append(matches, TextMatch{
			Start:    start,
			End:      end,
			Text:     matchText,
			Captures: captures,
		})
	}

	return matches, nil
}

// RegexError is returned when a regex pattern is invalid.
type RegexError struct {
	Pattern string
	Message string
	Span    syntax.Span
}

func (e *RegexError) Error() string {
	return "invalid regex pattern `" + e.Pattern + "`: " + e.Message
}

// SplitContentByMatches splits content at match boundaries.
// Returns segments: non-match, match, non-match, match, ..., non-match
// The matches are returned in their original positions within the result.
func SplitContentByMatches(content Content, matches []TextMatch) []ContentSegment {
	if len(matches) == 0 {
		return []ContentSegment{{
			Content:  content,
			IsMatch:  false,
			MatchIdx: -1,
		}}
	}

	collected := CollectText(content)
	text := collected.Text
	var segments []ContentSegment

	pos := 0
	for i, match := range matches {
		// Add non-match segment before this match
		if match.Start > pos {
			seg := createSegmentFromText(text[pos:match.Start], content, collected.Spans, pos, match.Start)
			if len(seg.Content.Elements) > 0 {
				seg.IsMatch = false
				seg.MatchIdx = -1
				segments = append(segments, seg)
			}
		}

		// Add the match segment
		matchSeg := createSegmentFromText(text[match.Start:match.End], content, collected.Spans, match.Start, match.End)
		matchSeg.IsMatch = true
		matchSeg.MatchIdx = i
		matchSeg.Match = &matches[i]
		segments = append(segments, matchSeg)

		pos = match.End
	}

	// Add final non-match segment
	if pos < len(text) {
		seg := createSegmentFromText(text[pos:], content, collected.Spans, pos, len(text))
		if len(seg.Content.Elements) > 0 {
			seg.IsMatch = false
			seg.MatchIdx = -1
			segments = append(segments, seg)
		}
	}

	return segments
}

// ContentSegment represents a segment of content, either a match or non-match.
type ContentSegment struct {
	// Content is the content in this segment.
	Content Content
	// IsMatch indicates if this segment is a pattern match.
	IsMatch bool
	// MatchIdx is the index of the match in the original matches slice.
	// -1 for non-match segments.
	MatchIdx int
	// Match is the match data if IsMatch is true.
	Match *TextMatch
}

// createSegmentFromText creates a content segment from a text range.
func createSegmentFromText(text string, _ Content, spans []TextSpan, start, end int) ContentSegment {
	// Simple implementation: create a TextElement with the matched text
	// A more sophisticated implementation would preserve the original elements
	return ContentSegment{
		Content: Content{
			Elements: []ContentElement{&TextElement{Text: text}},
		},
	}
}

// FindSpansInRange finds all spans that overlap with a byte range.
func FindSpansInRange(spans []TextSpan, start, end int) []TextSpan {
	var result []TextSpan
	for _, span := range spans {
		if span.End > start && span.Start < end {
			result = append(result, span)
		}
	}
	return result
}
