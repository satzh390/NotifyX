package template

import (
	"bytes"
	"text/template"
)

// Render renders a template string with the given data
func Render(tmplStr string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("template").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

