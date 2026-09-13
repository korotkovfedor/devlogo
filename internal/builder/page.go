package builder

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/korotkovfedor/devlogo/internal/config"
	"github.com/korotkovfedor/devlogo/internal/content"
	"github.com/korotkovfedor/devlogo/internal/render"
)

func buildPage(
	cfg config.Config,
	tmpl *template.Template,
	name string,
	slug string,
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

	outputName := slug + ".html"
	outputPath := filepath.Join(cfg.OutputDir, pagesDirName, outputName)

	if err := os.WriteFile(outputPath, html, 0644); err != nil {
		return content.Content{}, "", fmt.Errorf("write %q: %w", outputPath, err)
	}

	return page, outputName, nil
}
