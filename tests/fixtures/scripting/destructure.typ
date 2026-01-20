// Test destructuring patterns.

--- destructure-array-basic paged ---
// Basic array destructuring with exact length match.
#let (a, b, c) = (1, 2, 3)
#test(a, 1)
#test(b, 2)
#test(c, 3)

// Single element array.
#let (x) = (42,)
#test(x, 42)

// Two element array.
#let (first, second) = ("hello", "world")
#test(first, "hello")
#test(second, "world")

--- destructure-array-spread paged ---
// Spread at end collects remaining elements.
#let (first, ..rest) = (1, 2, 3, 4)
#test(first, 1)
#test(rest, (2, 3, 4))

// Spread at beginning.
#let (..init, last) = (1, 2, 3, 4)
#test(init, (1, 2, 3))
#test(last, 4)

// Spread in middle.
#let (head, ..middle, tail) = (1, 2, 3, 4, 5)
#test(head, 1)
#test(middle, (2, 3, 4))
#test(tail, 5)

// Spread with empty result.
#let (a, b, ..empty) = (1, 2)
#test(a, 1)
#test(b, 2)
#test(empty, ())

// Spread collecting all elements.
#let (..all) = (1, 2, 3)
#test(all, (1, 2, 3))

--- destructure-dict-shorthand paged ---
// Dictionary destructuring with shorthand syntax.
#let (name) = (name: "Alice")
#test(name, "Alice")

// Multiple shorthand bindings.
#let (name, age) = (name: "Bob", age: 25)
#test(name, "Bob")
#test(age, 25)

--- destructure-dict-named paged ---
// Dictionary destructuring with renamed bindings.
#let (name: n, age: a) = (name: "Charlie", age: 30)
#test(n, "Charlie")
#test(a, 30)

// Mixed shorthand and named.
#let (name, age: years) = (name: "Diana", age: 28)
#test(name, "Diana")
#test(years, 28)

--- destructure-dict-spread paged ---
// Dictionary spread collects remaining keys.
#let (a, ..rest) = (a: 1, b: 2, c: 3)
#test(a, 1)
#test(rest, (b: 2, c: 3))

// Spread with named binding.
#let (x: val, ..others) = (x: 10, y: 20, z: 30)
#test(val, 10)
#test(others, (y: 20, z: 30))

// Spread collecting all.
#let (..everything) = (foo: 1, bar: 2)
#test(everything, (foo: 1, bar: 2))

--- destructure-nested paged ---
// Nested array destructuring.
#let ((a, b), c) = ((1, 2), 3)
#test(a, 1)
#test(b, 2)
#test(c, 3)

// Deeply nested.
#let (((x))) = (((42)))
#test(x, 42)

// Mixed nesting with spread.
#let ((first, ..inner), outer) = ((1, 2, 3), 4)
#test(first, 1)
#test(inner, (2, 3))
#test(outer, 4)

--- destructure-placeholder paged ---
// Placeholder discards values.
#let (a, _, c) = (1, 2, 3)
#test(a, 1)
#test(c, 3)

// Multiple placeholders.
#let (_, x, _) = (1, 2, 3)
#test(x, 2)

// Placeholder with spread.
#let (_, ..rest) = (1, 2, 3)
#test(rest, (2, 3))

// All placeholders (just validates structure).
#let (_, _, _) = (1, 2, 3)

--- destructure-for-loop paged ---
// Destructuring in for loops.
#let sum = 0
#for (k, v) in (a: 1, b: 2, c: 3) {
  sum += v
}
#test(sum, 6)

// Array of arrays.
#let results = ()
#for (x, y) in ((1, 2), (3, 4), (5, 6)) {
  results += (x + y,)
}
#test(results, (3, 7, 11))

