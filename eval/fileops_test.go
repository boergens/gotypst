package eval

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Test Helper - Mock World
// ----------------------------------------------------------------------------

// mockWorld is a simple test World that serves files from a map
type mockWorld struct {
	files    map[string][]byte
	mainFile FileID
	library  *Scope
}

func newMockWorld(files map[string][]byte, mainPath string) *mockWorld {
	return &mockWorld{
		files:    files,
		mainFile: FileID{Path: mainPath},
		library:  NewScope(),
	}
}

func (w *mockWorld) Library() *Scope {
	return w.library
}

func (w *mockWorld) MainFile() FileID {
	return w.mainFile
}

func (w *mockWorld) Source(id FileID) (*syntax.Source, error) {
	return nil, nil
}

func (w *mockWorld) File(id FileID) ([]byte, error) {
	if data, ok := w.files[id.Path]; ok {
		return data, nil
	}
	return nil, &FileNotFoundError{Path: id.Path}
}

func (w *mockWorld) Today(offset *int) Date {
	return Date{Year: 2026, Month: 1, Day: 19}
}

func newTestVm(world World) *Vm {
	engine := NewEngine(world)
	scopes := NewScopes(nil)
	return NewVm(engine, NewContext(), scopes, syntax.Detached())
}

// ----------------------------------------------------------------------------
// ReadFunc Tests
// ----------------------------------------------------------------------------

func TestReadFunc(t *testing.T) {
	readFunc := ReadFunc()

	if readFunc == nil {
		t.Fatal("ReadFunc() returned nil")
	}

	if readFunc.Name == nil || *readFunc.Name != "read" {
		t.Errorf("expected function name 'read', got %v", readFunc.Name)
	}

	_, ok := readFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestReadNativeBasic(t *testing.T) {
	world := newMockWorld(map[string][]byte{
		"/test/file.txt": []byte("Hello, World!"),
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("file.txt"), syntax.Detached())

	result, err := readNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Fatalf("readNative() error: %v", err)
	}

	str, ok := result.(StrValue)
	if !ok {
		t.Fatalf("expected StrValue, got %T", result)
	}

	if string(str) != "Hello, World!" {
		t.Errorf("got %q, want %q", str, "Hello, World!")
	}
}

func TestReadNativeAbsolutePath(t *testing.T) {
	world := newMockWorld(map[string][]byte{
		"/absolute/path/file.txt": []byte("Absolute path content"),
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("/absolute/path/file.txt"), syntax.Detached())

	result, err := readNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Fatalf("readNative() error: %v", err)
	}

	str, ok := result.(StrValue)
	if !ok {
		t.Fatalf("expected StrValue, got %T", result)
	}

	if string(str) != "Absolute path content" {
		t.Errorf("got %q, want %q", str, "Absolute path content")
	}
}

func TestReadNativeBinaryEncoding(t *testing.T) {
	world := newMockWorld(map[string][]byte{
		"/test/file.bin": {0x00, 0x01, 0x02, 0xFF},
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("file.bin"), syntax.Detached())
	args.PushNamed("encoding", None, syntax.Detached())

	result, err := readNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Fatalf("readNative() error: %v", err)
	}

	bytes, ok := result.(BytesValue)
	if !ok {
		t.Fatalf("expected BytesValue, got %T", result)
	}

	expected := []byte{0x00, 0x01, 0x02, 0xFF}
	if len(bytes) != len(expected) {
		t.Errorf("got length %d, want %d", len(bytes), len(expected))
	}
}

func TestReadNativeFileNotFound(t *testing.T) {
	world := newMockWorld(map[string][]byte{}, "/test/main.typ")
	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("nonexistent.txt"), syntax.Detached())

	_, err := readNative(vm.Engine, vm.Context, args)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestReadNativeMissingPath(t *testing.T) {
	world := newMockWorld(map[string][]byte{}, "/test/main.typ")
	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())

	_, err := readNative(vm.Engine, vm.Context, args)
	if err == nil {
		t.Error("expected error for missing path argument")
	}
}

// ----------------------------------------------------------------------------
// JsonFunc Tests
// ----------------------------------------------------------------------------

func TestJsonFunc(t *testing.T) {
	jsonFunc := JsonFunc()

	if jsonFunc == nil {
		t.Fatal("JsonFunc() returned nil")
	}

	if jsonFunc.Name == nil || *jsonFunc.Name != "json" {
		t.Errorf("expected function name 'json', got %v", jsonFunc.Name)
	}
}

func TestJsonNativeObject(t *testing.T) {
	jsonData := `{"name": "John", "age": 30, "active": true}`
	world := newMockWorld(map[string][]byte{
		"/test/data.json": []byte(jsonData),
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("data.json"), syntax.Detached())

	result, err := jsonNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Fatalf("jsonNative() error: %v", err)
	}

	dict, ok := result.(DictValue)
	if !ok {
		t.Fatalf("expected DictValue, got %T", result)
	}

	// Check name
	nameVal, ok := dict.Get("name")
	if !ok {
		t.Error("expected 'name' key in dict")
	} else if str, ok := nameVal.(StrValue); !ok || string(str) != "John" {
		t.Errorf("name = %v, want 'John'", nameVal)
	}

	// Check age
	ageVal, ok := dict.Get("age")
	if !ok {
		t.Error("expected 'age' key in dict")
	} else if num, ok := ageVal.(IntValue); !ok || int64(num) != 30 {
		t.Errorf("age = %v, want 30", ageVal)
	}

	// Check active
	activeVal, ok := dict.Get("active")
	if !ok {
		t.Error("expected 'active' key in dict")
	} else if b, ok := activeVal.(BoolValue); !ok || bool(b) != true {
		t.Errorf("active = %v, want true", activeVal)
	}
}

func TestJsonNativeArray(t *testing.T) {
	jsonData := `[1, 2, 3, "four"]`
	world := newMockWorld(map[string][]byte{
		"/test/data.json": []byte(jsonData),
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("data.json"), syntax.Detached())

	result, err := jsonNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Fatalf("jsonNative() error: %v", err)
	}

	arr, ok := result.(ArrayValue)
	if !ok {
		t.Fatalf("expected ArrayValue, got %T", result)
	}

	if len(arr) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(arr))
	}

	// Check elements
	if v, ok := arr[0].(IntValue); !ok || int64(v) != 1 {
		t.Errorf("arr[0] = %v, want 1", arr[0])
	}
	if v, ok := arr[3].(StrValue); !ok || string(v) != "four" {
		t.Errorf("arr[3] = %v, want 'four'", arr[3])
	}
}

func TestJsonNativeInvalid(t *testing.T) {
	world := newMockWorld(map[string][]byte{
		"/test/data.json": []byte(`{invalid json}`),
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("data.json"), syntax.Detached())

	_, err := jsonNative(vm.Engine, vm.Context, args)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// ----------------------------------------------------------------------------
// YamlFunc Tests
// ----------------------------------------------------------------------------

func TestYamlFunc(t *testing.T) {
	yamlFunc := YamlFunc()

	if yamlFunc == nil {
		t.Fatal("YamlFunc() returned nil")
	}

	if yamlFunc.Name == nil || *yamlFunc.Name != "yaml" {
		t.Errorf("expected function name 'yaml', got %v", yamlFunc.Name)
	}
}

func TestYamlNativeObject(t *testing.T) {
	yamlData := `
name: John
age: 30
active: true
`
	world := newMockWorld(map[string][]byte{
		"/test/data.yaml": []byte(yamlData),
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("data.yaml"), syntax.Detached())

	result, err := yamlNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Fatalf("yamlNative() error: %v", err)
	}

	dict, ok := result.(DictValue)
	if !ok {
		t.Fatalf("expected DictValue, got %T", result)
	}

	// Check name
	nameVal, ok := dict.Get("name")
	if !ok {
		t.Error("expected 'name' key in dict")
	} else if str, ok := nameVal.(StrValue); !ok || string(str) != "John" {
		t.Errorf("name = %v, want 'John'", nameVal)
	}

	// Check age
	ageVal, ok := dict.Get("age")
	if !ok {
		t.Error("expected 'age' key in dict")
	} else if num, ok := ageVal.(IntValue); !ok || int64(num) != 30 {
		t.Errorf("age = %v, want 30", ageVal)
	}
}

func TestYamlNativeArray(t *testing.T) {
	yamlData := `
- apple
- banana
- cherry
`
	world := newMockWorld(map[string][]byte{
		"/test/data.yaml": []byte(yamlData),
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("data.yaml"), syntax.Detached())

	result, err := yamlNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Fatalf("yamlNative() error: %v", err)
	}

	arr, ok := result.(ArrayValue)
	if !ok {
		t.Fatalf("expected ArrayValue, got %T", result)
	}

	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}

	if v, ok := arr[0].(StrValue); !ok || string(v) != "apple" {
		t.Errorf("arr[0] = %v, want 'apple'", arr[0])
	}
}

// ----------------------------------------------------------------------------
// TomlFunc Tests
// ----------------------------------------------------------------------------

func TestTomlFunc(t *testing.T) {
	tomlFunc := TomlFunc()

	if tomlFunc == nil {
		t.Fatal("TomlFunc() returned nil")
	}

	if tomlFunc.Name == nil || *tomlFunc.Name != "toml" {
		t.Errorf("expected function name 'toml', got %v", tomlFunc.Name)
	}
}

func TestTomlNative(t *testing.T) {
	tomlData := `
name = "John"
age = 30

[database]
server = "localhost"
port = 5432
`
	world := newMockWorld(map[string][]byte{
		"/test/config.toml": []byte(tomlData),
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("config.toml"), syntax.Detached())

	result, err := tomlNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Fatalf("tomlNative() error: %v", err)
	}

	dict, ok := result.(DictValue)
	if !ok {
		t.Fatalf("expected DictValue, got %T", result)
	}

	// Check name
	nameVal, ok := dict.Get("name")
	if !ok {
		t.Error("expected 'name' key in dict")
	} else if str, ok := nameVal.(StrValue); !ok || string(str) != "John" {
		t.Errorf("name = %v, want 'John'", nameVal)
	}

	// Check nested database
	dbVal, ok := dict.Get("database")
	if !ok {
		t.Error("expected 'database' key in dict")
	} else {
		dbDict, ok := dbVal.(DictValue)
		if !ok {
			t.Fatalf("expected database to be DictValue, got %T", dbVal)
		}

		serverVal, ok := dbDict.Get("server")
		if !ok {
			t.Error("expected 'server' key in database dict")
		} else if str, ok := serverVal.(StrValue); !ok || string(str) != "localhost" {
			t.Errorf("database.server = %v, want 'localhost'", serverVal)
		}
	}
}

func TestTomlNativeInvalid(t *testing.T) {
	world := newMockWorld(map[string][]byte{
		"/test/config.toml": []byte(`[invalid toml`),
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("config.toml"), syntax.Detached())

	_, err := tomlNative(vm.Engine, vm.Context, args)
	if err == nil {
		t.Error("expected error for invalid TOML")
	}
}

// ----------------------------------------------------------------------------
// CsvFunc Tests
// ----------------------------------------------------------------------------

func TestCsvFunc(t *testing.T) {
	csvFunc := CsvFunc()

	if csvFunc == nil {
		t.Fatal("CsvFunc() returned nil")
	}

	if csvFunc.Name == nil || *csvFunc.Name != "csv" {
		t.Errorf("expected function name 'csv', got %v", csvFunc.Name)
	}
}

func TestCsvNativeArrayMode(t *testing.T) {
	csvData := `name,age,city
John,30,NYC
Jane,25,LA`
	world := newMockWorld(map[string][]byte{
		"/test/data.csv": []byte(csvData),
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("data.csv"), syntax.Detached())

	result, err := csvNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Fatalf("csvNative() error: %v", err)
	}

	arr, ok := result.(ArrayValue)
	if !ok {
		t.Fatalf("expected ArrayValue, got %T", result)
	}

	if len(arr) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(arr))
	}

	// Check first row (header)
	row0, ok := arr[0].(ArrayValue)
	if !ok {
		t.Fatalf("expected row 0 to be ArrayValue, got %T", arr[0])
	}
	if len(row0) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(row0))
	}
	if v, ok := row0[0].(StrValue); !ok || string(v) != "name" {
		t.Errorf("row0[0] = %v, want 'name'", row0[0])
	}
}

func TestCsvNativeDictMode(t *testing.T) {
	csvData := `name,age,city
John,30,NYC
Jane,25,LA`
	world := newMockWorld(map[string][]byte{
		"/test/data.csv": []byte(csvData),
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("data.csv"), syntax.Detached())
	args.PushNamed("row-type", Str("dict"), syntax.Detached())

	result, err := csvNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Fatalf("csvNative() error: %v", err)
	}

	arr, ok := result.(ArrayValue)
	if !ok {
		t.Fatalf("expected ArrayValue, got %T", result)
	}

	if len(arr) != 2 {
		t.Fatalf("expected 2 data rows (excluding header), got %d", len(arr))
	}

	// Check first data row
	row0, ok := arr[0].(DictValue)
	if !ok {
		t.Fatalf("expected row 0 to be DictValue, got %T", arr[0])
	}

	nameVal, ok := row0.Get("name")
	if !ok {
		t.Error("expected 'name' key in row")
	} else if str, ok := nameVal.(StrValue); !ok || string(str) != "John" {
		t.Errorf("name = %v, want 'John'", nameVal)
	}

	ageVal, ok := row0.Get("age")
	if !ok {
		t.Error("expected 'age' key in row")
	} else if str, ok := ageVal.(StrValue); !ok || string(str) != "30" {
		t.Errorf("age = %v, want '30'", ageVal)
	}
}

func TestCsvNativeCustomDelimiter(t *testing.T) {
	csvData := `name;age;city
John;30;NYC`
	world := newMockWorld(map[string][]byte{
		"/test/data.csv": []byte(csvData),
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("data.csv"), syntax.Detached())
	args.PushNamed("delimiter", Str(";"), syntax.Detached())

	result, err := csvNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Fatalf("csvNative() error: %v", err)
	}

	arr := result.(ArrayValue)
	row0 := arr[0].(ArrayValue)

	if len(row0) != 3 {
		t.Errorf("expected 3 columns with semicolon delimiter, got %d", len(row0))
	}
}

func TestCsvNativeInvalidDelimiter(t *testing.T) {
	world := newMockWorld(map[string][]byte{
		"/test/data.csv": []byte("a,b,c"),
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("data.csv"), syntax.Detached())
	args.PushNamed("delimiter", Str(";;"), syntax.Detached()) // Invalid: more than 1 char

	_, err := csvNative(vm.Engine, vm.Context, args)
	if err == nil {
		t.Error("expected error for invalid delimiter")
	}
}

// ----------------------------------------------------------------------------
// XmlFunc Tests
// ----------------------------------------------------------------------------

func TestXmlFunc(t *testing.T) {
	xmlFunc := XmlFunc()

	if xmlFunc == nil {
		t.Fatal("XmlFunc() returned nil")
	}

	if xmlFunc.Name == nil || *xmlFunc.Name != "xml" {
		t.Errorf("expected function name 'xml', got %v", xmlFunc.Name)
	}
}

func TestXmlNativeBasic(t *testing.T) {
	xmlData := `<?xml version="1.0"?>
<root>
  <item id="1">First</item>
  <item id="2">Second</item>
</root>`
	world := newMockWorld(map[string][]byte{
		"/test/data.xml": []byte(xmlData),
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("data.xml"), syntax.Detached())

	result, err := xmlNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Fatalf("xmlNative() error: %v", err)
	}

	dict, ok := result.(DictValue)
	if !ok {
		t.Fatalf("expected DictValue, got %T", result)
	}

	// Check tag name
	tagVal, ok := dict.Get("tag")
	if !ok {
		t.Error("expected 'tag' key in dict")
	} else if str, ok := tagVal.(StrValue); !ok || string(str) != "root" {
		t.Errorf("tag = %v, want 'root'", tagVal)
	}

	// Check children
	childrenVal, ok := dict.Get("children")
	if !ok {
		t.Error("expected 'children' key in dict")
	} else {
		children, ok := childrenVal.(ArrayValue)
		if !ok {
			t.Fatalf("expected children to be ArrayValue, got %T", childrenVal)
		}
		if len(children) != 2 {
			t.Errorf("expected 2 children, got %d", len(children))
		}

		// Check first child
		if len(children) > 0 {
			child0, ok := children[0].(DictValue)
			if !ok {
				t.Fatalf("expected child 0 to be DictValue, got %T", children[0])
			}

			// Check tag
			childTag, _ := child0.Get("tag")
			if str, ok := childTag.(StrValue); !ok || string(str) != "item" {
				t.Errorf("child tag = %v, want 'item'", childTag)
			}

			// Check attrs
			attrsVal, ok := child0.Get("attrs")
			if ok {
				attrs, ok := attrsVal.(DictValue)
				if !ok {
					t.Fatalf("expected attrs to be DictValue, got %T", attrsVal)
				}
				idVal, _ := attrs.Get("id")
				if str, ok := idVal.(StrValue); !ok || string(str) != "1" {
					t.Errorf("id attr = %v, want '1'", idVal)
				}
			}

			// Check text
			textVal, _ := child0.Get("text")
			if str, ok := textVal.(StrValue); !ok || string(str) != "First" {
				t.Errorf("text = %v, want 'First'", textVal)
			}
		}
	}
}

func TestXmlNativeInvalid(t *testing.T) {
	world := newMockWorld(map[string][]byte{
		"/test/data.xml": []byte(`<root><unclosed>`),
	}, "/test/main.typ")

	vm := newTestVm(world)

	args := NewArgs(syntax.Detached())
	args.Push(Str("data.xml"), syntax.Detached())

	_, err := xmlNative(vm.Engine, vm.Context, args)
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}

// ----------------------------------------------------------------------------
// Registration Tests
// ----------------------------------------------------------------------------

func TestRegisterFileOperations(t *testing.T) {
	scope := NewScope()
	RegisterFileOperations(scope)

	expectedFuncs := []string{"read", "json", "yaml", "toml", "csv", "xml"}

	for _, name := range expectedFuncs {
		binding := scope.Get(name)
		if binding == nil {
			t.Errorf("expected '%s' to be registered", name)
			continue
		}

		funcVal, ok := binding.Value.(FuncValue)
		if !ok {
			t.Errorf("expected FuncValue for '%s', got %T", name, binding.Value)
			continue
		}

		if funcVal.Func.Name == nil || *funcVal.Func.Name != name {
			t.Errorf("expected function name '%s', got %v", name, funcVal.Func.Name)
		}
	}
}

func TestFileOperations(t *testing.T) {
	funcs := FileOperations()

	expectedFuncs := []string{"read", "json", "yaml", "toml", "csv", "xml"}

	for _, name := range expectedFuncs {
		if _, ok := funcs[name]; !ok {
			t.Errorf("expected '%s' in FileOperations()", name)
		}
	}
}

// ----------------------------------------------------------------------------
// ConvertToValue Tests
// ----------------------------------------------------------------------------

func TestConvertToValueNil(t *testing.T) {
	result, err := convertToValue(nil)
	if err != nil {
		t.Fatalf("convertToValue(nil) error: %v", err)
	}
	if !IsNone(result) {
		t.Errorf("expected None, got %T", result)
	}
}

func TestConvertToValueBool(t *testing.T) {
	result, err := convertToValue(true)
	if err != nil {
		t.Fatalf("convertToValue(true) error: %v", err)
	}
	if v, ok := result.(BoolValue); !ok || bool(v) != true {
		t.Errorf("expected BoolValue(true), got %v", result)
	}
}

func TestConvertToValueInt(t *testing.T) {
	result, err := convertToValue(42)
	if err != nil {
		t.Fatalf("convertToValue(42) error: %v", err)
	}
	if v, ok := result.(IntValue); !ok || int64(v) != 42 {
		t.Errorf("expected IntValue(42), got %v", result)
	}
}

func TestConvertToValueFloat(t *testing.T) {
	result, err := convertToValue(3.14)
	if err != nil {
		t.Fatalf("convertToValue(3.14) error: %v", err)
	}
	if v, ok := result.(FloatValue); !ok || float64(v) != 3.14 {
		t.Errorf("expected FloatValue(3.14), got %v", result)
	}
}

func TestConvertToValueString(t *testing.T) {
	result, err := convertToValue("hello")
	if err != nil {
		t.Fatalf("convertToValue(\"hello\") error: %v", err)
	}
	if v, ok := result.(StrValue); !ok || string(v) != "hello" {
		t.Errorf("expected StrValue(\"hello\"), got %v", result)
	}
}

func TestConvertToValueNestedStructure(t *testing.T) {
	input := map[string]interface{}{
		"array": []interface{}{1, 2, 3},
		"nested": map[string]interface{}{
			"key": "value",
		},
	}

	result, err := convertToValue(input)
	if err != nil {
		t.Fatalf("convertToValue error: %v", err)
	}

	dict, ok := result.(DictValue)
	if !ok {
		t.Fatalf("expected DictValue, got %T", result)
	}

	// Check array
	arrVal, ok := dict.Get("array")
	if !ok {
		t.Error("expected 'array' key")
	} else {
		arr, ok := arrVal.(ArrayValue)
		if !ok {
			t.Fatalf("expected ArrayValue, got %T", arrVal)
		}
		if len(arr) != 3 {
			t.Errorf("expected 3 elements, got %d", len(arr))
		}
	}

	// Check nested
	nestedVal, ok := dict.Get("nested")
	if !ok {
		t.Error("expected 'nested' key")
	} else {
		nested, ok := nestedVal.(DictValue)
		if !ok {
			t.Fatalf("expected DictValue, got %T", nestedVal)
		}
		keyVal, ok := nested.Get("key")
		if !ok {
			t.Error("expected 'key' in nested dict")
		} else if str, ok := keyVal.(StrValue); !ok || string(str) != "value" {
			t.Errorf("nested.key = %v, want 'value'", keyVal)
		}
	}
}

// ----------------------------------------------------------------------------
// Integration Tests with Real Files (Optional)
// ----------------------------------------------------------------------------

func TestFileOperationsWithRealFiles(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "fileops_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := map[string]string{
		"test.txt":   "Hello, World!",
		"test.json":  `{"key": "value"}`,
		"test.yaml":  "key: value\n",
		"test.toml":  "key = \"value\"\n",
		"test.csv":   "a,b,c\n1,2,3\n",
		"test.xml":   "<root><item>test</item></root>",
	}

	for name, content := range testFiles {
		err := os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	// Create a FileWorld
	world, err := NewFileWorld(tempDir, "test.txt")
	if err != nil {
		t.Fatalf("failed to create FileWorld: %v", err)
	}

	vm := newTestVm(world)

	// Test read()
	args := NewArgs(syntax.Detached())
	args.Push(Str("test.txt"), syntax.Detached())
	result, err := readNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Errorf("read() error: %v", err)
	} else if str, ok := result.(StrValue); !ok || string(str) != "Hello, World!" {
		t.Errorf("read() = %v, want 'Hello, World!'", result)
	}

	// Test json()
	args = NewArgs(syntax.Detached())
	args.Push(Str("test.json"), syntax.Detached())
	result, err = jsonNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Errorf("json() error: %v", err)
	} else if dict, ok := result.(DictValue); ok {
		if v, ok := dict.Get("key"); ok {
			if str, ok := v.(StrValue); !ok || string(str) != "value" {
				t.Errorf("json().key = %v, want 'value'", v)
			}
		} else {
			t.Error("json() missing 'key'")
		}
	}

	// Test yaml()
	args = NewArgs(syntax.Detached())
	args.Push(Str("test.yaml"), syntax.Detached())
	result, err = yamlNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Errorf("yaml() error: %v", err)
	}

	// Test toml()
	args = NewArgs(syntax.Detached())
	args.Push(Str("test.toml"), syntax.Detached())
	result, err = tomlNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Errorf("toml() error: %v", err)
	}

	// Test csv()
	args = NewArgs(syntax.Detached())
	args.Push(Str("test.csv"), syntax.Detached())
	result, err = csvNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Errorf("csv() error: %v", err)
	}

	// Test xml()
	args = NewArgs(syntax.Detached())
	args.Push(Str("test.xml"), syntax.Detached())
	result, err = xmlNative(vm.Engine, vm.Context, args)
	if err != nil {
		t.Errorf("xml() error: %v", err)
	}
}

// ----------------------------------------------------------------------------
// Error Type Tests
// ----------------------------------------------------------------------------

func TestFileReadError(t *testing.T) {
	err := &FileReadError{
		Path:    "/path/to/file.txt",
		Message: "file not found",
	}

	expected := "cannot read file '/path/to/file.txt': file not found"
	if err.Error() != expected {
		t.Errorf("error = %q, want %q", err.Error(), expected)
	}
}

func TestFileParseError(t *testing.T) {
	err := &FileParseError{
		Path:    "/path/to/file.json",
		Format:  "JSON",
		Message: "unexpected end of input",
	}

	expected := "cannot parse JSON file '/path/to/file.json': unexpected end of input"
	if err.Error() != expected {
		t.Errorf("error = %q, want %q", err.Error(), expected)
	}
}
