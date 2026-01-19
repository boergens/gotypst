// Test the string methods.

--- str-constructor paged ---
// Test conversion to string.
#test(str(123), "123")
#test(str(123, base: 3), "11120")
#test(str(-123, base: 16), "-7b")
#test(str(9223372036854775807, base: 36), "1y2p0ij32e8e7")
#test(str(50.14), "50.14")
#test(str(10 / 3).len() > 10, true)

--- str-from-int paged ---
// Test the `str` function with integers.
#test(str(12), "12")
#test(str(1234567890), "1234567890")
#test(str(0123456789), "123456789")
#test(str(0), "0")
#test(str(-0), "0")
#test(str(-1), "-1")
#test(str(-9876543210), "-9876543210")
#test(str(-0987654321), "-987654321")
#test(str(4 - 8), "-4")

--- str-constructor-bad-type paged ---
// Error: 6-8 expected integer, float, decimal, version, bytes, label, type, or string, found content
#str([])

--- str-constructor-bad-base paged ---
// Error: 17-19 base must be between 2 and 36
#str(123, base: 99)

--- str-from-and-to-unicode paged ---
// Test the unicode function.
#test(str.from-unicode(97), "a")
#test(str.to-unicode("a"), 97)

--- str-from-unicode-bad-type paged ---
// Error: 19-22 expected integer, found content
#str.from-unicode([a])

--- str-to-unicode-bad-type paged ---
// Error: 17-21 expected exactly one character
#str.to-unicode("ab")

--- str-from-unicode-negative paged ---
// Error: 19-21 number must be at least zero
#str.from-unicode(-1)

--- str-from-unicode-bad-value paged ---
// Error: 2-28 0x110000 is not a valid codepoint
#str.from-unicode(0x110000) // 0x10ffff is the highest valid code point

--- string-len paged ---
// Test the `len` method.
#test("Hello World!".len(), 12)

--- string-first-and-last paged ---
// Test the `first` and `last` methods.
#test("Hello".first(), "H")
#test("Hello".last(), "o")
#test("hey".first(default: "d"), "h")
#test("".first(default: "d"), "d")
#test("hey".last(default: "d"), "y")
#test("".last(default: "d"), "d")

--- string-first-empty paged ---
// Error: 2-12 string is empty
#"".first()

--- string-last-empty paged ---
// Error: 2-11 string is empty
#"".last()

--- string-at paged ---
// Test the `at` method.
#test("Hello".at(1), "e")
#test("Hello".at(4), "o")
#test("Hello".at(-1), "o")
#test("Hello".at(-2), "l")

--- string-at-default paged ---
// Test `at`'s 'default' parameter.
#test("z", "Hello".at(5, default: "z"))

--- string-at-out-of-bounds paged ---
// Error: 2-15 no default value was specified and string index out of bounds (index: 5, len: 5)
#"Hello".at(5)

--- string-slice paged ---
// Test the `slice` method.
#test("abc".slice(1, 2), "b")
#test("abc".slice(2, -1), "")

--- string-contains paged ---
// Test the `contains` method.
#test("abc".contains("b"), true)
#test("b" in "abc", true)
#test("abc".contains("d"), false)

--- string-starts-with paged ---
// Test the `starts-with` and `ends-with` methods.
#test("Typst".starts-with("Ty"), true)
#test("Typst".starts-with("st"), false)

--- string-ends-with paged ---
#test("Typst".ends-with("st"), true)
#test("Typst".ends-with("Ty"), false)

--- string-find-and-position paged ---
// Test the `find` and `position` methods.
#test("Hello World".find("World"), "World")
#test("Hello World".position("World"), 6)

--- string-split paged ---
// Test the `split` method.
#test("abc".split(""), ("", "a", "b", "c", ""))
#test("abc".split("b"), ("a", "c"))

--- string-rev paged ---
// Test the `rev` method.
#test("abc".rev(), "cba")

--- string-unclosed paged ---
// Error: 2-2:1 unclosed string
#"hello\"
