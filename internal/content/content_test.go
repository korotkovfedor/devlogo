package content

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

func TestParseMarkdown(t *testing.T) {
	const validHeader = `title: "Test entry"
date: 2026-09-13
type: feature
status: done
`
	const body = "\n## Change\n\nSome **Markdown**.\n\n---\n\nLast line"
	header := Header{
		Title:  "Test entry",
		Date:   time.Date(2026, time.September, 13, 0, 0, 0, 0, time.UTC),
		Type:   "feature",
		Status: "done",
	}
	headerWithTags := header
	headerWithTags.Tags = []string{"go", "cli"}
	headerWithEmptyTags := header
	headerWithEmptyTags.Tags = []string{}
	document := "---\n" + validHeader + "---\n" + body

	type testCase struct {
		name    string
		input   string
		want    Content
		wantErr string
	}
	tests := []testCase{
		{"valid entry without tags", document, Content{Header: header, Markdown: Markdown(body)}, ""},
		{"valid entry with tags", "---\n" + validHeader + "tags: [go, cli]\n---\n" + body, Content{Header: headerWithTags, Markdown: Markdown(body)}, ""},
		{"empty tags", "---\n" + validHeader + "tags: []\n---\n" + body, Content{Header: headerWithEmptyTags, Markdown: Markdown(body)}, ""},
		{"CRLF line endings", strings.ReplaceAll(document, "\n", "\r\n"), Content{Header: header, Markdown: Markdown(strings.ReplaceAll(body, "\n", "\r\n"))}, ""},
		{"empty body", "---\n" + validHeader + "---\n", Content{Header: header}, ""},
		{"closing delimiter without final newline", "---\n" + validHeader + "---", Content{Header: header}, ""},
		{"empty input", "", Content{}, "read first line: EOF"},
		{"missing opening delimiter", "# Entry\n", Content{}, "front matter not found"},
		{"blank line before opening delimiter", "\n" + document, Content{}, "front matter not found"},
		{"spaces around opening delimiter", " --- \n" + validHeader + "---\n", Content{}, "front matter not found"},
		{"missing closing delimiter", "---\n" + validHeader, Content{}, "front matter is not closed"},
		{"spaces around closing delimiter", "---\n" + validHeader + " --- \n", Content{}, "front matter is not closed"},
		{"empty header", "---\n---\n", Content{}, "EOF"},
		{"invalid YAML", "---\ntitle: [\n---\n", Content{}, "did not find expected node content"},
		{"unknown field", "---\n" + validHeader + "unknown: true\n---\n", Content{}, "field unknown not found"},
		{"invalid date", strings.Replace(document, "2026-09-13", "not-a-date", 1), Content{}, "cannot parse"},
		{"missing date", strings.Replace(document, "date: 2026-09-13\n", "", 1), Content{}, "empty date"},
		{"null date", strings.Replace(document, "date: 2026-09-13", "date: null", 1), Content{}, "empty date"},
		{"invalid tags type", "---\n" + validHeader + "tags: go\n---\n", Content{}, "cannot unmarshal"},
	}

	for _, field := range []struct {
		name string
		line string
	}{
		{"title", "title: \"Test entry\"\n"},
		{"type", "type: feature\n"},
		{"status", "status: done\n"},
	} {
		for _, value := range []struct {
			name        string
			replacement string
		}{
			{"missing", ""},
			{"empty", field.name + ": \"\"\n"},
			{"whitespace", field.name + ": \"   \"\n"},
		} {
			tests = append(tests, testCase{
				name:    value.name + " " + field.name,
				input:   strings.Replace(document, field.line, value.replacement, 1),
				wantErr: "empty " + field.name,
			})
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMarkdown(strings.NewReader(tt.input))
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ParseMarkdown() unexpected error: %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ParseMarkdown() error = %v; want error containing %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMarkdown() got = %#v; want %#v", got, tt.want)
			}
		})
	}
}

func TestParseMarkdownReadErrors(t *testing.T) {
	readErr := errors.New("test read failure")
	tests := []struct {
		name   string
		prefix string
		stage  string
	}{
		{"first line", "", "read first line"},
		{"header", "---\ntitle: Test\n", "read front matter"},
		{"body", "---\ntitle: Test\n---\nSome body", "read body"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := io.MultiReader(strings.NewReader(tt.prefix), iotest.ErrReader(readErr))
			got, err := ParseMarkdown(reader)
			if !errors.Is(err, readErr) {
				t.Fatalf("ParseMarkdown() error = %v; want wrapped %v", err, readErr)
			}
			if !strings.Contains(err.Error(), tt.stage) {
				t.Errorf("ParseMarkdown() error = %v; want stage %q", err, tt.stage)
			}
			if !reflect.DeepEqual(got, Content{}) {
				t.Errorf("ParseMarkdown() got = %#v; want empty Content on read failure", got)
			}
		})
	}
}
