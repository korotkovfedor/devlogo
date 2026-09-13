package builder

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/korotkovfedor/devlogo/internal/config"
)

func TestBuild(t *testing.T) {
	cfg := newBuildTestConfig(t)
	writeBuildTestFile(t, filepath.Join(cfg.ContentDir, "01 First.md"),
		buildTestEntry("First & entry", "[\"Go Lang\", backend]", "# First\n"))
	writeBuildTestFile(t, filepath.Join(cfg.ContentDir, "02 Second.md"),
		buildTestEntry("Second", "[go-lang]", "**Second**\n"))
	// These files must be skipped, even though their contents are invalid Markdown entries.
	writeBuildTestFile(t, filepath.Join(cfg.ContentDir, "ignored.txt"), "invalid entry")
	writeBuildTestFile(t, filepath.Join(cfg.ContentDir, "nested.md", "ignored.md"), "invalid entry")

	if err := Build(cfg); err != nil {
		t.Fatalf("Build() unexpected error: %v", err)
	}

	assertBuildTestFiles(t, cfg.OutputDir, map[string]string{
		"index.html": "<h1>Test log</h1>" +
			`<a href="pages/01-first.html">First &amp; entry</a>` +
			`<a href="pages/02-second.html">Second</a>`,
		"pages/01-first.html": "<h1>First &amp; entry</h1><h1>First</h1>\n" +
			`<a href="../tags/go-lang.html">Go Lang</a>` +
			`<a href="../tags/backend.html">backend</a>`,
		"pages/02-second.html": "<h1>Second</h1><p><strong>Second</strong></p>\n" +
			`<a href="../tags/go-lang.html">go-lang</a>`,
		"tags/go-lang.html": "<h1>go-lang</h1>" +
			`<a href="../pages/01-first.html">First &amp; entry</a>` +
			`<a href="../pages/02-second.html">Second</a>`,
		"tags/backend.html": "<h1>backend</h1>" +
			`<a href="../pages/01-first.html">First &amp; entry</a>`,
	})
}

func TestBuildEmptyContent(t *testing.T) {
	cfg := newBuildTestConfig(t)
	if err := Build(cfg); err != nil {
		t.Fatalf("Build() unexpected error: %v", err)
	}
	assertBuildTestFiles(t, cfg.OutputDir, map[string]string{
		"index.html": "<h1>Test log</h1>",
	})
	for _, dir := range []string{"pages", "tags"} {
		info, err := os.Stat(filepath.Join(cfg.OutputDir, dir))
		if err != nil {
			t.Fatal(err)
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

func TestBuildWithoutTags(t *testing.T) {
	cfg := newBuildTestConfig(t)
	writeBuildTestFile(t, filepath.Join(cfg.ContentDir, "entry.md"),
		buildTestEntry("Entry", "", "Text\n"))
	if err := Build(cfg); err != nil {
		t.Fatalf("Build() unexpected error: %v", err)
	}
	assertBuildTestFiles(t, cfg.OutputDir, map[string]string{
		"index.html":       `<h1>Test log</h1><a href="pages/entry.html">Entry</a>`,
		"pages/entry.html": "<h1>Entry</h1><p>Text</p>\n",
	})
}

func TestBuildReplacesPreviousOutput(t *testing.T) {
	cfg := newBuildTestConfig(t)
	entryPath := filepath.Join(cfg.ContentDir, "entry.md")
	writeBuildTestFile(t, entryPath, buildTestEntry("Old", "[old]", "Old body\n"))
	if err := Build(cfg); err != nil {
		t.Fatalf("first Build() failed: %v", err)
	}
	writeBuildTestFile(t, filepath.Join(cfg.OutputDir, "stale.html"), "stale page")
	writeBuildTestFile(t, entryPath, buildTestEntry("Updated", "", "New body\n"))

	if err := Build(cfg); err != nil {
		t.Fatalf("second Build() failed: %v", err)
	}
	assertBuildTestFiles(t, cfg.OutputDir, map[string]string{
		"index.html":       `<h1>Test log</h1><a href="pages/entry.html">Updated</a>`,
		"pages/entry.html": "<h1>Updated</h1><p>New body</p>\n",
	})

	if err := os.Remove(entryPath); err != nil {
		t.Fatal(err)
	}
	if err := Build(cfg); err != nil {
		t.Fatalf("Build() after deleting entry failed: %v", err)
	}
	assertBuildTestFiles(t, cfg.OutputDir, map[string]string{
		"index.html": "<h1>Test log</h1>",
	})
}

func TestBuildInputErrors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T, *config.Config)
		wantErr []string
	}{
		{
			name: "overlapping directories",
			setup: func(t *testing.T, cfg *config.Config) {
				cfg.OutputDir = cfg.ContentDir
			},
			wantErr: []string{"overlaps content directory"},
		},
		{
			name: "missing content directory",
			setup: func(t *testing.T, cfg *config.Config) {
				cfg.ContentDir = filepath.Join(cfg.ContentDir, "missing")
			},
			wantErr: []string{"stat", "missing"},
		},
		{
			name: "content path is a file",
			setup: func(t *testing.T, cfg *config.Config) {
				cfg.ContentDir = filepath.Join(cfg.ContentDir, "file")
				writeBuildTestFile(t, cfg.ContentDir, "not a directory")
			},
			wantErr: []string{"read content dir"},
		},
		{
			name: "invalid entry",
			setup: func(t *testing.T, cfg *config.Config) {
				writeBuildTestFile(t, filepath.Join(cfg.ContentDir, "broken.md"), "# No front matter\n")
			},
			wantErr: []string{"process", "broken.md", "front matter not found"},
		},
		{
			name: "page slug collision",
			setup: func(t *testing.T, cfg *config.Config) {
				for _, name := range []string{"Go Lang.md", "go-lang.md"} {
					writeBuildTestFile(t, filepath.Join(cfg.ContentDir, name), buildTestEntry("Entry", "", "Text\n"))
				}
			},
			wantErr: []string{"page slug collision", "Go Lang.md", "go-lang.md", `both produce "go-lang"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newBuildTestConfig(t)
			tt.setup(t, &cfg)
			assertBuildTestError(t, Build(cfg), tt.wantErr...)
		})
	}
}

func TestBuildTemplateErrors(t *testing.T) {
	for _, template := range []struct {
		file        string
		renderError string
	}{
		{"page.html", `render page "entry"`},
		{"index.html", "render index"},
		{"tags.html", `render tag "go"`},
	} {
		for _, failure := range []string{"missing", "invalid syntax", "execution error"} {
			t.Run(template.file+"/"+failure, func(t *testing.T) {
				cfg := newBuildTestConfig(t)
				writeBuildTestFile(t, filepath.Join(cfg.ContentDir, "entry.md"), buildTestEntry("Entry", "[go]", "Text\n"))
				path := filepath.Join(cfg.TemplateDir, template.file)
				wantErr := []string{fmt.Sprintf("parse template %q", template.file)}
				switch failure {
				case "missing":
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
				case "invalid syntax":
					writeBuildTestFile(t, path, "{{")
				case "execution error":
					writeBuildTestFile(t, path, "{{.MissingField}}")
					wantErr = []string{template.renderError, "execute template", "MissingField"}
				}
				assertBuildTestError(t, Build(cfg), wantErr...)
			})
		}
	}
}

func newBuildTestConfig(t *testing.T) config.Config {
	t.Helper()
	// Resolve system symlinks (such as /var on macOS) before validating paths.
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Title:         "Test log",
		ContentDir:    filepath.Join(root, "entries"),
		TemplateDir:   filepath.Join(root, "templates"),
		PageTemplate:  "page.html",
		IndexTemplate: "index.html",
		TagsTemplate:  "tags.html",
		OutputDir:     filepath.Join(root, "output", "dist"),
		ServerPort:    8080,
	}
	if err := os.MkdirAll(cfg.ContentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, source := range map[string]string{
		"page.html":  `<h1>{{.Header.Title}}</h1>{{.Markdown | markdown}}{{range .Header.Tags}}<a href="../tags/{{. | tagSlug}}.html">{{.}}</a>{{end}}`,
		"index.html": `<h1>{{.Title}}</h1>{{range .Entries}}<a href="{{.URL}}">{{.Title}}</a>{{end}}`,
		"tags.html":  `<h1>{{.Tag}}</h1>{{range .Entries}}<a href="../{{.URL}}">{{.Title}}</a>{{end}}`,
	} {
		writeBuildTestFile(t, filepath.Join(cfg.TemplateDir, name), source)
	}
	return cfg
}

func buildTestEntry(title, tags, body string) string {
	header := fmt.Sprintf("---\ntitle: %q\ndate: 2026-09-13\ntype: feature\nstatus: done\n", title)
	if tags != "" {
		header += "tags: " + tags + "\n"
	}
	return header + "---\n" + body
}

func writeBuildTestFile(t *testing.T, path, value string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertBuildTestFiles(t *testing.T, root string, want map[string]string) {
	t.Helper()
	got := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		got[filepath.ToSlash(relative)] = string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("generated files = %#v; want %#v", got, want)
	}
}

func assertBuildTestError(t *testing.T, err error, fragments ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Build() succeeded; want error containing %q", fragments)
	}
	for _, fragment := range fragments {
		if !strings.Contains(err.Error(), fragment) {
			t.Errorf("Build() error = %v; want error containing %q", err, fragment)
		}
	}
}

func TestBuildSlug(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty input", "", ""},
		{"single letter", "A", "a"},
		{"already normalized", "go-cli", "go-cli"},
		{"mixed case", "My First POST", "my-first-post"},
		{"digits", "Go 123", "go-123"},
		{"digits only", "2026", "2026"},
		{"cyrillic", "ПЕРВАЯ Запись Ёж", "первая-запись-ёж"},
		{"accented letters", "CAFÉ Déjà Vu", "café-déjà-vu"},
		{"non-Latin letters", "学习 Go", "学习-go"},
		{"Unicode numbers", "Версия １２３", "версия-１２３"},
		{"repeated spaces", "go   cli", "go-cli"},
		{"repeated dashes", "go---cli", "go-cli"},
		{"mixed separators", "go _./!? cli", "go-cli"},
		{"leading separators", " --Go", "go"},
		{"trailing separators", "Go-- ", "go"},
		{"separators on both ends", "-- Go CLI --", "go-cli"},
		{"multiple words", "one  two___three", "one-two-three"},
		{"tabs and line endings", "go\tcli\r\ntest", "go-cli-test"},
		{"Unicode whitespace", "go\u00a0cli", "go-cli"},
		{"emoji separator", "go🚀cli", "go-cli"},
		{"whitespace only", " \t\r\n", ""},
		{"punctuation only", "---_./!?", ""},
		{"emoji only", "🚀🔥", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildSlug(tt.input)
			if got != tt.want {
				t.Errorf("buildSlug(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}
