// Package flow provides block-level flow layout for GoTypst.
//
// This package handles the distribution of content into regions, including:
// - Block-level flow layout
// - Content collection and preprocessing
// - Frame composition with floats and footnotes
// - Content distribution across multiple regions
//
// The flow layout pipeline:
// 1. Collect children into preprocessed structures
// 2. Compose frames by distributing content into regions
// 3. Handle floats and footnotes as out-of-flow elements
// 4. Apply sticky blocks and spacing logic
package flow
