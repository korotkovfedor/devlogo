package content

import (
	"errors"
	"io"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Header struct {
	Title  string    `yaml:"title"`
	Date   time.Time `yaml:"date"`
	Type   string    `yaml:"type"`
	Tags   []string  `yaml:"tags"`
	Status string    `yaml:"status"`
}

func newHeaderFromYAML(reader io.Reader) (Header, error) {
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true)

	var header Header
	if err := decoder.Decode(&header); err != nil {
		return Header{}, err
	}

	if err := header.validate(); err != nil {
		return Header{}, err
	}

	return header, nil
}

func (h Header) validate() error {
	if strings.TrimSpace(h.Title) == "" {
		return errors.New("empty title")
	}

	if h.Date.IsZero() {
		return errors.New("empty date")
	}

	if strings.TrimSpace(h.Type) == "" {
		return errors.New("empty type")
	}

	if strings.TrimSpace(h.Status) == "" {
		return errors.New("empty status")
	}

	return nil

}
