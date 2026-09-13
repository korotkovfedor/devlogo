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
	page content.Content,
	slug string,
) error {
	html, err := render.ToHTML(page, tmpl)
	if err != nil {
		return fmt.Errorf("render page %q: %w", slug, err)
	}

	outputPath := filepath.Join(
		cfg.OutputDir,
		pagesDirName,
		slug+".html",
	)

	if err := os.WriteFile(outputPath, html, 0644); err != nil {
		return fmt.Errorf("write %q: %w", outputPath, err)
	}

	return nil
}
