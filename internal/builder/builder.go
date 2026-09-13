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

const pagesDirName = "pages"
const tagsDirName = "tags"

type IndexEntry struct {
	Title string
	URL   string
}

func Build(cfg config.Config) error {
	pageTemplate, err := loadTemplate(cfg, cfg.PageTemplate)
	if err != nil {
		return err
	}

	indexTemplate, err := loadTemplate(cfg, cfg.IndexTemplate)
	if err != nil {
		return err
	}

	tagsTemplate, err := loadTemplate(cfg, cfg.TagsTemplate)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	entries, err := os.ReadDir(cfg.ContentDir)
	if err != nil {
		return fmt.Errorf("read content dir: %w", err)
	}

	var indexEntries []IndexEntry
	tags := make(map[string][]IndexEntry)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		page, outputName, err := buildPage(cfg, pageTemplate, entry.Name())
		if err != nil {
			return err
		}

		indexEntry := IndexEntry{
			Title: page.Title,
			URL:   filepath.ToSlash(filepath.Join("pages", outputName)),
		}

		indexEntries = append(indexEntries, indexEntry)

		for _, tag := range page.Tags {
			tags[tag] = append(tags[tag], indexEntry)
		}
	}

	if err := buildIndex(cfg, indexTemplate, indexEntries); err != nil {
		return err
	}

	if err := buildTags(cfg, tagsTemplate, tags); err != nil {

	}

	return nil
}

func loadTemplate(
	cfg config.Config,
	name string,
) (*template.Template, error) {
	path := filepath.Join(cfg.TemplateDir, name)

	tmpl, err := template.New(name).
		Funcs(template.FuncMap{
			"markdown": render.MarkdownToHTML,
			"tagSlug":  tagSlug,
		}).
		ParseFiles(path)
	if err != nil {
		return nil, fmt.Errorf("parse template %q: %w", name, err)
	}

	return tmpl, nil
}

func parseEntry(path string) (content.Content, error) {
	file, err := os.Open(path)
	if err != nil {
		return content.Content{}, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	result, err := content.ParseMarkdown(file)
	if err != nil {
		return content.Content{}, fmt.Errorf("parse markdown: %w", err)
	}

	return result, nil
}
