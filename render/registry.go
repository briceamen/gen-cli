package render

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
)

// OutputFormat specifies the output format
type OutputFormat string

const (
	FormatTable  OutputFormat = "table"
	FormatDetail OutputFormat = "detail"
	FormatJSON   OutputFormat = "json"
)

// RenderResult renders any result based on its type using convention-based mapping
// Returns the rendered string and any error
func RenderResult(data any, format OutputFormat) (string, error) {
	if format == FormatJSON {
		return renderJSON(data)
	}

	// Check for nil
	if data == nil {
		return RenderSuccess("Operation completed successfully"), nil
	}

	v := reflect.ValueOf(data)

	// Handle pointer to http.Response - read and return body content
	if v.Type() == reflect.TypeOf(&http.Response{}) {
		resp := data.(*http.Response)
		if resp.Body != nil {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", fmt.Errorf("failed to read response body: %w", err)
			}
			if len(body) > 0 {
				return string(body), nil
			}
		}
		// Fallback to status if no body
		return RenderHTTPStatus(resp.StatusCode), nil
	}

	// Dereference pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return RenderSuccess("Operation completed successfully"), nil
		}
		v = v.Elem()
	}

	// Convention-based type mapping
	switch v.Kind() {
	case reflect.Slice:
		// []Type or []*Type -> Table
		renderer, err := NewTableRenderer(data)
		if err != nil {
			return "", err
		}
		return renderer.RenderSimple(), nil

	case reflect.Struct:
		// *Type -> Detail view
		typeName := v.Type().Name()
		renderer, err := NewDetailRenderer(data, typeName)
		if err != nil {
			return "", err
		}
		return renderer.Render(), nil

	default:
		// Fallback to simple string representation
		return fmt.Sprintf("%v", data), nil
	}
}

// InferFormat infers the appropriate format from the return type string
func InferFormat(returnType string) OutputFormat {
	returnType = strings.TrimSpace(returnType)

	if strings.HasPrefix(returnType, "[]") {
		return FormatTable
	}
	if strings.HasPrefix(returnType, "*") || strings.HasPrefix(returnType, "map[") {
		return FormatDetail
	}
	return FormatDetail
}

func renderJSON(data any) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(b), nil
}
