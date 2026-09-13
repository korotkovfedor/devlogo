package render

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/korotkovfedor/devlogo/internal/content"
	"github.com/yuin/goldmark"
)

func MarkdownToHTML(md content.Markdown) (template.HTML, error) {
	var buf bytes.Buffer

	if err := goldmark.Convert([]byte(md), &buf); err != nil {
		return "", fmt.Errorf("convert markdown: %w", err)
	}

	return template.HTML(buf.String()), nil
}

func ToHTML(data any, tmpl *template.Template) ([]byte, error) {
	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return buf.Bytes(), nil
}
