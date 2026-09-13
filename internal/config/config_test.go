package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewFromYAML(t *testing.T) {
	const validFields = `title: "Test log"
content_dir: "./entries"
template_dir: "./templates"
page_template: "page.html"
index_template: "index.html"
tags_template: "tags.html"
output_dir: "./dist"
`
	maxPortConfig := Config{
		Title:         "Test log",
		ContentDir:    "./entries",
		TemplateDir:   "./templates",
		PageTemplate:  "page.html",
		IndexTemplate: "index.html",
		TagsTemplate:  "tags.html",
		OutputDir:     "./dist",
		ServerPort:    65535,
	}
	minPortConfig := maxPortConfig
	minPortConfig.ServerPort = 1

	tests := []struct {
		name    string
		reader  io.Reader
		want    Config
		wantErr string
	}{
		{"valid config with maximum port", strings.NewReader(validFields + "server_port: 65535"), maxPortConfig, ""},
		{"minimum port", strings.NewReader(validFields + "server_port: 1"), minPortConfig, ""},
		{"port less than 1", strings.NewReader(validFields + "server_port: 0"), Config{}, "invalid server port"},
		{"port higher than 65535", strings.NewReader(validFields + "server_port: 65536"), Config{}, "invalid server port"},
		{"invalid yaml", strings.NewReader("title: ["), Config{}, "did not find expected node content"},
		{"empty yaml", strings.NewReader(""), Config{}, io.EOF.Error()},
		{"unknown field", strings.NewReader(validFields + "server_port: 1\nnew_field: 42"), Config{}, "field new_field not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFromYAML(tt.reader)
			switch tt.wantErr {
			case "":
				if err != nil {
					t.Fatalf("NewFromYAML() unexpected error: %v", err)
				}
			case io.EOF.Error():
				if !errors.Is(err, io.EOF) {
					t.Fatalf("NewFromYAML() error = %v; want EOF", err)
				}
			default:
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("NewFromYAML() error = %v; want error containing %q", err, tt.wantErr)
				}
			}
			if got != tt.want {
				t.Errorf("NewFromYAML() got = %+v; want %+v", got, tt.want)
			}
		})
	}
}

func TestValidatePaths(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		templates string
		output    string
		wantErr   string
	}{
		{"separate directories", "entries", "templates", "dist", ""},
		{"missing output", "entries", "templates", "new/dist", ""},
		{"similar directory names", "entries", "templates", "entries-backup", ""},
		{"shared input directories", "entries", "entries", "dist", ""},
		{"nested input directories", "entries", "entries/nested", "dist", ""},
		{"output equals content", "entries", "templates", "entries", "overlaps content directory"},
		{"output equals templates", "entries", "templates", "templates", "overlaps template directory"},
		{"output inside content", "entries", "templates", "entries/nested", "overlaps content directory"},
		{"output inside templates", "entries", "templates", "templates/nested", "overlaps template directory"},
		{"missing output inside content", "entries", "templates", "entries/new/dist", "overlaps content directory"},
		{"missing output inside templates", "entries", "templates", "templates/new/dist", "overlaps template directory"},
		{"content inside output", "dist/entries", "templates", "dist", "overlaps content directory"},
		{"templates inside output", "entries", "dist/templates", "dist", "overlaps template directory"},
		{"missing content", "missing-entries", "templates", "dist", "no such file"},
		{"missing templates", "entries", "missing-templates", "dist", "no such file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := pathTestRoot(t)
			for _, dir := range []string{
				"entries/nested", "templates/nested", "dist/entries", "dist/templates", "entries-backup",
			} {
				if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
					t.Fatal(err)
				}
			}

			cfg := Config{
				ContentDir:  filepath.Join(root, tt.content),
				TemplateDir: filepath.Join(root, tt.templates),
				OutputDir:   filepath.Join(root, tt.output),
			}
			err := ValidatePaths(cfg)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidatePaths() unexpected error: %v", err)
				}
				return
			}
			if tt.wantErr == "no such file" {
				if !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("ValidatePaths() error = %v; want a missing path error", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ValidatePaths() error = %v; want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePathsSymlinks(t *testing.T) {
	for _, field := range []string{"content", "templates", "output"} {
		for _, ancestor := range []bool{false, true} {
			name := field + "/direct"
			if ancestor {
				name = field + "/ancestor"
			}
			t.Run(name, func(t *testing.T) {
				root := pathTestRoot(t)
				for _, dir := range []string{"entries", "templates", "target/child"} {
					if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
						t.Fatal(err)
					}
				}
				link := filepath.Join(root, "link")
				if err := os.Symlink(filepath.Join(root, "target"), link); err != nil {
					if runtime.GOOS == "windows" && errors.Is(err, os.ErrPermission) {
						t.Skipf("creating symlinks is not permitted: %v", err)
					}
					t.Fatal(err)
				}
				if ancestor {
					link = filepath.Join(link, "child")
					if field == "output" {
						link = filepath.Join(link, "new-dist")
					}
				}

				cfg := Config{
					ContentDir:  filepath.Join(root, "entries"),
					TemplateDir: filepath.Join(root, "templates"),
					OutputDir:   filepath.Join(root, "dist"),
				}
				switch field {
				case "content":
					cfg.ContentDir = link
				case "templates":
					cfg.TemplateDir = link
				case "output":
					cfg.OutputDir = link
				}

				err := ValidatePaths(cfg)
				if err == nil || !strings.Contains(err.Error(), "symlinks are not allowed") {
					t.Fatalf("ValidatePaths() error = %v; want a symlink rejection", err)
				}
			})
		}
	}
}

func pathTestRoot(t *testing.T) string {
	t.Helper()
	// The system temporary directory may itself contain symlinks, e.g. /var on macOS.
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return root
}
