package builder

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/korotkovfedor/devlogo/internal/config"
	"github.com/korotkovfedor/devlogo/internal/content"
	"github.com/korotkovfedor/devlogo/internal/render"
)

func buildPage(
	cfg config.Config,
	tmpl *template.Template,
	name string,
) (content.Content, string, error) {
	inputPath := filepath.Join(cfg.ContentDir, name)

	page, err := parseEntry(inputPath)
	if err != nil {
		return content.Content{}, "", fmt.Errorf("process %q: %w", inputPath, err)
	}

	html, err := render.ToHTML(page, tmpl)
	if err != nil {
		return content.Content{}, "", fmt.Errorf("render %q: %w", inputPath, err)
	}

	outputName := strings.TrimSuffix(name, filepath.Ext(name)) + ".html"
	pagesDir := filepath.Join(cfg.OutputDir, pagesDirName)

	if err := os.MkdirAll(pagesDir, 0755); err != nil {
		return content.Content{}, "", fmt.Errorf("create pages dir: %w", err)
	}
	outputPath := filepath.Join(pagesDir, outputName)

	if err := os.WriteFile(outputPath, html, 0644); err != nil {
		return content.Content{}, "", fmt.Errorf("write %q: %w", outputPath, err)
	}

	return page, outputName, nil
}
