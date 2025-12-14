// Package output provides output formatting for sekai-cli.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Format represents an output format type.
type Format string

const (
	// FormatText is human-readable text format.
	FormatText Format = "text"

	// FormatJSON is JSON format.
	FormatJSON Format = "json"

	// FormatYAML is YAML format.
	FormatYAML Format = "yaml"
)

// Formatter formats data for output.
type Formatter interface {
	// Format formats the data and writes to the writer.
	Format(w io.Writer, data interface{}) error

	// FormatString formats the data and returns as string.
	FormatString(data interface{}) (string, error)
}

// NewFormatter creates a formatter for the given format.
func NewFormatter(format Format) Formatter {
	switch format {
	case FormatJSON:
		return &JSONFormatter{Indent: true}
	case FormatYAML:
		return &YAMLFormatter{}
	default:
		return &TextFormatter{}
	}
}

// NewFormatterFromString creates a formatter from a format string.
func NewFormatterFromString(format string) Formatter {
	return NewFormatter(Format(strings.ToLower(format)))
}

// TextFormatter formats data as human-readable text.
type TextFormatter struct {
	// Indent is the indentation string for nested structures.
	Indent string
}

// Format formats data as text.
func (f *TextFormatter) Format(w io.Writer, data interface{}) error {
	s, err := f.FormatString(data)
	if err != nil {
		return err
	}
	// Ensure output ends with a newline
	if s != "" && !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	_, err = fmt.Fprint(w, s)
	return err
}

// FormatString formats data as text string.
func (f *TextFormatter) FormatString(data interface{}) (string, error) {
	if data == nil {
		return "", nil
	}

	// Handle string directly
	if s, ok := data.(string); ok {
		return s, nil
	}

	// Handle Stringer interface
	if s, ok := data.(fmt.Stringer); ok {
		return s.String(), nil
	}

	// Handle json.RawMessage as JSON string
	if raw, ok := data.(json.RawMessage); ok {
		// Try to pretty-print the JSON
		var parsed interface{}
		if err := json.Unmarshal(raw, &parsed); err == nil {
			return f.formatValue(reflect.ValueOf(parsed), 0), nil
		}
		return string(raw), nil
	}

	// Handle byte slice as string
	if b, ok := data.([]byte); ok {
		return string(b), nil
	}

	// Use reflection for structs and maps
	return f.formatValue(reflect.ValueOf(data), 0), nil
}

func (f *TextFormatter) formatValue(v reflect.Value, depth int) string {
	if !v.IsValid() {
		return ""
	}

	indent := f.Indent
	if indent == "" {
		indent = "  "
	}
	prefix := strings.Repeat(indent, depth)

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return ""
		}
		return f.formatValue(v.Elem(), depth)

	case reflect.Struct:
		var sb strings.Builder
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			name := field.Name
			// Use json tag if available
			if tag := field.Tag.Get("json"); tag != "" {
				parts := strings.Split(tag, ",")
				if parts[0] != "" && parts[0] != "-" {
					name = parts[0]
				}
			}

			fieldValue := v.Field(i)
			if isEmptyValue(fieldValue) {
				continue
			}

			formatted := f.formatValue(fieldValue, depth+1)
			if formatted == "" {
				continue
			}

			if fieldValue.Kind() == reflect.Struct || fieldValue.Kind() == reflect.Map ||
				(fieldValue.Kind() == reflect.Slice && fieldValue.Len() > 0) {
				sb.WriteString(fmt.Sprintf("%s%s:\n%s", prefix, name, formatted))
			} else {
				sb.WriteString(fmt.Sprintf("%s%s: %s\n", prefix, name, formatted))
			}
		}
		return sb.String()

	case reflect.Map:
		var sb strings.Builder
		iter := v.MapRange()
		for iter.Next() {
			key := fmt.Sprintf("%v", iter.Key().Interface())
			val := f.formatValue(iter.Value(), depth+1)
			if val == "" {
				continue
			}
			if iter.Value().Kind() == reflect.Struct || iter.Value().Kind() == reflect.Map {
				sb.WriteString(fmt.Sprintf("%s%s:\n%s", prefix, key, val))
			} else {
				sb.WriteString(fmt.Sprintf("%s%s: %s\n", prefix, key, val))
			}
		}
		return sb.String()

	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return ""
		}
		var sb strings.Builder
		for i := 0; i < v.Len(); i++ {
			item := f.formatValue(v.Index(i), depth)
			if item == "" {
				continue
			}
			// Remove leading whitespace for list items
			item = strings.TrimPrefix(item, prefix)
			sb.WriteString(fmt.Sprintf("%s- %s", prefix, item))
			if !strings.HasSuffix(item, "\n") {
				sb.WriteString("\n")
			}
		}
		return sb.String()

	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

// JSONFormatter formats data as JSON.
type JSONFormatter struct {
	// Indent enables pretty-printing with indentation.
	Indent bool
}

// Format formats data as JSON.
func (f *JSONFormatter) Format(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	if f.Indent {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}

// FormatString formats data as JSON string.
func (f *JSONFormatter) FormatString(data interface{}) (string, error) {
	var b []byte
	var err error

	if f.Indent {
		b, err = json.MarshalIndent(data, "", "  ")
	} else {
		b, err = json.Marshal(data)
	}

	if err != nil {
		return "", err
	}
	return string(b), nil
}

// YAMLFormatter formats data as YAML.
// This is a simple implementation without external dependencies.
type YAMLFormatter struct{}

// Format formats data as YAML.
func (f *YAMLFormatter) Format(w io.Writer, data interface{}) error {
	s, err := f.FormatString(data)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, s)
	return err
}

// FormatString formats data as YAML string.
func (f *YAMLFormatter) FormatString(data interface{}) (string, error) {
	if data == nil {
		return "", nil
	}

	return f.formatValue(reflect.ValueOf(data), 0), nil
}

func (f *YAMLFormatter) formatValue(v reflect.Value, depth int) string {
	if !v.IsValid() {
		return ""
	}

	prefix := strings.Repeat("  ", depth)

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return "null\n"
		}
		return f.formatValue(v.Elem(), depth)

	case reflect.Struct:
		var sb strings.Builder
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			name := field.Name
			// Use yaml tag first, then json tag
			if tag := field.Tag.Get("yaml"); tag != "" {
				parts := strings.Split(tag, ",")
				if parts[0] != "" && parts[0] != "-" {
					name = parts[0]
				}
			} else if tag := field.Tag.Get("json"); tag != "" {
				parts := strings.Split(tag, ",")
				if parts[0] != "" && parts[0] != "-" {
					name = parts[0]
				}
			}

			fieldValue := v.Field(i)
			if isEmptyValue(fieldValue) {
				continue
			}

			switch fieldValue.Kind() {
			case reflect.Struct, reflect.Map:
				sb.WriteString(fmt.Sprintf("%s%s:\n", prefix, name))
				sb.WriteString(f.formatValue(fieldValue, depth+1))
			case reflect.Slice, reflect.Array:
				if fieldValue.Len() == 0 {
					continue
				}
				sb.WriteString(fmt.Sprintf("%s%s:\n", prefix, name))
				sb.WriteString(f.formatValue(fieldValue, depth+1))
			default:
				sb.WriteString(fmt.Sprintf("%s%s: %s\n", prefix, name, f.formatScalar(fieldValue)))
			}
		}
		return sb.String()

	case reflect.Map:
		var sb strings.Builder
		iter := v.MapRange()
		for iter.Next() {
			key := fmt.Sprintf("%v", iter.Key().Interface())
			val := iter.Value()
			switch val.Kind() {
			case reflect.Struct, reflect.Map:
				sb.WriteString(fmt.Sprintf("%s%s:\n", prefix, key))
				sb.WriteString(f.formatValue(val, depth+1))
			case reflect.Slice, reflect.Array:
				sb.WriteString(fmt.Sprintf("%s%s:\n", prefix, key))
				sb.WriteString(f.formatValue(val, depth+1))
			default:
				sb.WriteString(fmt.Sprintf("%s%s: %s\n", prefix, key, f.formatScalar(val)))
			}
		}
		return sb.String()

	case reflect.Slice, reflect.Array:
		var sb strings.Builder
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			switch item.Kind() {
			case reflect.Struct, reflect.Map:
				sb.WriteString(fmt.Sprintf("%s-\n", prefix))
				// Indent the nested structure
				nested := f.formatValue(item, depth+1)
				sb.WriteString(nested)
			default:
				sb.WriteString(fmt.Sprintf("%s- %s\n", prefix, f.formatScalar(item)))
			}
		}
		return sb.String()

	default:
		return f.formatScalar(v) + "\n"
	}
}

func (f *YAMLFormatter) formatScalar(v reflect.Value) string {
	if !v.IsValid() {
		return "null"
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		// Quote strings that might be ambiguous
		if needsQuoting(s) {
			return fmt.Sprintf("%q", s)
		}
		return s
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", v.Float())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

// isEmptyValue checks if a value is empty/zero.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// needsQuoting checks if a string needs YAML quoting.
func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	// Check for special characters or reserved words
	special := []string{"true", "false", "null", "yes", "no", "on", "off"}
	lower := strings.ToLower(s)
	for _, sp := range special {
		if lower == sp {
			return true
		}
	}
	// Check for special characters
	for _, r := range s {
		if r == ':' || r == '#' || r == '\n' || r == '\t' || r == '"' || r == '\'' {
			return true
		}
	}
	// Check if starts with special characters
	if s[0] == '-' || s[0] == '?' || s[0] == '*' || s[0] == '&' || s[0] == '!' ||
		s[0] == '|' || s[0] == '>' || s[0] == '%' || s[0] == '@' || s[0] == '`' {
		return true
	}
	return false
}

// Print prints data using the specified format.
func Print(w io.Writer, format Format, data interface{}) error {
	return NewFormatter(format).Format(w, data)
}

// PrintJSON prints data as JSON.
func PrintJSON(w io.Writer, data interface{}) error {
	return Print(w, FormatJSON, data)
}

// PrintYAML prints data as YAML.
func PrintYAML(w io.Writer, data interface{}) error {
	return Print(w, FormatYAML, data)
}

// PrintText prints data as text.
func PrintText(w io.Writer, data interface{}) error {
	return Print(w, FormatText, data)
}
