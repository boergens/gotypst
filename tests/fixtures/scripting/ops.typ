// Test binary expressions.

--- ops-add-content paged ---
// Test adding content.
#([*Hello* ] + [world!])

--- ops-unary-basic paged ---
// Test math operators.

// Test plus and minus.
#for v in (1, 3.14, 12pt, 45deg, 90%, 13% + 10pt, 6.3fr) {
  // Test plus.
  test(+v, v)

  // Test minus.
  test(-v, -1 * v)
  test(--v, v)

  // Test combination.
  test(-++ --v, -v)
}

#test(-(4 + 2), 6-12)

// Addition.
#test(2 + 4, 6)
#test("a" + "b", "ab")
#test("a" + if false { "b" }, "a")
#test("a" + if true { "b" }, "ab")
#test(13 * "a" + "bbbbbb", "aaaaaaaaaaaaabbbbbb")
#test((1, 2) + (3, 4), (1, 2, 3, 4))
#test((a: 1) + (b: 2, c: 3), (a: 1, b: 2, c: 3))

--- ops-add-too-large paged ---
// Error: 3-26 value is too large
#(9223372036854775807 + 1)

--- ops-binary-basic paged ---
// Subtraction.
#test(1-4, 3*-1)
#test(4cm - 2cm, 2cm)
#test(1e+2-1e-2, 99.99)

// Multiplication.
#test(2 * 4, 8)

// Division.
#test(12pt/.4, 30pt)
#test(7 / 2, 3.5)

// Combination.
#test(3-4 * 5 < -10, true)
#test({ let x; x = 1 + 4*5 >= 21 and { x = "a"; x + "b" == "ab" }; x }, true)

// With block.
#test(if true {
  1
} + 2, 3)

--- ops-unary-bool paged ---
// Test boolean operators.

// Test not.
#test(not true, false)
#test(not false, true)

// And.
#test(false and false, false)
#test(false and true, false)
#test(true and false, false)
#test(true and true, true)

// Or.
#test(false or false, false)
#test(false or true, true)
#test(true or false, true)
#test(true or true, true)

// Short-circuiting.
#test(false and dont-care, false)
#test(true or dont-care, true)

--- ops-equality paged ---
// Test equality operators.

// Most things compare by value.
#test(1 == "hi", false)
#test(1 == 1.0, true)
#test(30% == 30% + 0cm, true)
#test(1in == 0% + 72pt, true)
#test(30% == 30% + 1cm, false)
#test("ab" == "a" + "b", true)
#test(() == (1,), false)
#test((1, 2, 3) == (1, 2.0) + (3,), true)
#test((:) == (a: 1), false)
#test((a: 2 - 1.0, b: 2) == (b: 2, a: 1), true)
#test("a" != "a", false)

// Functions compare by identity.
#test(test == test, true)
#test((() => {}) == (() => {}), false)

// Content compares field by field.
#let t = [a]
#test(t == t, true)
#test([] == [], true)
#test([a] == [a], true)

--- ops-compare paged ---
// Test comparison operators.

#test(13 * 3 < 14 * 4, true)
#test(5 < 10, true)
#test(5 > 5, false)
#test(5 <= 5, true)
#test(5 <= 4, false)
#test(45deg < 1rad, true)
#test(10% < 20%, true)
#test(50% < 40% + 0pt, false)
#test(40% + 0pt < 50% + 0pt, true)
#test(1em < 2em, true)
#test((0, 1, 2, 4) < (0, 1, 2, 5), true)
#test((0, 1, 2, 4) < (0, 1, 2, 3), false)
#test((0, 1, 2, 3.3) > (0, 1, 2, 4), false)
#test((0, 1, 2) < (0, 1, 2, 3), true)
#test((0, 1, "b") > (0, 1, "a", 3), true)
#test((0, 1.1, 3) >= (0, 1.1, 3), true)
#test(("a", 23, 40, "b") > ("a", 23, 40), true)
#test(() <= (), true)
#test(() >= (), true)
#test(() <= (1,), true)
#test((1,) <= (), false)

--- ops-in paged ---
// Test `in` operator.
#test("hi" in "worship", true)
#test("hi" in ("we", "hi", "bye"), true)
#test("Hey" in "abHeyCd", true)
#test("Hey" in "abheyCd", false)
#test(5 in range(10), true)
#test(12 in range(10), false)
#test("" in (), false)
#test("key" in (key: "value"), true)
#test("value" in (key: "value"), false)
#test("Hey" not in "abheyCd", true)

--- ops-not-trailing paged ---
// Error: 10 expected keyword `in`
#("a" not)

--- ops-precedence-basic paged ---
// Multiplication binds stronger than addition.
#test(1+2*-3, -5)

// Subtraction binds stronger than comparison.
#test(3 == 5 - 2, true)

// Boolean operations bind stronger than '=='.
#test("a" == "a" and 2 < 3, true)
#test(not "b" == "b", false)

--- ops-precedence-parentheses paged ---
// Parentheses override precedence.
#test((1), 1)
#test((1+2)*-3, -9)

// Error: 8-9 unclosed delimiter
#test({(1 + 1}, 2)

--- ops-associativity-left paged ---
// Math operators are left-associative.
#test(10 / 2 / 2 == (10 / 2) / 2, true)
#test(10 / 2 / 2 == 10 / (2 / 2), false)
#test(1 / 2 * 3, 1.5)

--- ops-associativity-right paged ---
// Assignment is right-associative.
#{
  let x = 1
  let y = 2
  x = y = "ok"
  test(x, none)
  test(y, "ok")
}

--- ops-unary-minus-missing-expr paged ---
// Error: 4 expected expression
#(-)

--- ops-add-missing-rhs paged ---
// Error: 10 expected expression
#test({1+}, 1)

--- ops-mul-missing-rhs paged ---
// Error: 10 expected expression
#test({2*}, 2)

--- ops-unary-plus-on-content paged ---
// Error: 3-13 cannot apply unary '+' to content
#(+([] + []))

--- ops-unary-plus-on-string paged ---
// Error: 3-6 cannot apply '-' to string
#(-"")

--- ops-not-on-array paged ---
// Error: 3-9 cannot apply 'not' to array
#(not ())

--- ops-divide-by-zero-float paged ---
// Error: 3-12 cannot divide by zero
#(1.2 / 0.0)

--- ops-divide-by-zero-int paged ---
// Error: 3-8 cannot divide by zero
#(1 / 0)

--- ops-binary-arithmetic-error-message paged ---
// Special messages for +, -, * and /.
// Error: 3-10 cannot add integer and string
#(1 + "2", 40% - 1)

--- ops-assign paged ---
// Test assignment operators.

#let x = 0
#(x = 10)       #test(x, 10)
#(x -= 5)       #test(x, 5)
#(x += 1)       #test(x, 6)
#(x *= x)       #test(x, 36)
#(x /= 2.0)     #test(x, 18.0)
#(x = "some")   #test(x, "some")
#(x += "thing") #test(x, "something")

--- ops-assign-unknown-var-lhs paged ---
#{
  // Error: 3-6 unknown variable: a-1
  // Hint: 3-6 if you meant to use subtraction, try adding spaces around the minus sign: `a - 1`
  a-1 = 2
}

--- ops-assign-to-temporary paged ---
// Error: 3-8 cannot mutate a temporary value
#(1 + 2 += 3)

--- ops-assign-unknown-variable paged ---
// Error: 3-4 unknown variable: z
#(z = 1)

--- ops-assign-to-std-constant paged ---
// Error: 3-7 cannot mutate a constant: rect
#(rect = "hi")
