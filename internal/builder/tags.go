package builder

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

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

	if err := os.MkdirAll(tagsDir, 0755); err != nil {
		return fmt.Errorf("create tags dir: %w", err)
	}

	for tag, entries := range tags {
		data := TagPageData{
			Tag:     tag,
			Entries: entries,
		}

		html, err := render.ToHTML(data, tmpl)
		if err != nil {
			return fmt.Errorf("render tag %q: %w", tag, err)
		}

		name := tagSlug(tag) + ".html"
		path := filepath.Join(tagsDir, name)

		if err := os.WriteFile(path, html, 0644); err != nil {
			return fmt.Errorf("write tag %q: %w", tag, err)
		}
	}

	return nil
}

func tagSlug(tag string) string {
	return strings.ToLower(
		strings.ReplaceAll(strings.TrimSpace(tag), " ", "-"),
	)
}
