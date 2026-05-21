package parser

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

var separator = []byte("---")

// ParseFrontMatter splits YAML front matter from markdown body.
// Returns the metadata map, the remaining body bytes, and any error.
func ParseFrontMatter(data []byte) (map[string]any, []byte, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, nil, nil
	}

	if !bytes.HasPrefix(data, separator) {
		return nil, data, nil
	}

	rest := data[len(separator):]
	end := bytes.Index(rest, separator)
	if end == -1 {
		return nil, data, nil
	}

	frontMatter := rest[:end]
	body := bytes.TrimSpace(rest[end+len(separator):])

	meta := make(map[string]any)
	if err := yaml.Unmarshal(frontMatter, &meta); err != nil {
		return nil, nil, err
	}

	return meta, body, nil
}
