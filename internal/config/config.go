package config

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Title         string `yaml:"title"`
	ContentDir    string `yaml:"content_dir"`
	TemplateDir   string `yaml:"template_dir"`
	PageTemplate  string `yaml:"page_template"`
	IndexTemplate string `yaml:"index_template"`
	TagsTemplate  string `yaml:"tags_template"`
	OutputDir     string `yaml:"output_dir"`
	ServerBaseURL string `yaml:"base_url"`
}

func NewFromYAML(reader io.Reader) (Config, error) {
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true)

	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, err
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) validate() error {
	output, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		return err
	}

	content, err := filepath.Abs(cfg.ContentDir)
	if err != nil {
		return err
	}

	templates, err := filepath.Abs(cfg.TemplateDir)
	if err != nil {
		return err
	}

	if pathsOverlap(output, content) {
		return fmt.Errorf(
			"output directory %q overlaps content directory %q",
			output,
			content,
		)
	}

	if pathsOverlap(output, templates) {
		return fmt.Errorf(
			"output directory %q overlaps template directory %q",
			output,
			templates,
		)
	}

	return nil
}

func pathsOverlap(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)

	rel, err := filepath.Rel(a, b)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return true
	}

	rel, err = filepath.Rel(b, a)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return true
	}

	return false
}
