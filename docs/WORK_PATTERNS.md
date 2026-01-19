# Polecat Work Patterns for GoTypst

This document describes the work patterns and conventions for polecat agents working on the GoTypst project.

## Core Principle: Small, Focused Beads

**Problem**: Long-running polecats hit context limits, compact, and produce lower quality work.

**Solution**: Break work into focused beads that can be completed with fresh context.

### Why This Matters

- Polecats are ephemeral workers - one task, then nuked
- Context accumulation leads to degraded performance
- Fresh context = better reasoning = higher quality output
- Parallel execution becomes possible with independent beads

## Work Patterns

### Pattern 1: Splitter Tasks

When facing complex work, use a splitter bead to decompose it:

```
Complex work → Splitter bead → Creates N focused beads
```

**How it works:**
1. Splitter bead analyzes the problem
2. Creates multiple focused sub-beads
3. Each sub-bead has clear, narrow scope
4. Fresh polecat tackles each with full context
5. Work proceeds in parallel when possible

**Example:**
```
"Implement Typst math rendering" →
  - "Implement math parser"
  - "Implement math layout engine"
  - "Implement math symbol rendering"
  - "Add math test coverage"
```

### Pattern 2: Review Chains

For implementations requiring thorough review:

```
Implementation → Split into review beads → Master reviewer
```

**Review dimensions:**
- **Correctness review**: Does it do what it claims?
- **Test coverage review**: Are edge cases tested?
- **Integration review**: Does it fit with existing code?
- **Performance review**: Any obvious bottlenecks?

**How it works:**
1. Implementation completes
2. Spawn focused review beads for each dimension
3. Each reviewer examines only their concern
4. Master synthesizes findings into actionable items

### Pattern 3: Research → Implementation

**Never implement in the same bead as research.**

```
Research bead → docs/analysis.md
Implementation bead reads docs, writes code
```

**Why separate?**
- Research accumulates context rapidly
- Implementation needs fresh context for code quality
- Documentation serves as handoff artifact
- Future polecats can reference research without re-doing it

**Example workflow:**
1. Research bead explores Typst's layout algorithm
2. Produces `docs/layout-analysis.md`
3. Implementation bead reads analysis, writes Go code
4. Clean context, clear requirements

## Guidelines

### Target Execution Time

- **Target**: <15 minutes per polecat
- If approaching this limit, consider splitting
- Compaction degrades quality - avoid it when possible

### When to Split

Ask yourself:
- "Does this task have multiple distinct phases?"
- "Am I accumulating context I won't need later?"
- "Could parallel execution speed this up?"

If yes to any: file sub-beads instead.

### Bead Sizing Heuristic

**Prefer 3 small beads over 1 large bead**

| Approach | Context Quality | Parallelism | Risk |
|----------|----------------|-------------|------|
| 1 large bead | Degrades over time | None | High (all-or-nothing) |
| 3 small beads | Fresh each time | Possible | Low (isolated failures) |

### Documentation as Handoff

Every bead should consider:
- What does the next polecat need to know?
- What did I discover that isn't in the code?
- What decisions did I make and why?

Capture this in docs or comments - your context dies when you complete.

## Anti-Patterns

### The Kitchen Sink Bead

**Bad**: "Implement feature X, write tests, update docs, refactor related code"

**Good**: Split into focused beads:
- "Implement feature X core logic"
- "Add test coverage for feature X"
- "Update docs for feature X"
- "Refactor Y to integrate with feature X"

### Research While Implementing

**Bad**: Exploring the codebase while writing new code in the same session.

**Good**: Research first, document findings, fresh polecat implements.

### Unbounded Exploration

**Bad**: "Investigate and fix all performance issues"

**Good**:
- "Profile application and identify top 3 bottlenecks" (produces report)
- Then spawn specific fix beads for each finding

## Integration with Gas Town

### Hook Protocol

Work on your hook triggers immediate execution. The hook is your assignment.

### Molecule Workflow

For complex work, molecules provide structured execution:
1. Attach molecule to pinned bead
2. Work through defined steps
3. Squash into digest on completion

### Handoff Pattern

If context is getting heavy mid-task:
1. Document current state
2. `gt handoff` to fresh session
3. New polecat continues with clean context

## Summary

1. **Keep beads small** - <15 min execution target
2. **Separate concerns** - research vs implementation, different review dimensions
3. **Fresh context = quality** - split before compaction
4. **Document for handoff** - your context dies, your docs live
5. **Prefer parallel** - 3 small beads beat 1 large bead
