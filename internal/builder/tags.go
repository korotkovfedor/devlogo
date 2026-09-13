package builder

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/korotkovfedor/devlogo/internal/config"
	"github.com/korotkovfedor/devlogo/internal/render"
)

type TagPageData struct {
	Tag     string
	Entries []IndexEntry
}

func buildTags(
	cfg config.Config,
	tmpl *template.Template,
	tags map[string][]IndexEntry,
) error {
	tagsDir := filepath.Join(cfg.OutputDir, tagsDirName)

	for tag, entries := range tags {
		data := TagPageData{
			Tag:     tag,
			Entries: entries,
		}

		html, err := render.ToHTML(data, tmpl)
		if err != nil {
			return fmt.Errorf("render tag %q: %w", tag, err)
		}

		name := tag + ".html"
		path := filepath.Join(tagsDir, name)

		if err := os.WriteFile(path, html, 0644); err != nil {
			return fmt.Errorf("write tag %q: %w", tag, err)
		}
	}

	return nil
}
