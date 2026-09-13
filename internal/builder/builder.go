package builder

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"unicode"

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
	if err := config.ValidatePaths(cfg); err != nil {
		return err
	}

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

	if err := os.RemoveAll(cfg.OutputDir); err != nil {
		return fmt.Errorf("remove old output dir: %w", err)
	}

	if err := prepareOutput(cfg); err != nil {
		return err
	}

	entries, err := os.ReadDir(cfg.ContentDir)
	if err != nil {
		return fmt.Errorf("read content dir: %w", err)
	}

	var indexEntries []IndexEntry
	tags := make(map[string][]IndexEntry)
	seen := make(map[string]string)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		name := entry.Name()
		slug := buildSlug(strings.TrimSuffix(name, filepath.Ext(name)))

		if previous, ok := seen[slug]; ok {
			return fmt.Errorf(
				"page slug collision: %q and %q both produce %q",
				previous,
				name,
				slug,
			)
		}
		seen[slug] = name

		inputPath := filepath.Join(cfg.ContentDir, name)

		page, err := parseEntry(inputPath)
		if err != nil {
			return fmt.Errorf("process %q: %w", inputPath, err)
		}

		if err := buildPage(cfg, pageTemplate, page, slug); err != nil {
			return err
		}

		indexEntry := IndexEntry{
			Title: page.Title,
			URL: filepath.ToSlash(
				filepath.Join(pagesDirName, slug+".html"),
			),
		}

		indexEntries = append(indexEntries, indexEntry)

		for _, tag := range page.Tags {
			tagSlug := buildSlug(tag)
			tags[tagSlug] = append(tags[tagSlug], indexEntry)
		}
	}

	if err := buildIndex(cfg, indexTemplate, indexEntries); err != nil {
		return err
	}

	if err := buildTags(cfg, tagsTemplate, tags); err != nil {
		return err
	}

	return nil
}

func prepareOutput(cfg config.Config) error {
	for _, dir := range []string{
		cfg.OutputDir,
		filepath.Join(cfg.OutputDir, pagesDirName),
		filepath.Join(cfg.OutputDir, tagsDirName),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create output dir %q: %w", dir, err)
		}
	}

	return nil
}

func loadTemplate(
	cfg config.Config,
	name string,
) (*template.Template, error) {
	path := filepath.Join(cfg.TemplateDir, name)

	tmpl, err := template.New(filepath.Base(name)).
		Funcs(template.FuncMap{
			"markdown": render.MarkdownToHTML,
			"tagSlug":  buildSlug,
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

func buildSlug(value string) string {
	value = strings.ToLower(value)

	var b strings.Builder
	isLastDash := false

	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
			isLastDash = false
			continue
		}

		if b.Len() > 0 && !isLastDash {
			b.WriteRune('-')
			isLastDash = true
		}
	}

	return strings.TrimRight(b.String(), "-")
}
