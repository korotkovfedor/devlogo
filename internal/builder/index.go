package builder

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/korotkovfedor/devlogo/internal/config"
	"github.com/korotkovfedor/devlogo/internal/render"
)

type IndexData struct {
	Title   string
	Entries []IndexEntry
}

func buildIndex(
	cfg config.Config,
	tmpl *template.Template,
	entries []IndexEntry,
) error {
	data := IndexData{
		Title:   cfg.Title,
		Entries: entries,
	}

	html, err := render.ToHTML(data, tmpl)
	if err != nil {
		return fmt.Errorf("render index: %w", err)
	}

	outputPath := filepath.Join(cfg.OutputDir, "index.html")

	if err := os.WriteFile(outputPath, html, 0644); err != nil {
		return fmt.Errorf("write index: %w", err)
	}

	return nil
}
