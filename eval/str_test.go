package eval

import (
	"testing"

	"github.com/boergens/gotypst/syntax"
)

func TestStrContains(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		pattern  string
		expected bool
	}{
		{"contains_yes", "abc", "b", true},
		{"contains_no", "abc", "d", false},
		{"contains_full", "abc", "abc", true},
		{"contains_empty", "abc", "", true},
		{"contains_in_empty", "", "a", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := NewArgs(syntax.Detached())
			args.Push(Str(tc.pattern), syntax.Detached())

			result, err := StrContains(Str(tc.target), args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			boolVal, ok := result.(BoolValue)
			if !ok {
				t.Fatalf("expected BoolValue, got %T", result)
			}

			if bool(boolVal) != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, boolVal)
			}
		})
	}
}

func TestStrStartsWith(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		pattern  string
		expected bool
	}{
		{"starts_yes", "Typst", "Ty", true},
		{"starts_no", "Typst", "st", false},
		{"starts_full", "abc", "abc", true},
		{"starts_empty", "abc", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := NewArgs(syntax.Detached())
			args.Push(Str(tc.pattern), syntax.Detached())

			result, err := StrStartsWith(Str(tc.target), args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			boolVal, ok := result.(BoolValue)
			if !ok {
				t.Fatalf("expected BoolValue, got %T", result)
			}

			if bool(boolVal) != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, boolVal)
			}
		})
	}
}

func TestStrEndsWith(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		pattern  string
		expected bool
	}{
		{"ends_yes", "Typst", "st", true},
		{"ends_no", "Typst", "Ty", false},
		{"ends_full", "abc", "abc", true},
		{"ends_empty", "abc", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := NewArgs(syntax.Detached())
			args.Push(Str(tc.pattern), syntax.Detached())

			result, err := StrEndsWith(Str(tc.target), args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			boolVal, ok := result.(BoolValue)
			if !ok {
				t.Fatalf("expected BoolValue, got %T", result)
			}

			if bool(boolVal) != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, boolVal)
			}
		})
	}
}

func TestStrFind(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		pattern  string
		expected interface{} // string or nil for none
	}{
		{"find_yes", "Hello World", "World", "World"},
		{"find_no", "Hello World", "xyz", nil},
		{"find_empty", "abc", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := NewArgs(syntax.Detached())
			args.Push(Str(tc.pattern), syntax.Detached())

			result, err := StrFind(Str(tc.target), args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.expected == nil {
				if _, ok := result.(NoneValue); !ok {
					t.Errorf("expected None, got %v", result)
				}
			} else {
				strVal, ok := result.(StrValue)
				if !ok {
					t.Fatalf("expected StrValue, got %T", result)
				}
				if string(strVal) != tc.expected.(string) {
					t.Errorf("expected %q, got %q", tc.expected, strVal)
				}
			}
		})
	}
}

func TestStrPosition(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		pattern  string
		expected interface{} // int64 or nil for none
	}{
		{"position_yes", "Hello World", "World", int64(6)},
		{"position_no", "Hello World", "xyz", nil},
		{"position_start", "abc", "a", int64(0)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := NewArgs(syntax.Detached())
			args.Push(Str(tc.pattern), syntax.Detached())

			result, err := StrPosition(Str(tc.target), args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.expected == nil {
				if _, ok := result.(NoneValue); !ok {
					t.Errorf("expected None, got %v", result)
				}
			} else {
				intVal, ok := result.(IntValue)
				if !ok {
					t.Fatalf("expected IntValue, got %T", result)
				}
				if int64(intVal) != tc.expected.(int64) {
					t.Errorf("expected %d, got %d", tc.expected, intVal)
				}
			}
		})
	}
}

func TestStrLen(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		expected int64
	}{
		{"len_normal", "Hello World!", 12},
		{"len_empty", "", 0},
		{"len_unicode", "日本語", 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := NewArgs(syntax.Detached())

			result, err := StrLen(Str(tc.target), args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			intVal, ok := result.(IntValue)
			if !ok {
				t.Fatalf("expected IntValue, got %T", result)
			}

			if int64(intVal) != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, intVal)
			}
		})
	}
}

func TestStrFirst(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		defVal   *string
		expected interface{}
		hasError bool
	}{
		{"first_normal", "Hello", nil, "H", false},
		{"first_with_default", "hey", strPtr("d"), "h", false},
		{"first_empty_with_default", "", strPtr("d"), "d", false},
		{"first_empty_no_default", "", nil, nil, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := NewArgs(syntax.Detached())
			if tc.defVal != nil {
				args.PushNamed("default", Str(*tc.defVal), syntax.Detached())
			}

			result, err := StrFirst(Str(tc.target), args, syntax.Detached())

			if tc.hasError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			strVal, ok := result.(StrValue)
			if !ok {
				t.Fatalf("expected StrValue, got %T", result)
			}

			if string(strVal) != tc.expected.(string) {
				t.Errorf("expected %q, got %q", tc.expected, strVal)
			}
		})
	}
}

func TestStrLast(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		defVal   *string
		expected interface{}
		hasError bool
	}{
		{"last_normal", "Hello", nil, "o", false},
		{"last_with_default", "hey", strPtr("d"), "y", false},
		{"last_empty_with_default", "", strPtr("d"), "d", false},
		{"last_empty_no_default", "", nil, nil, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := NewArgs(syntax.Detached())
			if tc.defVal != nil {
				args.PushNamed("default", Str(*tc.defVal), syntax.Detached())
			}

			result, err := StrLast(Str(tc.target), args, syntax.Detached())

			if tc.hasError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			strVal, ok := result.(StrValue)
			if !ok {
				t.Fatalf("expected StrValue, got %T", result)
			}

			if string(strVal) != tc.expected.(string) {
				t.Errorf("expected %q, got %q", tc.expected, strVal)
			}
		})
	}
}

func TestStrAt(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		index    int64
		defVal   *string
		expected interface{}
		hasError bool
	}{
		{"at_positive", "Hello", 1, nil, "e", false},
		{"at_last", "Hello", 4, nil, "o", false},
		{"at_negative", "Hello", -1, nil, "o", false},
		{"at_negative_two", "Hello", -2, nil, "l", false},
		{"at_with_default", "Hello", 5, strPtr("z"), "z", false},
		{"at_out_of_bounds", "Hello", 5, nil, nil, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := NewArgs(syntax.Detached())
			args.Push(Int(tc.index), syntax.Detached())
			if tc.defVal != nil {
				args.PushNamed("default", Str(*tc.defVal), syntax.Detached())
			}

			result, err := StrAt(Str(tc.target), args, syntax.Detached())

			if tc.hasError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			strVal, ok := result.(StrValue)
			if !ok {
				t.Fatalf("expected StrValue, got %T", result)
			}

			if string(strVal) != tc.expected.(string) {
				t.Errorf("expected %q, got %q", tc.expected, strVal)
			}
		})
	}
}

func TestStrSlice(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		start    int64
		end      *int64
		expected string
	}{
		{"slice_normal", "abc", 1, int64Ptr(2), "b"},
		{"slice_negative_end", "abc", 2, int64Ptr(-1), ""},
		{"slice_to_end", "abc", 1, nil, "bc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := NewArgs(syntax.Detached())
			args.Push(Int(tc.start), syntax.Detached())
			if tc.end != nil {
				args.Push(Int(*tc.end), syntax.Detached())
			}

			result, err := StrSlice(Str(tc.target), args, syntax.Detached())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			strVal, ok := result.(StrValue)
			if !ok {
				t.Fatalf("expected StrValue, got %T", result)
			}

			if string(strVal) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, strVal)
			}
		})
	}
}

func TestStrSplit(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		pattern  *string
		expected []string
	}{
		{"split_empty", "abc", strPtr(""), []string{"", "a", "b", "c", ""}},
		{"split_char", "abc", strPtr("b"), []string{"a", "c"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := NewArgs(syntax.Detached())
			if tc.pattern != nil {
				args.Push(Str(*tc.pattern), syntax.Detached())
			}

			result, err := StrSplit(Str(tc.target), args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			arrVal, ok := result.(ArrayValue)
			if !ok {
				t.Fatalf("expected ArrayValue, got %T", result)
			}

			if len(arrVal) != len(tc.expected) {
				t.Fatalf("expected %d elements, got %d", len(tc.expected), len(arrVal))
			}

			for i, exp := range tc.expected {
				strVal, ok := arrVal[i].(StrValue)
				if !ok {
					t.Errorf("element %d: expected StrValue, got %T", i, arrVal[i])
					continue
				}
				if string(strVal) != exp {
					t.Errorf("element %d: expected %q, got %q", i, exp, strVal)
				}
			}
		})
	}
}

func TestStrRev(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		expected string
	}{
		{"rev_normal", "abc", "cba"},
		{"rev_empty", "", ""},
		{"rev_single", "a", "a"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := NewArgs(syntax.Detached())

			result, err := StrRev(Str(tc.target), args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			strVal, ok := result.(StrValue)
			if !ok {
				t.Fatalf("expected StrValue, got %T", result)
			}

			if string(strVal) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, strVal)
			}
		})
	}
}

func TestStrMatch(t *testing.T) {
	t.Run("match_found", func(t *testing.T) {
		args := NewArgs(syntax.Detached())
		args.Push(Str("World"), syntax.Detached())

		result, err := StrMatch(Str("Hello World"), args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		dictVal, ok := result.(DictValue)
		if !ok {
			t.Fatalf("expected DictValue, got %T", result)
		}

		// Check start
		start, ok := dictVal.Get("start")
		if !ok || int64(start.(IntValue)) != 6 {
			t.Errorf("expected start=6, got %v", start)
		}

		// Check end
		end, ok := dictVal.Get("end")
		if !ok || int64(end.(IntValue)) != 11 {
			t.Errorf("expected end=11, got %v", end)
		}

		// Check text
		text, ok := dictVal.Get("text")
		if !ok || string(text.(StrValue)) != "World" {
			t.Errorf("expected text='World', got %v", text)
		}
	})

	t.Run("match_not_found", func(t *testing.T) {
		args := NewArgs(syntax.Detached())
		args.Push(Str("xyz"), syntax.Detached())

		result, err := StrMatch(Str("Hello World"), args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := result.(NoneValue); !ok {
			t.Errorf("expected None, got %v", result)
		}
	})
}

func TestStrMatches(t *testing.T) {
	t.Run("matches_found", func(t *testing.T) {
		args := NewArgs(syntax.Detached())
		args.Push(Str("o"), syntax.Detached())

		result, err := StrMatches(Str("Hello World"), args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arrVal, ok := result.(ArrayValue)
		if !ok {
			t.Fatalf("expected ArrayValue, got %T", result)
		}

		// "Hello World" has 'o' at positions 4 and 7
		if len(arrVal) != 2 {
			t.Errorf("expected 2 matches, got %d", len(arrVal))
		}
	})

	t.Run("matches_not_found", func(t *testing.T) {
		args := NewArgs(syntax.Detached())
		args.Push(Str("xyz"), syntax.Detached())

		result, err := StrMatches(Str("Hello World"), args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arrVal, ok := result.(ArrayValue)
		if !ok {
			t.Fatalf("expected ArrayValue, got %T", result)
		}

		if len(arrVal) != 0 {
			t.Errorf("expected 0 matches, got %d", len(arrVal))
		}
	})
}

// Helper functions for pointers
func strPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}
