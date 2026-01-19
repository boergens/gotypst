# Rust to Go Translation Guide

This guide covers common patterns for translating Rust code to idiomatic Go.

## Table of Contents

1. [Enums with Data](#enums-with-data)
2. [Result<T, E>](#resultt-e)
3. [Option<T>](#optiont)
4. [Traits](#traits)
5. [Impl Blocks](#impl-blocks)
6. [Pattern Matching](#pattern-matching)
7. [Ownership and Borrowing](#ownership-and-borrowing)
8. [Macros](#macros)
9. [Error Handling Patterns](#error-handling-patterns)
10. [Naming Conventions](#naming-conventions)

---

## Enums with Data

Rust enums with associated data translate to Go using an interface plus concrete types.

### Rust

```rust
enum Message {
    Quit,
    Move { x: i32, y: i32 },
    Write(String),
    ChangeColor(u8, u8, u8),
}

fn handle_message(msg: Message) {
    match msg {
        Message::Quit => println!("Quit"),
        Message::Move { x, y } => println!("Move to ({}, {})", x, y),
        Message::Write(text) => println!("Write: {}", text),
        Message::ChangeColor(r, g, b) => println!("Color: {}, {}, {}", r, g, b),
    }
}
```

### Go

```go
// Define a sealed interface with an unexported method
type Message interface {
    isMessage() // unexported method prevents external implementations
}

// Concrete types implement the interface
type QuitMessage struct{}

func (QuitMessage) isMessage() {}

type MoveMessage struct {
    X, Y int
}

func (MoveMessage) isMessage() {}

type WriteMessage struct {
    Text string
}

func (WriteMessage) isMessage() {}

type ChangeColorMessage struct {
    R, G, B uint8
}

func (ChangeColorMessage) isMessage() {}

func handleMessage(msg Message) {
    switch m := msg.(type) {
    case QuitMessage:
        fmt.Println("Quit")
    case MoveMessage:
        fmt.Printf("Move to (%d, %d)\n", m.X, m.Y)
    case WriteMessage:
        fmt.Printf("Write: %s\n", m.Text)
    case ChangeColorMessage:
        fmt.Printf("Color: %d, %d, %d\n", m.R, m.G, m.B)
    }
}
```

---

## Result<T, E>

Rust's `Result<T, E>` maps to Go's multiple return values with `(T, error)`.

### Rust

```rust
use std::fs::File;
use std::io::{self, Read};

fn read_file(path: &str) -> Result<String, io::Error> {
    let mut file = File::open(path)?;
    let mut contents = String::new();
    file.read_to_string(&mut contents)?;
    Ok(contents)
}

fn main() {
    match read_file("config.txt") {
        Ok(contents) => println!("File contents: {}", contents),
        Err(e) => eprintln!("Error reading file: {}", e),
    }
}
```

### Go

```go
func readFile(path string) (string, error) {
    contents, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }
    return string(contents), nil
}

func main() {
    contents, err := readFile("config.txt")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
        return
    }
    fmt.Printf("File contents: %s\n", contents)
}
```

### The `?` Operator

Rust's `?` operator for early returns translates to explicit error checking:

**Rust:**
```rust
fn process() -> Result<Value, Error> {
    let a = step_one()?;
    let b = step_two(a)?;
    let c = step_three(b)?;
    Ok(c)
}
```

**Go:**
```go
func process() (Value, error) {
    a, err := stepOne()
    if err != nil {
        return Value{}, err
    }
    b, err := stepTwo(a)
    if err != nil {
        return Value{}, err
    }
    c, err := stepThree(b)
    if err != nil {
        return Value{}, err
    }
    return c, nil
}
```

---

## Option<T>

Rust's `Option<T>` can be translated to Go using either pointers or the comma-ok idiom.

### Using Pointers

**Rust:**
```rust
struct User {
    name: String,
    email: Option<String>,
}

fn get_email(user: &User) -> Option<&str> {
    user.email.as_deref()
}
```

**Go:**
```go
type User struct {
    Name  string
    Email *string // nil represents None
}

func getEmail(user *User) *string {
    return user.Email
}

// Usage
func printEmail(user *User) {
    if email := getEmail(user); email != nil {
        fmt.Printf("Email: %s\n", *email)
    } else {
        fmt.Println("No email provided")
    }
}
```

### Using the Comma-Ok Idiom

For map lookups and type assertions, Go has a built-in comma-ok pattern:

**Rust:**
```rust
use std::collections::HashMap;

fn get_value(map: &HashMap<String, i32>, key: &str) -> Option<i32> {
    map.get(key).copied()
}
```

**Go:**
```go
func getValue(m map[string]int, key string) (int, bool) {
    val, ok := m[key]
    return val, ok
}

// Usage
if val, ok := getValue(myMap, "key"); ok {
    fmt.Printf("Found: %d\n", val)
} else {
    fmt.Println("Not found")
}
```

### Custom Optional Type (when needed)

For complex scenarios, define an explicit optional type:

```go
type Optional[T any] struct {
    value T
    valid bool
}

func Some[T any](v T) Optional[T] {
    return Optional[T]{value: v, valid: true}
}

func None[T any]() Optional[T] {
    return Optional[T]{}
}

func (o Optional[T]) IsSome() bool {
    return o.valid
}

func (o Optional[T]) Unwrap() T {
    if !o.valid {
        panic("called Unwrap on None value")
    }
    return o.value
}

func (o Optional[T]) UnwrapOr(def T) T {
    if o.valid {
        return o.value
    }
    return def
}
```

---

## Traits

Rust traits translate directly to Go interfaces.

### Rust

```rust
trait Writer {
    fn write(&mut self, data: &[u8]) -> Result<usize, Error>;
    fn flush(&mut self) -> Result<(), Error>;
}

trait Reader {
    fn read(&mut self, buf: &mut [u8]) -> Result<usize, Error>;
}

// Trait bounds
fn copy<R: Reader, W: Writer>(reader: &mut R, writer: &mut W) -> Result<usize, Error> {
    let mut buf = [0u8; 1024];
    let mut total = 0;
    loop {
        let n = reader.read(&mut buf)?;
        if n == 0 {
            break;
        }
        writer.write(&buf[..n])?;
        total += n;
    }
    Ok(total)
}
```

### Go

```go
type Writer interface {
    Write(data []byte) (int, error)
    Flush() error
}

type Reader interface {
    Read(buf []byte) (int, error)
}

// Interface parameters instead of trait bounds
func copy(reader Reader, writer Writer) (int, error) {
    buf := make([]byte, 1024)
    total := 0
    for {
        n, err := reader.Read(buf)
        if err == io.EOF {
            break
        }
        if err != nil {
            return total, err
        }
        _, err = writer.Write(buf[:n])
        if err != nil {
            return total, err
        }
        total += n
    }
    return total, nil
}
```

### Default Trait Methods

Rust's default trait methods become separate functions or embedded types:

**Rust:**
```rust
trait Greeter {
    fn name(&self) -> &str;

    fn greet(&self) -> String {
        format!("Hello, {}!", self.name())
    }
}
```

**Go:**
```go
type Greeter interface {
    Name() string
}

// Default behavior as a function
func Greet(g Greeter) string {
    return fmt.Sprintf("Hello, %s!", g.Name())
}

// Or embed in a wrapper for method syntax
type GreeterWithDefaults struct {
    Greeter
}

func (g GreeterWithDefaults) Greet() string {
    return fmt.Sprintf("Hello, %s!", g.Name())
}
```

---

## Impl Blocks

Rust's `impl` blocks become methods defined on types in Go.

### Rust

```rust
struct Rectangle {
    width: u32,
    height: u32,
}

impl Rectangle {
    // Associated function (constructor)
    fn new(width: u32, height: u32) -> Self {
        Rectangle { width, height }
    }

    // Method with &self
    fn area(&self) -> u32 {
        self.width * self.height
    }

    // Method with &mut self
    fn scale(&mut self, factor: u32) {
        self.width *= factor;
        self.height *= factor;
    }

    // Associated function (not a method)
    fn square(size: u32) -> Self {
        Rectangle { width: size, height: size }
    }
}
```

### Go

```go
type Rectangle struct {
    Width  uint32
    Height uint32
}

// Constructor function (no receiver)
func NewRectangle(width, height uint32) Rectangle {
    return Rectangle{Width: width, Height: height}
}

// Method with value receiver (like &self)
func (r Rectangle) Area() uint32 {
    return r.Width * r.Height
}

// Method with pointer receiver (like &mut self)
func (r *Rectangle) Scale(factor uint32) {
    r.Width *= factor
    r.Height *= factor
}

// Another constructor function
func NewSquare(size uint32) Rectangle {
    return Rectangle{Width: size, Height: size}
}
```

---

## Pattern Matching

Rust's `match` translates to Go's `switch` statement, particularly type switches for enums.

### Basic Match

**Rust:**
```rust
fn describe_number(n: i32) -> &'static str {
    match n {
        0 => "zero",
        1 | 2 | 3 => "small",
        4..=10 => "medium",
        _ if n < 0 => "negative",
        _ => "large",
    }
}
```

**Go:**
```go
func describeNumber(n int) string {
    switch {
    case n == 0:
        return "zero"
    case n >= 1 && n <= 3:
        return "small"
    case n >= 4 && n <= 10:
        return "medium"
    case n < 0:
        return "negative"
    default:
        return "large"
    }
}
```

### Type Switch (for Enums)

**Rust:**
```rust
enum Shape {
    Circle { radius: f64 },
    Rectangle { width: f64, height: f64 },
    Triangle { base: f64, height: f64 },
}

fn area(shape: &Shape) -> f64 {
    match shape {
        Shape::Circle { radius } => std::f64::consts::PI * radius * radius,
        Shape::Rectangle { width, height } => width * height,
        Shape::Triangle { base, height } => 0.5 * base * height,
    }
}
```

**Go:**
```go
type Shape interface {
    isShape()
}

type Circle struct {
    Radius float64
}
func (Circle) isShape() {}

type Rectangle struct {
    Width, Height float64
}
func (Rectangle) isShape() {}

type Triangle struct {
    Base, Height float64
}
func (Triangle) isShape() {}

func area(shape Shape) float64 {
    switch s := shape.(type) {
    case Circle:
        return math.Pi * s.Radius * s.Radius
    case Rectangle:
        return s.Width * s.Height
    case Triangle:
        return 0.5 * s.Base * s.Height
    default:
        return 0
    }
}
```

### Destructuring in Match

**Rust:**
```rust
struct Point { x: i32, y: i32 }

fn classify_point(p: Point) -> &'static str {
    match p {
        Point { x: 0, y: 0 } => "origin",
        Point { x: 0, .. } => "on y-axis",
        Point { y: 0, .. } => "on x-axis",
        Point { x, y } if x == y => "diagonal",
        _ => "elsewhere",
    }
}
```

**Go:**
```go
type Point struct {
    X, Y int
}

func classifyPoint(p Point) string {
    switch {
    case p.X == 0 && p.Y == 0:
        return "origin"
    case p.X == 0:
        return "on y-axis"
    case p.Y == 0:
        return "on x-axis"
    case p.X == p.Y:
        return "diagonal"
    default:
        return "elsewhere"
    }
}
```

---

## Ownership and Borrowing

Go uses garbage collection, so Rust's ownership system has no direct equivalent. However, some patterns are worth noting.

### Pass by Value vs Pointer

**Rust:**
```rust
fn takes_ownership(s: String) { /* s is moved */ }
fn borrows(s: &String) { /* s is borrowed */ }
fn borrows_mut(s: &mut String) { /* s is mutably borrowed */ }
```

**Go:**
```go
func takesValue(s string) { /* s is copied (strings are immutable) */ }
func takesPointer(s *MyStruct) { /* can modify the struct */ }
func takesValue2(s MyStruct) { /* s is copied */ }
```

### Guidelines

- Small types (primitives, small structs): pass by value
- Large structs: pass by pointer to avoid copying
- When mutation is needed: pass by pointer
- Slices and maps: already reference types, pass by value

### Clone

**Rust:**
```rust
let original = expensive_struct.clone();
```

**Go:**
```go
// For simple structs, assignment copies
original := expensiveStruct

// For deep copies with pointers/slices, implement a Clone method
func (s *MyStruct) Clone() MyStruct {
    clone := *s
    clone.Slice = make([]int, len(s.Slice))
    copy(clone.Slice, s.Slice)
    return clone
}
```

---

## Macros

Rust macros translate to different Go patterns depending on their purpose.

### Declarative Macros → Functions or Generics

**Rust:**
```rust
macro_rules! vec {
    ( $( $x:expr ),* ) => {
        {
            let mut temp_vec = Vec::new();
            $(
                temp_vec.push($x);
            )*
            temp_vec
        }
    };
}

let v = vec![1, 2, 3];
```

**Go:**
```go
// Use variadic functions
func newSlice[T any](items ...T) []T {
    return items
}

v := newSlice(1, 2, 3)

// Or just use slice literals
v := []int{1, 2, 3}
```

### Derive Macros → Code Generation

**Rust:**
```rust
#[derive(Debug, Clone, PartialEq)]
struct Point {
    x: i32,
    y: i32,
}
```

**Go:**
```go
// Use go generate with tools like stringer, or implement manually
//go:generate stringer -type=Status

type Point struct {
    X, Y int
}

// Implement methods manually or use code generation
func (p Point) String() string {
    return fmt.Sprintf("Point{X: %d, Y: %d}", p.X, p.Y)
}

func (p Point) Equal(other Point) bool {
    return p.X == other.X && p.Y == other.Y
}
```

### Procedural Macros → Code Generation or Reflection

For complex macros, use:
- `go generate` with custom tools
- Build-time code generation
- Runtime reflection (with care)

```go
//go:generate go run gen.go

// Or use struct tags with reflection
type User struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"email"`
}
```

---

## Error Handling Patterns

### anyhow Equivalent

Rust's `anyhow` provides easy error wrapping. In Go, use `fmt.Errorf` with `%w`:

**Rust:**
```rust
use anyhow::{Context, Result};

fn read_config() -> Result<Config> {
    let contents = std::fs::read_to_string("config.json")
        .context("failed to read config file")?;
    let config: Config = serde_json::from_str(&contents)
        .context("failed to parse config")?;
    Ok(config)
}
```

**Go:**
```go
func readConfig() (Config, error) {
    contents, err := os.ReadFile("config.json")
    if err != nil {
        return Config{}, fmt.Errorf("failed to read config file: %w", err)
    }
    var config Config
    if err := json.Unmarshal(contents, &config); err != nil {
        return Config{}, fmt.Errorf("failed to parse config: %w", err)
    }
    return config, nil
}
```

### thiserror Equivalent

Rust's `thiserror` creates custom error types. In Go, define error types with `Error()` method:

**Rust:**
```rust
use thiserror::Error;

#[derive(Error, Debug)]
enum DataError {
    #[error("validation failed: {0}")]
    Validation(String),
    #[error("not found: {id}")]
    NotFound { id: u64 },
    #[error("database error")]
    Database(#[from] sqlx::Error),
}
```

**Go:**
```go
// Define sentinel errors for simple cases
var (
    ErrNotFound   = errors.New("not found")
    ErrValidation = errors.New("validation failed")
)

// Define custom error types for complex cases
type ValidationError struct {
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed: %s", e.Message)
}

type NotFoundError struct {
    ID uint64
}

func (e NotFoundError) Error() string {
    return fmt.Sprintf("not found: %d", e.ID)
}

type DatabaseError struct {
    Err error
}

func (e DatabaseError) Error() string {
    return "database error"
}

func (e DatabaseError) Unwrap() error {
    return e.Err
}

// Check error types with errors.Is and errors.As
func handleError(err error) {
    var notFound NotFoundError
    if errors.As(err, &notFound) {
        fmt.Printf("Resource %d not found\n", notFound.ID)
    }
}
```

---

## Naming Conventions

### General Rules

| Rust | Go |
|------|-----|
| `snake_case` for functions | `camelCase` for unexported, `PascalCase` for exported |
| `snake_case` for variables | `camelCase` for variables |
| `PascalCase` for types | `PascalCase` for types |
| `SCREAMING_SNAKE_CASE` for constants | `PascalCase` for exported constants |
| `snake_case` for modules | `lowercase` for packages |

### Examples

**Rust:**
```rust
mod user_service;

const MAX_CONNECTIONS: u32 = 100;

struct UserAccount {
    user_id: u64,
    display_name: String,
}

fn get_user_by_id(user_id: u64) -> Option<UserAccount> {
    // ...
}

impl UserAccount {
    fn full_name(&self) -> String {
        // ...
    }
}
```

**Go:**
```go
package userservice

const MaxConnections = 100 // exported
const maxRetries = 3       // unexported

type UserAccount struct {
    UserID      uint64 // exported
    DisplayName string // exported
}

func GetUserByID(userID uint64) (*UserAccount, bool) {
    // ...
}

// Unexported function
func validateUser(u *UserAccount) error {
    // ...
}

func (u *UserAccount) FullName() string {
    // ...
}
```

### Acronyms

Go prefers all-caps for acronyms:

| Rust | Go |
|------|-----|
| `HttpServer` | `HTTPServer` |
| `JsonParser` | `JSONParser` |
| `user_id` | `userID` |
| `Url` | `URL` |

### Getter Methods

Go doesn't use `get_` prefix:

| Rust | Go |
|------|-----|
| `fn get_name(&self)` | `func (u *User) Name()` |
| `fn get_id(&self)` | `func (u *User) ID()` |

### Setter Methods

Go uses `Set` prefix:

| Rust | Go |
|------|-----|
| `fn set_name(&mut self, name: String)` | `func (u *User) SetName(name string)` |

---

## Quick Reference Table

| Rust | Go |
|------|-----|
| `enum Foo { A, B(T) }` | `type Foo interface { isFoo() }` + concrete types |
| `Result<T, E>` | `(T, error)` |
| `Option<T>` | `*T` or `(T, bool)` |
| `trait Foo {}` | `type Foo interface {}` |
| `impl Foo for Bar` | Methods on `Bar` matching `Foo` interface |
| `impl Bar {}` | Functions and methods with `Bar` receiver |
| `match x {}` | `switch x {}` or `switch x.(type) {}` |
| `x?` | `if err != nil { return err }` |
| `clone()` | Assignment (shallow) or custom `Clone()` method |
| `&self` | Value receiver `(t T)` |
| `&mut self` | Pointer receiver `(t *T)` |
| `Vec<T>` | `[]T` |
| `HashMap<K, V>` | `map[K]V` |
| `Box<T>` | `*T` |
| `Arc<T>` | `*T` (GC handles sharing) |
| `Mutex<T>` | `sync.Mutex` + `T` separately |
| `pub` | `PascalCase` (exported) |
| `pub(crate)` | `camelCase` (unexported) |
