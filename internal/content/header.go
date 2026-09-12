package content

import (
	"io"
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

	return header, nil
}
