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

const pagesDirName = "pages"

type IndexEntry struct {
	Title string
	URL   string
}

type IndexData struct {
	Title   string
	Entries []IndexEntry
}

func Build(cfg config.Config) error {
	pageTemplate, err := loadPageTemplate(cfg)
	if err != nil {
		return err
	}

	indexTemplate, err := loadIndexTemplate(cfg)
	if err != nil {
		return err
	}
	var indexEntries []IndexEntry

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	entries, err := os.ReadDir(cfg.ContentDir)
	if err != nil {
		return fmt.Errorf("read content dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		page, outputName, err := buildPage(cfg, pageTemplate, entry.Name())
		if err != nil {
			return err
		}

		indexEntries = append(indexEntries, IndexEntry{
			Title: page.Title,
			URL:   filepath.ToSlash(filepath.Join(pagesDirName, outputName)),
		})
	}

	if err := buildIndex(cfg, indexTemplate, indexEntries); err != nil {
		return err
	}

	return nil
}

func loadPageTemplate(cfg config.Config) (*template.Template, error) {
	path := filepath.Join(cfg.TemplateDir, cfg.PageTemplate)

	tmpl, err := template.New(cfg.PageTemplate).
		Funcs(template.FuncMap{
			"markdown": render.MarkdownToHTML,
		}).
		ParseFiles(path)

	if err != nil {
		return nil, fmt.Errorf("parse page template: %w", err)
	}

	return tmpl, nil
}

func loadIndexTemplate(cfg config.Config) (*template.Template, error) {
	path := filepath.Join(cfg.TemplateDir, cfg.IndexTemplate)

	tmpl, err := template.New(cfg.IndexTemplate).
		Funcs(template.FuncMap{
			"markdown": render.MarkdownToHTML,
		}).
		ParseFiles(path)

	if err != nil {
		return nil, fmt.Errorf("parse index template: %w", err)
	}

	return tmpl, nil
}
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
