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
	outputDir, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("resolve output dir: %w", err)
	}
	outputDir = filepath.Clean(outputDir)

	parent := filepath.Dir(outputDir)
	base := filepath.Base(outputDir)

	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("create output parent dir: %w", err)
	}

	tmpDir, err := os.MkdirTemp(parent, "."+base+"-build-*")
	if err != nil {
		return fmt.Errorf("create temporary build dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	buildCfg := cfg
	buildCfg.OutputDir = tmpDir

	if err := buildSite(buildCfg); err != nil {
		return err
	}

	if err := os.Chmod(tmpDir, 0755); err != nil {
		return fmt.Errorf("set output dir permissions: %w", err)
	}

	if err := replaceOutput(tmpDir, outputDir); err != nil {
		return err
	}

	return nil
}

func buildSite(cfg config.Config) error {
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
	seen := make(map[string]string)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		slug := buildSlug(strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))

		if previous, ok := seen[slug]; ok {
			return fmt.Errorf(
				"page slug collision: %q and %q both produce %q",
				previous,
				entry.Name(),
				slug,
			)
		}

		seen[slug] = entry.Name()

		page, outputName, err := buildPage(cfg, pageTemplate, entry.Name(), slug)
		if err != nil {
			return err
		}

		indexEntry := IndexEntry{
			Title: page.Title,
			URL:   filepath.ToSlash(filepath.Join(pagesDirName, outputName)),
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

func replaceOutput(tmpDir, outputDir string) error {
	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("remove old output: %w", err)
	}

	if err := os.Rename(tmpDir, outputDir); err != nil {
		return fmt.Errorf("replace output: %w", err)
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
