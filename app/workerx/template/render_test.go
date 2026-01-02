package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	t.Run("simple template", func(t *testing.T) {
		tmpl := "Hello {{.Name}}"
		data := map[string]interface{}{
			"Name": "World",
		}

		result, err := Render(tmpl, data)
		assert.NoError(t, err)
		assert.Equal(t, "Hello World", result)
	})

	t.Run("template with multiple variables", func(t *testing.T) {
		tmpl := "Order {{.OrderID}} for {{.CustomerName}}"
		data := map[string]interface{}{
			"OrderID":     "12345",
			"CustomerName": "John Doe",
		}

		result, err := Render(tmpl, data)
		assert.NoError(t, err)
		assert.Equal(t, "Order 12345 for John Doe", result)
	})

	t.Run("template with missing variable", func(t *testing.T) {
		tmpl := "Hello {{.Name}}"
		data := map[string]interface{}{}

		result, err := Render(tmpl, data)
		assert.NoError(t, err)
		assert.Equal(t, "Hello <no value>", result)
	})

	t.Run("template with number", func(t *testing.T) {
		tmpl := "Total: ${{.Amount}}"
		data := map[string]interface{}{
			"Amount": 99.99,
		}

		result, err := Render(tmpl, data)
		assert.NoError(t, err)
		assert.Equal(t, "Total: $99.99", result)
	})

	t.Run("invalid template syntax", func(t *testing.T) {
		tmpl := "Hello {{.Name" // Missing closing brace
		data := map[string]interface{}{
			"Name": "World",
		}

		_, err := Render(tmpl, data)
		assert.Error(t, err)
	})

	t.Run("empty template", func(t *testing.T) {
		tmpl := ""
		data := map[string]interface{}{
			"Name": "World",
		}

		result, err := Render(tmpl, data)
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("template with conditional logic", func(t *testing.T) {
		tmpl := "{{if .Enabled}}Enabled{{else}}Disabled{{end}}"
		data := map[string]interface{}{
			"Enabled": true,
		}

		result, err := Render(tmpl, data)
		assert.NoError(t, err)
		assert.Equal(t, "Enabled", result)
	})

	t.Run("template with range", func(t *testing.T) {
		tmpl := "Items: {{range .Items}}{{.}}{{end}}"
		data := map[string]interface{}{
			"Items": []string{"apple", "banana", "cherry"},
		}

		result, err := Render(tmpl, data)
		assert.NoError(t, err)
		assert.Contains(t, result, "Items:")
	})
}

