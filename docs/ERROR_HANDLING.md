# Error Handling Design for GoTypst

This document describes the Go error handling patterns that translate Typst's Rust error system.

## Table of Contents

1. [Rust Patterns Overview](#rust-patterns-overview)
2. [Go Translations](#go-translations)
3. [Existing Implementations](#existing-implementations)
4. [Recommended Error Types](#recommended-error-types)
5. [Code Examples](#code-examples)
6. [Error Aggregation](#error-aggregation)

---

## Rust Patterns Overview

### SourceResult<T>

In Typst's Rust codebase, `SourceResult<T>` is the primary result type:

```rust
/// The result type for operations that can produce source diagnostics.
pub type SourceResult<T> = Result<T, EcoVec<SourceDiagnostic>>;
```

Key characteristics:
- Returns either a value `T` or a vector of diagnostics
- Diagnostics carry span information for source location
- Multiple diagnostics can be returned at once (not just one error)

### SourceDiagnostic

```rust
pub struct SourceDiagnostic {
    /// The severity of the diagnostic.
    pub severity: Severity,
    /// The span of the relevant source code.
    pub span: Span,
    /// The diagnostic message.
    pub message: EcoString,
    /// Additional hints about the diagnostic.
    pub hints: Vec<EcoString>,
    /// A stack trace leading up to the diagnostic.
    pub trace: Vec<Spanned<Tracepoint>>,
}

pub enum Severity {
    Error,
    Warning,
}
```

### Span in Errors

Typst errors carry `Span` values that encode:
- File ID (16 bits)
- Position within file (48 bits) - either numbered spans or raw byte ranges

This allows precise error reporting: "error at file X, position Y".

### Error Chaining

Rust uses the `?` operator with `anyhow::Context`:

```rust
fn process(world: &dyn World, id: FileId) -> SourceResult<Value> {
    let source = world.source(id).at(span)?;
    let content = parse(&source).at(span)?;
    Ok(content)
}
```

The `.at(span)` method attaches source location to errors.

---

## Go Translations

### SourceResult<T> → (T, []Diagnostic)

Go's multiple return values replace Rust's `Result`:

```go
func Process(world World, id FileID) (Value, []Diagnostic) {
    source, diags := world.Source(id)
    if diags != nil {
        return Value{}, diags
    }
    content, diags := Parse(source)
    if diags != nil {
        return Value{}, diags
    }
    return content, nil
}
```

For operations that can only fail (not warn), use standard `(T, error)`:

```go
func ReadFile(path string) ([]byte, error) {
    contents, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read %s: %w", path, err)
    }
    return contents, nil
}
```

### ? Operator → Explicit Checks

Rust's `?` becomes explicit error checks:

| Rust | Go |
|------|-----|
| `let x = foo()?;` | `x, err := foo(); if err != nil { return ..., err }` |
| `foo().at(span)?` | `x, err := foo(); if err != nil { return ..., WithSpan(err, span) }` |

### Error Wrapping with Context

Use `fmt.Errorf` with `%w` for error chains:

```go
func compile(source Source) (*Document, error) {
    ast, err := parse(source)
    if err != nil {
        return nil, fmt.Errorf("parsing %s: %w", source.ID.Path, err)
    }
    // ...
}
```

---

## Existing Implementations

### Span (syntax/span.go)

The `Span` type is already translated:

```go
type Span struct {
    bits uint64  // 16 bits file ID + 48 bits number
}

func (s Span) IsDetached() bool
func (s Span) Id() *FileId
func (s Span) Number() uint64
func (s Span) Range() (start, end int, ok bool)
```

### Spanned[T] (syntax/span.go)

Generic type pairing values with source locations:

```go
type Spanned[T any] struct {
    V    T
    Span Span
}

func NewSpanned[T any](v T, span Span) Spanned[T]
func SpannedDetached[T any](v T) Spanned[T]
```

### SyntaxError (syntax/node.go)

Syntax-specific errors with span and hints:

```go
type SyntaxError struct {
    Span    Span
    Message string
    Hints   []string
}

func (e *SyntaxError) Error() string { return e.Message }
func (e *SyntaxError) AddHint(hint string)
```

---

## Recommended Error Types

### Severity

Distinguish errors from warnings:

```go
// Severity indicates how serious a diagnostic is.
type Severity int

const (
    // SeverityError indicates a fatal error that prevents compilation.
    SeverityError Severity = iota
    // SeverityWarning indicates a non-fatal issue.
    SeverityWarning
)

func (s Severity) String() string {
    switch s {
    case SeverityError:
        return "error"
    case SeverityWarning:
        return "warning"
    default:
        return "unknown"
    }
}
```

### Diagnostic

The primary diagnostic type for all phases of compilation:

```go
// Diagnostic represents an error or warning with source location.
type Diagnostic struct {
    // Severity indicates error vs warning.
    Severity Severity
    // Span is the source location of the diagnostic.
    Span Span
    // Message is the diagnostic message.
    Message string
    // Hints provides additional guidance.
    Hints []string
    // Trace is a stack of locations leading to the diagnostic.
    Trace []Spanned[Tracepoint]
}

// Error implements the error interface.
func (d *Diagnostic) Error() string {
    return d.Message
}

// IsError returns true if this is an error (not a warning).
func (d *Diagnostic) IsError() bool {
    return d.Severity == SeverityError
}

// WithHint adds a hint and returns the diagnostic for chaining.
func (d *Diagnostic) WithHint(hint string) *Diagnostic {
    d.Hints = append(d.Hints, hint)
    return d
}

// WithTrace adds a trace entry and returns the diagnostic for chaining.
func (d *Diagnostic) WithTrace(span Span, tp Tracepoint) *Diagnostic {
    d.Trace = append(d.Trace, NewSpanned(tp, span))
    return d
}
```

### Tracepoint

For stack traces in evaluation errors:

```go
// Tracepoint represents a point in an evaluation trace.
type Tracepoint interface {
    isTracepoint()
    String() string
}

// CallTracepoint represents a function call in a trace.
type CallTracepoint struct {
    Name string
}

func (CallTracepoint) isTracepoint() {}
func (t CallTracepoint) String() string {
    return fmt.Sprintf("in call to %s", t.Name)
}

// ShowTracepoint represents a show rule application.
type ShowTracepoint struct {
    Selector string
}

func (ShowTracepoint) isTracepoint() {}
func (t ShowTracepoint) String() string {
    return fmt.Sprintf("in show rule for %s", t.Selector)
}

// ImportTracepoint represents an import in a trace.
type ImportTracepoint struct {
    Path string
}

func (ImportTracepoint) isTracepoint() {}
func (t ImportTracepoint) String() string {
    return fmt.Sprintf("in import of %s", t.Path)
}
```

### Diagnostic Constructors

```go
// NewError creates an error diagnostic.
func NewError(span Span, message string) *Diagnostic {
    return &Diagnostic{
        Severity: SeverityError,
        Span:     span,
        Message:  message,
    }
}

// NewWarning creates a warning diagnostic.
func NewWarning(span Span, message string) *Diagnostic {
    return &Diagnostic{
        Severity: SeverityWarning,
        Span:     span,
        Message:  message,
    }
}

// NewErrorf creates an error with a formatted message.
func NewErrorf(span Span, format string, args ...any) *Diagnostic {
    return NewError(span, fmt.Sprintf(format, args...))
}

// NewWarningf creates a warning with a formatted message.
func NewWarningf(span Span, format string, args ...any) *Diagnostic {
    return NewWarning(span, fmt.Sprintf(format, args...))
}
```

### SpannedError Interface

For errors that carry span information:

```go
// SpannedError is an error that has a source location.
type SpannedError interface {
    error
    GetSpan() Span
}

// Ensure Diagnostic implements SpannedError.
func (d *Diagnostic) GetSpan() Span {
    return d.Span
}

// WithSpan wraps any error with a span, creating a Diagnostic.
func WithSpan(err error, span Span) *Diagnostic {
    if err == nil {
        return nil
    }
    // If already a diagnostic, update span if detached
    if d, ok := err.(*Diagnostic); ok {
        if d.Span.IsDetached() {
            d.Span = span
        }
        return d
    }
    // Wrap in new diagnostic
    return NewError(span, err.Error())
}
```

---

## Code Examples

### Basic Error Creation

```go
// Syntax error with span
err := NewError(span, "expected identifier")

// With a hint
err := NewError(span, "unknown variable").
    WithHint("did you mean 'content'?")

// Warning
warn := NewWarning(span, "unused variable")
```

### Attaching Spans to Errors

```go
func lookup(scope *Scope, name string, span Span) (Value, error) {
    val, ok := scope.Get(name)
    if !ok {
        return nil, NewErrorf(span, "unknown variable: %s", name)
    }
    return val, nil
}
```

### Error Propagation with Context

```go
func evalImport(world World, path string, span Span) (Module, error) {
    source, err := world.Source(FileID{Path: path})
    if err != nil {
        return Module{}, WithSpan(err, span).WithTrace(span, ImportTracepoint{Path: path})
    }
    return compile(source)
}
```

### Converting SyntaxError to Diagnostic

```go
func SyntaxErrorToDiagnostic(e *SyntaxError) *Diagnostic {
    return &Diagnostic{
        Severity: SeverityError,
        Span:     e.Span,
        Message:  e.Message,
        Hints:    e.Hints,
    }
}

func SyntaxErrorsToDiagnostics(errs []*SyntaxError) []*Diagnostic {
    diags := make([]*Diagnostic, len(errs))
    for i, e := range errs {
        diags[i] = SyntaxErrorToDiagnostic(e)
    }
    return diags
}
```

---

## Error Aggregation

### Diagnostics Collector

For collecting multiple errors/warnings during compilation:

```go
// Diagnostics collects diagnostics during compilation.
type Diagnostics struct {
    items []*Diagnostic
}

// NewDiagnostics creates an empty diagnostics collector.
func NewDiagnostics() *Diagnostics {
    return &Diagnostics{}
}

// Add adds a diagnostic to the collection.
func (d *Diagnostics) Add(diag *Diagnostic) {
    d.items = append(d.items, diag)
}

// AddError adds an error diagnostic.
func (d *Diagnostics) AddError(span Span, message string) {
    d.Add(NewError(span, message))
}

// AddWarning adds a warning diagnostic.
func (d *Diagnostics) AddWarning(span Span, message string) {
    d.Add(NewWarning(span, message))
}

// Extend adds all diagnostics from another collection.
func (d *Diagnostics) Extend(other *Diagnostics) {
    d.items = append(d.items, other.items...)
}

// HasErrors returns true if any error diagnostics exist.
func (d *Diagnostics) HasErrors() bool {
    for _, diag := range d.items {
        if diag.IsError() {
            return true
        }
    }
    return false
}

// Errors returns only the error diagnostics.
func (d *Diagnostics) Errors() []*Diagnostic {
    var errs []*Diagnostic
    for _, diag := range d.items {
        if diag.IsError() {
            errs = append(errs, diag)
        }
    }
    return errs
}

// Warnings returns only the warning diagnostics.
func (d *Diagnostics) Warnings() []*Diagnostic {
    var warns []*Diagnostic
    for _, diag := range d.items {
        if !diag.IsError() {
            warns = append(warns, diag)
        }
    }
    return warns
}

// All returns all diagnostics.
func (d *Diagnostics) All() []*Diagnostic {
    return d.items
}

// IsEmpty returns true if no diagnostics exist.
func (d *Diagnostics) IsEmpty() bool {
    return len(d.items) == 0
}

// Len returns the number of diagnostics.
func (d *Diagnostics) Len() int {
    return len(d.items)
}
```

### Usage Pattern

```go
func compile(world World, source Source) (*Document, *Diagnostics) {
    diags := NewDiagnostics()

    // Parse phase collects syntax errors
    ast, syntaxErrs := parse(source)
    for _, e := range syntaxErrs {
        diags.Add(SyntaxErrorToDiagnostic(e))
    }

    // Continue even with syntax errors to find more issues
    if diags.HasErrors() {
        return nil, diags
    }

    // Eval phase may add warnings and errors
    content, evalDiags := eval(world, ast)
    diags.Extend(evalDiags)

    if diags.HasErrors() {
        return nil, diags
    }

    // Layout phase
    doc, layoutDiags := layout(content)
    diags.Extend(layoutDiags)

    return doc, diags
}
```

### Returning Diagnostics as Error

When a function signature requires `error`:

```go
// Error implements the error interface for Diagnostics.
func (d *Diagnostics) Error() string {
    if len(d.items) == 0 {
        return "no diagnostics"
    }
    if len(d.items) == 1 {
        return d.items[0].Error()
    }
    var sb strings.Builder
    for i, diag := range d.items {
        if i > 0 {
            sb.WriteString("\n")
        }
        sb.WriteString(diag.Error())
    }
    return sb.String()
}

// ToError returns nil if no errors, otherwise returns self.
func (d *Diagnostics) ToError() error {
    if !d.HasErrors() {
        return nil
    }
    return d
}
```

---

## Integration with Span Types

The error system integrates with the existing span infrastructure:

### Span Resolution for Display

```go
// DiagnosticDisplay formats a diagnostic for user display.
type DiagnosticDisplay struct {
    Diagnostic *Diagnostic
    Source     *Source  // Source file for context
}

func (d DiagnosticDisplay) String() string {
    var sb strings.Builder

    // Severity and message
    fmt.Fprintf(&sb, "%s: %s\n", d.Diagnostic.Severity, d.Diagnostic.Message)

    // Source location (if span is not detached)
    if !d.Diagnostic.Span.IsDetached() {
        if start, end, ok := d.Diagnostic.Span.Range(); ok && d.Source != nil {
            // Show line/column from byte range
            line, col := lineColumn(d.Source.Text, start)
            fmt.Fprintf(&sb, "  --> %s:%d:%d\n", d.Source.ID.Path, line, col)
            // Could also show source snippet here
        }
    }

    // Hints
    for _, hint := range d.Diagnostic.Hints {
        fmt.Fprintf(&sb, "  hint: %s\n", hint)
    }

    // Trace
    for _, tp := range d.Diagnostic.Trace {
        fmt.Fprintf(&sb, "  at: %s\n", tp.V)
    }

    return sb.String()
}

func lineColumn(text string, byteOffset int) (line, col int) {
    line = 1
    col = 1
    for i, r := range text {
        if i >= byteOffset {
            break
        }
        if r == '\n' {
            line++
            col = 1
        } else {
            col++
        }
    }
    return
}
```

---

## Summary

| Rust Pattern | Go Equivalent |
|--------------|---------------|
| `SourceResult<T>` | `(T, *Diagnostics)` or `(T, error)` |
| `SourceDiagnostic` | `*Diagnostic` |
| `Severity::Error/Warning` | `SeverityError/SeverityWarning` |
| `span.at(err)` | `WithSpan(err, span)` |
| `err.with_hints(...)` | `diag.WithHint(...)` |
| `EcoVec<SourceDiagnostic>` | `*Diagnostics` |
| `trace.push(span, tracepoint)` | `diag.WithTrace(span, tp)` |

The key principles:
1. Use `*Diagnostic` for errors with source locations
2. Use `*Diagnostics` to collect multiple errors/warnings
3. Use `WithSpan()` to attach locations to errors
4. Use hints for user-friendly suggestions
5. Use traces for evaluation stack context
