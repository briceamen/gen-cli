package render

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// DetailRenderer renders a single struct as a key-value detail view
type DetailRenderer struct {
	title  string
	fields []fieldPair
}

type fieldPair struct {
	key   string
	value string
}

// NewDetailRenderer creates a detail renderer from a struct
func NewDetailRenderer(data any, title string) (*DetailRenderer, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, fmt.Errorf("nil pointer")
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", v.Kind())
	}

	dr := &DetailRenderer{
		title: title,
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		val := v.Field(i)
		formatted := formatDetailValue(val, 0)
		if formatted != "" {
			dr.fields = append(dr.fields, fieldPair{
				key:   field.Name,
				value: formatted,
			})
		}
	}

	return dr, nil
}

// Render renders the detail view to a string
func (dr *DetailRenderer) Render() string {
	var sb strings.Builder

	if dr.title != "" {
		sb.WriteString(TitleStyle.Render(dr.title))
		sb.WriteString("\n\n")
	}

	maxKeyLen := 0
	for _, f := range dr.fields {
		if len(f.key) > maxKeyLen {
			maxKeyLen = len(f.key)
		}
	}

	for _, f := range dr.fields {
		key := KeyStyle.Width(maxKeyLen + 2).Render(f.key + ":")
		value := ValueStyle.Render(f.value)
		sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, key, value))
		sb.WriteString("\n")
	}

	return BoxStyle.Render(sb.String())
}

func formatDetailValue(v reflect.Value, depth int) string {
	if depth > 2 {
		return "..."
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if s == "" {
			return SubtitleStyle.Render("(empty)")
		}
		return s

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())

	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", v.Float())

	case reflect.Bool:
		if v.Bool() {
			return SuccessStyle.Render("true")
		}
		return ErrorStyle.Render("false")

	case reflect.Ptr:
		if v.IsNil() {
			return SubtitleStyle.Render("(nil)")
		}
		return formatDetailValue(v.Elem(), depth)

	case reflect.Slice:
		if v.Len() == 0 {
			return SubtitleStyle.Render("(empty)")
		}
		if v.Len() > 5 {
			return fmt.Sprintf("[%d items]", v.Len())
		}
		var items []string
		for i := 0; i < v.Len(); i++ {
			items = append(items, formatDetailValue(v.Index(i), depth+1))
		}
		return strings.Join(items, ", ")

	case reflect.Map:
		if v.Len() == 0 {
			return SubtitleStyle.Render("(empty)")
		}
		return fmt.Sprintf("{%d entries}", v.Len())

	case reflect.Struct:
		// Handle time.Time specially
		if v.Type() == reflect.TypeOf(time.Time{}) {
			t := v.Interface().(time.Time)
			if t.IsZero() {
				return SubtitleStyle.Render("(not set)")
			}
			return t.Format(time.RFC3339)
		}
		// For other structs, show type name
		return fmt.Sprintf("<%s>", v.Type().Name())

	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}
