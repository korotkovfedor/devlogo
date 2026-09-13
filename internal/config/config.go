package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
	ServerPort    int    `yaml:"server_port"`
}

func NewFromYAML(reader io.Reader) (Config, error) {
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true)

	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, err
	}

	if cfg.ServerPort <= 0 || cfg.ServerPort > 65535 {
		return Config{}, errors.New("invalid server port")
	}

	return cfg, nil
}

func ValidatePaths(cfg Config) error {
	for _, path := range []string{
		cfg.ContentDir,
		cfg.TemplateDir,
		cfg.OutputDir,
	} {
		if err := rejectSymlinks(path); err != nil {
			return err
		}
	}

	overlap, err := pathsOverlap(cfg.OutputDir, cfg.ContentDir)
	if err != nil {
		return err
	}

	if overlap {
		return fmt.Errorf(
			"output directory %q overlaps content directory %q",
			cfg.OutputDir,
			cfg.ContentDir,
		)
	}

	overlap, err = pathsOverlap(cfg.OutputDir, cfg.TemplateDir)
	if err != nil {
		return err
	}

	if overlap {
		return fmt.Errorf(
			"output directory %q overlaps template directory %q",
			cfg.OutputDir,
			cfg.TemplateDir,
		)
	}

	return nil
}

func rejectSymlinks(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	for current := path; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)

		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlinks are not allowed in path %q", path)
		}

		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect %q: %w", current, err)
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}

	return nil
}

func pathsOverlap(output, source string) (bool, error) {
	inside, err := isInside(output, source)
	if err != nil {
		return false, err
	}
	if inside {
		return true, nil
	}

	// there may be no output upon the first launch
	if _, err := os.Stat(output); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	return isInside(source, output)
}

func isInside(child, parent string) (bool, error) {
	parentInfo, err := os.Stat(parent)
	if err != nil {
		return false, fmt.Errorf("stat %q: %w", parent, err)
	}

	child, err = filepath.Abs(child)
	if err != nil {
		return false, err
	}

	for {
		info, err := os.Stat(child)

		switch {
		case err == nil:
			if os.SameFile(info, parentInfo) {
				return true, nil
			}

		case !errors.Is(err, os.ErrNotExist):
			return false, fmt.Errorf("stat %q: %w", child, err)
		}

		next := filepath.Dir(child)
		if next == child {
			break
		}

		child = next
	}

	return false, nil
}
