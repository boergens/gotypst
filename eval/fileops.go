package eval

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/boergens/gotypst/syntax"
	"gopkg.in/yaml.v3"
)

// ----------------------------------------------------------------------------
// File Loading Functions
// ----------------------------------------------------------------------------

// This file contains data loading functions for reading and parsing
// various file formats: read(), json(), yaml(), toml(), csv(), xml().

// ReadFunc creates the read() function for raw file reading.
func ReadFunc() *Func {
	name := "read"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: readNative,
			Info: &FuncInfo{
				Name: "read",
				Params: []ParamInfo{
					{Name: "path", Type: TypeStr, Named: false},
					{Name: "encoding", Type: TypeStr, Default: Str("utf-8"), Named: true},
				},
			},
		},
	}
}

// readNative implements read(path, encoding: "utf-8").
// Reads a file and returns its contents as a string (for text) or bytes (for binary).
func readNative(engine *Engine, context *Context, args *Args) (Value, error) {
	// Get required path argument
	pathArg, err := args.Expect("path")
	if err != nil {
		return nil, err
	}

	path, ok := AsStr(pathArg.V)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "string",
			Got:      pathArg.V.Type().String(),
			Span:     pathArg.Span,
		}
	}

	// Get optional encoding argument (default: "utf-8")
	// If encoding is none (NoneValue), return as bytes (binary mode)
	encoding := "utf-8"
	binaryMode := false
	if encArg := args.Find("encoding"); encArg != nil {
		if IsNone(encArg.V) {
			binaryMode = true
		} else {
			encStr, ok := AsStr(encArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string or none",
					Got:      encArg.V.Type().String(),
					Span:     encArg.Span,
				}
			}
			encoding = encStr
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Resolve and read the file
	data, err := readFileFromWorld(engine, path)
	if err != nil {
		return nil, err
	}

	// Handle encoding
	if binaryMode {
		return BytesValue(data), nil
	}
	switch strings.ToLower(encoding) {
	case "utf-8", "utf8":
		return Str(string(data)), nil
	case "binary":
		return BytesValue(data), nil
	default:
		return nil, &FileReadError{
			Path:    path,
			Message: fmt.Sprintf("unsupported encoding: %s", encoding),
		}
	}
}

// JsonFunc creates the json() function for parsing JSON files.
func JsonFunc() *Func {
	name := "json"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: jsonNative,
			Info: &FuncInfo{
				Name: "json",
				Params: []ParamInfo{
					{Name: "path", Type: TypeStr, Named: false},
				},
			},
		},
	}
}

// jsonNative implements json(path).
// Reads and parses a JSON file, returning a dictionary or array.
func jsonNative(engine *Engine, context *Context, args *Args) (Value, error) {
	// Get required path argument
	pathArg, err := args.Expect("path")
	if err != nil {
		return nil, err
	}

	path, ok := AsStr(pathArg.V)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "string",
			Got:      pathArg.V.Type().String(),
			Span:     pathArg.Span,
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Read the file
	data, err := readFileFromWorld(engine, path)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, &FileParseError{
			Path:    path,
			Format:  "JSON",
			Message: err.Error(),
		}
	}

	return convertToValue(raw)
}

// YamlFunc creates the yaml() function for parsing YAML files.
func YamlFunc() *Func {
	name := "yaml"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: yamlNative,
			Info: &FuncInfo{
				Name: "yaml",
				Params: []ParamInfo{
					{Name: "path", Type: TypeStr, Named: false},
				},
			},
		},
	}
}

// yamlNative implements yaml(path).
// Reads and parses a YAML file, returning a dictionary or array.
func yamlNative(engine *Engine, context *Context, args *Args) (Value, error) {
	// Get required path argument
	pathArg, err := args.Expect("path")
	if err != nil {
		return nil, err
	}

	path, ok := AsStr(pathArg.V)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "string",
			Got:      pathArg.V.Type().String(),
			Span:     pathArg.Span,
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Read the file
	data, err := readFileFromWorld(engine, path)
	if err != nil {
		return nil, err
	}

	// Parse YAML
	var raw interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, &FileParseError{
			Path:    path,
			Format:  "YAML",
			Message: err.Error(),
		}
	}

	return convertToValue(raw)
}

// TomlFunc creates the toml() function for parsing TOML files.
func TomlFunc() *Func {
	name := "toml"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: tomlNative,
			Info: &FuncInfo{
				Name: "toml",
				Params: []ParamInfo{
					{Name: "path", Type: TypeStr, Named: false},
				},
			},
		},
	}
}

// tomlNative implements toml(path).
// Reads and parses a TOML file, returning a dictionary.
func tomlNative(engine *Engine, context *Context, args *Args) (Value, error) {
	// Get required path argument
	pathArg, err := args.Expect("path")
	if err != nil {
		return nil, err
	}

	path, ok := AsStr(pathArg.V)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "string",
			Got:      pathArg.V.Type().String(),
			Span:     pathArg.Span,
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Read the file
	data, err := readFileFromWorld(engine, path)
	if err != nil {
		return nil, err
	}

	// Parse TOML
	var raw map[string]interface{}
	if _, err := toml.Decode(string(data), &raw); err != nil {
		return nil, &FileParseError{
			Path:    path,
			Format:  "TOML",
			Message: err.Error(),
		}
	}

	return convertToValue(raw)
}

// CsvFunc creates the csv() function for parsing CSV files.
func CsvFunc() *Func {
	name := "csv"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: csvNative,
			Info: &FuncInfo{
				Name: "csv",
				Params: []ParamInfo{
					{Name: "path", Type: TypeStr, Named: false},
					{Name: "delimiter", Type: TypeStr, Default: Str(","), Named: true},
					{Name: "row-type", Type: TypeStr, Default: Str("array"), Named: true},
				},
			},
		},
	}
}

// csvNative implements csv(path, delimiter: ",", row-type: "array").
// Reads and parses a CSV file, returning an array of arrays or dictionaries.
func csvNative(engine *Engine, context *Context, args *Args) (Value, error) {
	// Get required path argument
	pathArg, err := args.Expect("path")
	if err != nil {
		return nil, err
	}

	path, ok := AsStr(pathArg.V)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "string",
			Got:      pathArg.V.Type().String(),
			Span:     pathArg.Span,
		}
	}

	// Get optional delimiter argument (default: ",")
	delimiter := ','
	if delimArg := args.Find("delimiter"); delimArg != nil {
		if !IsNone(delimArg.V) {
			delimStr, ok := AsStr(delimArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string or none",
					Got:      delimArg.V.Type().String(),
					Span:     delimArg.Span,
				}
			}
			if len(delimStr) != 1 {
				return nil, &FileParseError{
					Path:    path,
					Format:  "CSV",
					Message: "delimiter must be a single character",
				}
			}
			delimiter = rune(delimStr[0])
		}
	}

	// Get optional row-type argument (default: "array")
	rowType := "array"
	if rowTypeArg := args.Find("row-type"); rowTypeArg != nil {
		if !IsNone(rowTypeArg.V) {
			rowTypeStr, ok := AsStr(rowTypeArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string or none",
					Got:      rowTypeArg.V.Type().String(),
					Span:     rowTypeArg.Span,
				}
			}
			rowType = rowTypeStr
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Read the file
	data, err := readFileFromWorld(engine, path)
	if err != nil {
		return nil, err
	}

	// Parse CSV
	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = delimiter

	records, err := reader.ReadAll()
	if err != nil {
		return nil, &FileParseError{
			Path:    path,
			Format:  "CSV",
			Message: err.Error(),
		}
	}

	// Convert based on row-type
	switch rowType {
	case "array":
		return convertCSVToArrays(records), nil
	case "dict", "dictionary":
		if len(records) == 0 {
			return ArrayValue{}, nil
		}
		return convertCSVToDicts(records), nil
	default:
		return nil, &FileParseError{
			Path:    path,
			Format:  "CSV",
			Message: fmt.Sprintf("invalid row-type: %s (expected \"array\" or \"dict\")", rowType),
		}
	}
}

// XmlFunc creates the xml() function for parsing XML files.
func XmlFunc() *Func {
	name := "xml"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: xmlNative,
			Info: &FuncInfo{
				Name: "xml",
				Params: []ParamInfo{
					{Name: "path", Type: TypeStr, Named: false},
				},
			},
		},
	}
}

// xmlNative implements xml(path).
// Reads and parses an XML file, returning a nested dictionary structure.
func xmlNative(engine *Engine, context *Context, args *Args) (Value, error) {
	// Get required path argument
	pathArg, err := args.Expect("path")
	if err != nil {
		return nil, err
	}

	path, ok := AsStr(pathArg.V)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "string",
			Got:      pathArg.V.Type().String(),
			Span:     pathArg.Span,
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Read the file
	data, err := readFileFromWorld(engine, path)
	if err != nil {
		return nil, err
	}

	// Parse XML
	result, err := parseXML(data)
	if err != nil {
		return nil, &FileParseError{
			Path:    path,
			Format:  "XML",
			Message: err.Error(),
		}
	}

	return result, nil
}

// ----------------------------------------------------------------------------
// Helper Functions
// ----------------------------------------------------------------------------

// readFileFromWorld reads a file using the Engine's World interface.
func readFileFromWorld(engine *Engine, path string) ([]byte, error) {
	// Handle absolute paths and relative paths
	var resolvedPath string
	if filepath.IsAbs(path) {
		resolvedPath = path
	} else {
		// Resolve relative to the main file's directory
		mainFile := engine.world.MainFile()
		if mainFile.Path != "" {
			dir := filepath.Dir(mainFile.Path)
			resolvedPath = filepath.Join(dir, path)
		} else {
			resolvedPath = path
		}
	}

	// Create FileID and read through World interface
	fileID := FileID{Path: resolvedPath}
	data, err := engine.world.File(fileID)
	if err != nil {
		return nil, &FileReadError{
			Path:    path,
			Message: err.Error(),
		}
	}

	return data, nil
}

// convertToValue converts a Go interface{} value to a Typst Value.
func convertToValue(v interface{}) (Value, error) {
	if v == nil {
		return None, nil
	}

	switch val := v.(type) {
	case bool:
		return Bool(val), nil
	case int:
		return Int(int64(val)), nil
	case int64:
		return Int(val), nil
	case float64:
		// Check if it's actually an integer
		if val == float64(int64(val)) {
			return Int(int64(val)), nil
		}
		return Float(val), nil
	case string:
		return Str(val), nil
	case []interface{}:
		arr := make(ArrayValue, len(val))
		for i, elem := range val {
			v, err := convertToValue(elem)
			if err != nil {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	case map[string]interface{}:
		dict := NewDict()
		for k, v := range val {
			converted, err := convertToValue(v)
			if err != nil {
				return nil, err
			}
			dict.Set(k, converted)
		}
		return dict, nil
	case map[interface{}]interface{}:
		// YAML sometimes produces this type
		dict := NewDict()
		for k, v := range val {
			keyStr, ok := k.(string)
			if !ok {
				keyStr = fmt.Sprintf("%v", k)
			}
			converted, err := convertToValue(v)
			if err != nil {
				return nil, err
			}
			dict.Set(keyStr, converted)
		}
		return dict, nil
	default:
		// Try to handle as a string representation
		return Str(fmt.Sprintf("%v", val)), nil
	}
}

// convertCSVToArrays converts CSV records to an array of arrays.
func convertCSVToArrays(records [][]string) Value {
	result := make(ArrayValue, len(records))
	for i, record := range records {
		row := make(ArrayValue, len(record))
		for j, field := range record {
			row[j] = Str(field)
		}
		result[i] = row
	}
	return result
}

// convertCSVToDicts converts CSV records to an array of dictionaries.
// The first row is used as the header/keys.
func convertCSVToDicts(records [][]string) Value {
	if len(records) == 0 {
		return ArrayValue{}
	}

	headers := records[0]
	result := make(ArrayValue, len(records)-1)

	for i := 1; i < len(records); i++ {
		dict := NewDict()
		for j, header := range headers {
			if j < len(records[i]) {
				dict.Set(header, Str(records[i][j]))
			} else {
				dict.Set(header, None)
			}
		}
		result[i-1] = dict
	}

	return result
}

// parseXML parses XML data into a Typst value structure.
// Returns a dictionary representing the XML document.
func parseXML(data []byte) (Value, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	return parseXMLElement(decoder, nil)
}

// parseXMLElement recursively parses XML elements.
func parseXMLElement(decoder *xml.Decoder, startToken *xml.StartElement) (Value, error) {
	if startToken == nil {
		// Find the root element
		for {
			token, err := decoder.Token()
			if err == io.EOF {
				return None, nil
			}
			if err != nil {
				return nil, err
			}

			if start, ok := token.(xml.StartElement); ok {
				return parseXMLElement(decoder, &start)
			}
		}
	}

	// Create dictionary for this element
	dict := NewDict()

	// Add tag name
	dict.Set("tag", Str(startToken.Name.Local))

	// Add attributes
	if len(startToken.Attr) > 0 {
		attrs := NewDict()
		for _, attr := range startToken.Attr {
			attrs.Set(attr.Name.Local, Str(attr.Value))
		}
		dict.Set("attrs", attrs)
	}

	// Parse children and text content
	var children ArrayValue
	var textContent strings.Builder

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			child, err := parseXMLElement(decoder, &t)
			if err != nil {
				return nil, err
			}
			children = append(children, child)

		case xml.EndElement:
			// End of current element
			if len(children) > 0 {
				dict.Set("children", children)
			}
			text := strings.TrimSpace(textContent.String())
			if text != "" {
				dict.Set("text", Str(text))
			}
			return dict, nil

		case xml.CharData:
			textContent.Write(t)

		case xml.Comment:
			// Ignore comments
		}
	}

	// Shouldn't reach here for well-formed XML
	return dict, nil
}

// ----------------------------------------------------------------------------
// Error Types
// ----------------------------------------------------------------------------

// FileReadError is returned when a file cannot be read.
type FileReadError struct {
	Path    string
	Message string
}

func (e *FileReadError) Error() string {
	return fmt.Sprintf("cannot read file '%s': %s", e.Path, e.Message)
}

// FileParseError is returned when a file cannot be parsed.
type FileParseError struct {
	Path    string
	Format  string
	Message string
}

func (e *FileParseError) Error() string {
	return fmt.Sprintf("cannot parse %s file '%s': %s", e.Format, e.Path, e.Message)
}

// ----------------------------------------------------------------------------
// Library Registration
// ----------------------------------------------------------------------------

// RegisterFileOperations registers all file loading functions in the given scope.
// Call this when setting up the standard library scope.
func RegisterFileOperations(scope *Scope) {
	scope.DefineFunc("read", ReadFunc())
	scope.DefineFunc("json", JsonFunc())
	scope.DefineFunc("yaml", YamlFunc())
	scope.DefineFunc("toml", TomlFunc())
	scope.DefineFunc("csv", CsvFunc())
	scope.DefineFunc("xml", XmlFunc())
}

// FileOperations returns a map of all file operation function names to their functions.
// This is useful for introspection and testing.
func FileOperations() map[string]*Func {
	return map[string]*Func{
		"read": ReadFunc(),
		"json": JsonFunc(),
		"yaml": YamlFunc(),
		"toml": TomlFunc(),
		"csv":  CsvFunc(),
		"xml":  XmlFunc(),
	}
}
