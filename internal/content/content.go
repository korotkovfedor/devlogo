package content

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Content struct {
	Header
	Markdown
}

func ParseMarkdown(reader io.Reader) (Content, error) {
	headerRaw, bodyRaw, err := splitFrontMatter(reader)
	if err != nil {
		return Content{}, err
	}

	header, err := newHeaderFromYAML(bytes.NewReader(headerRaw))
	if err != nil {
		return Content{}, err
	}

	body := Markdown(bodyRaw)

	return Content{Header: header, Markdown: body}, nil

}

func splitFrontMatter(r io.Reader) ([]byte, []byte, error) {
	reader := bufio.NewReader(r)

	firstLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, nil, fmt.Errorf("read first line: %w", err)
	}

	if !isDelimiter(firstLine) {
		return nil, nil, errors.New("front matter not found")
	}

	var header bytes.Buffer

	for {
		line, err := reader.ReadString('\n')
		if isDelimiter(line) {
			break
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, nil, errors.New("front matter is not closed")
			}

			return nil, nil, fmt.Errorf("read front matter: %w", err)
		}

		header.WriteString(line)
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("read body: %w", err)
	}

	return header.Bytes(), body, nil
}

func isDelimiter(line string) bool {
	return strings.TrimRight(line, "\r\n") == "---"
}
