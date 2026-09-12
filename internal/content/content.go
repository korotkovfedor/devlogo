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
	Body
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

	body := string(bodyRaw)

	return Content{Header: header, Body: body}, nil

}

func splitFrontMatter(r io.Reader) ([]byte, []byte, error) {
	reader := bufio.NewReader(r)

	firstLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, nil, fmt.Errorf("read first line: %w", err)
	}

	if strings.TrimSpace(firstLine) != "---" {
		return nil, nil, errors.New("front matter not found")
	}

	var header bytes.Buffer

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("read front matter: %w", err)
		}

		if strings.TrimSpace(line) == "---" {
			break
		}

		header.WriteString(line)
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("read body: %w", err)
	}

	return header.Bytes(), body, nil
}
