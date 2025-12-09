package render

import (
	"fmt"
	"os"
	"reflect"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/term"
)

// TableRenderer renders slice types as tables
type TableRenderer struct {
	columns []string
	rows    [][]string
}

// NewTableRenderer creates a table renderer from a slice of structs
func NewTableRenderer(data any) (*TableRenderer, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected slice, got %s", v.Kind())
	}

	if v.Len() == 0 {
		return &TableRenderer{}, nil
	}

	// Get columns from first element's struct fields
	first := v.Index(0)
	if first.Kind() == reflect.Ptr {
		first = first.Elem()
	}

	if first.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct elements, got %s", first.Kind())
	}

	t := first.Type()
	var columns []string
	var fieldIndices []int

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// Skip unexported fields and complex types
		if !field.IsExported() {
			continue
		}
		kind := field.Type.Kind()
		if kind == reflect.Struct || kind == reflect.Slice || kind == reflect.Map || kind == reflect.Ptr {
			// Skip complex types for table display
			continue
		}
		columns = append(columns, field.Name)
		fieldIndices = append(fieldIndices, i)
	}

	// Build rows
	var rows [][]string
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		var row []string
		for _, idx := range fieldIndices {
			val := elem.Field(idx)
			row = append(row, formatValue(val))
		}
		rows = append(rows, row)
	}

	return &TableRenderer{
		columns: columns,
		rows:    rows,
	}, nil
}

// getTerminalWidth returns the terminal width or a default value
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 120 // default fallback
	}
	return width
}

// Render renders the table using lipgloss/table with terminal-adaptive width
func (tr *TableRenderer) Render() string {
	if len(tr.columns) == 0 {
		return SubtitleStyle.Render("No data to display")
	}

	termWidth := getTerminalWidth()

	// Create table with lipgloss/table - it handles width distribution automatically
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(borderColor)).
		Width(termWidth).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().
					Bold(true).
					Foreground(primaryColor).
					Padding(0, 1)
			}
			// Alternate row colors for better readability
			if row%2 == 0 {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("#9CA3AF")).
					Padding(0, 1)
			}
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D1D5DB")).
				Padding(0, 1)
		}).
		Headers(tr.columns...).
		Rows(tr.rows...)

	return t.Render()
}

// RenderSimple renders a simple table (kept for compatibility)
func (tr *TableRenderer) RenderSimple() string {
	return tr.Render()
}

func formatValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", v.Float())
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case reflect.Ptr:
		if v.IsNil() {
			return "-"
		}
		return formatValue(v.Elem())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}
