// Test arrays.

--- array-basic paged ---
// Empty array.
#test((), ())

// Array with items.
#test((1, 2, 3), (1, 2, 3))

// Array with trailing comma.
#test((1, 2, 3,), (1, 2, 3))

// Single item needs trailing comma.
#test((1,), (1,))

--- array-len paged ---
// Test the `len` method.
#test(().len(), 0)
#test((1, 2, 3).len(), 3)

--- array-first-and-last paged ---
// Test the `first` and `last` methods.
#test((1, 2, 3).first(), 1)
#test((1, 2, 3).last(), 3)
#test((1,).first(), 1)
#test((1,).last(), 1)

--- array-first-empty paged ---
// Error: 2-12 array is empty
#().first()

--- array-last-empty paged ---
// Error: 2-11 array is empty
#().last()

--- array-at paged ---
// Test the `at` method.
#test((1, 2, 3).at(0), 1)
#test((1, 2, 3).at(2), 3)
#test((1, 2, 3).at(-1), 3)
#test((1, 2, 3).at(-3), 1)

--- array-at-default paged ---
// Test `at`'s 'default' parameter.
#test((1, 2, 3).at(3, default: 0), 0)
#test((1, 2, 3).at(-4, default: 0), 0)

--- array-at-out-of-bounds paged ---
// Error: 2-16 array index out of bounds (index: 3, len: 3) and no default value was specified
#(1, 2, 3).at(3)

--- array-push paged ---
// Test the `push` method.
#{
  let arr = (1, 2)
  arr.push(3)
  test(arr, (1, 2, 3))
}

--- array-pop paged ---
// Test the `pop` method.
#{
  let arr = (1, 2, 3)
  test(arr.pop(), 3)
  test(arr, (1, 2))
}

--- array-pop-empty paged ---
// Error: 2-10 array is empty
#().pop()

--- array-insert paged ---
// Test the `insert` method.
#{
  let arr = (1, 3)
  arr.insert(1, 2)
  test(arr, (1, 2, 3))
}

--- array-remove paged ---
// Test the `remove` method.
#{
  let arr = (1, 2, 3)
  test(arr.remove(1), 2)
  test(arr, (1, 3))
}

--- array-slice paged ---
// Test the `slice` method.
#test((1, 2, 3, 4).slice(1, 3), (2, 3))
#test((1, 2, 3, 4).slice(2), (3, 4))
#test((1, 2, 3, 4).slice(-2), (3, 4))
#test((1, 2, 3, 4).slice(1, -1), (2, 3))

--- array-contains paged ---
// Test the `contains` method.
#test((1, 2, 3).contains(2), true)
#test(2 in (1, 2, 3), true)
#test((1, 2, 3).contains(4), false)
#test(4 in (1, 2, 3), false)

--- array-find paged ---
// Test the `find` method.
#test((1, 2, 3).find(x => x > 1), 2)
#test((1, 2, 3).find(x => x > 3), none)

--- array-position paged ---
// Test the `position` method.
#test((1, 2, 3).position(x => x > 1), 1)
#test((1, 2, 3).position(x => x > 3), none)

--- array-filter paged ---
// Test the `filter` method.
#test((1, 2, 3, 4).filter(x => calc.rem(x, 2) == 0), (2, 4))

--- array-map paged ---
// Test the `map` method.
#test((1, 2, 3).map(x => x * 2), (2, 4, 6))

--- array-fold paged ---
// Test the `fold` method.
#test((1, 2, 3).fold(0, (acc, x) => acc + x), 6)

--- array-sum paged ---
// Test the `sum` method.
#test((1, 2, 3).sum(), 6)
#test(().sum(default: 0), 0)

--- array-product paged ---
// Test the `product` method.
#test((1, 2, 3, 4).product(), 24)
#test(().product(default: 1), 1)

--- array-rev paged ---
// Test the `rev` method.
#test((1, 2, 3).rev(), (3, 2, 1))

--- array-join paged ---
// Test the `join` method.
#test(("a", "b", "c").join("-"), "a-b-c")
#test((1, 2, 3).join([, ]), [1, 2, 3])

--- array-sorted paged ---
// Test the `sorted` method.
#test((3, 1, 2).sorted(), (1, 2, 3))
#test(("c", "a", "b").sorted(), ("a", "b", "c"))

--- array-zip paged ---
// Test the `zip` method.
#test((1, 2).zip(("a", "b")), ((1, "a"), (2, "b")))

--- array-enumerate paged ---
// Test the `enumerate` method.
#test(("a", "b", "c").enumerate(), ((0, "a"), (1, "b"), (2, "c")))

--- array-dedup paged ---
// Test the `dedup` method.
#test((1, 2, 2, 3, 3, 3).dedup(), (1, 2, 3))
