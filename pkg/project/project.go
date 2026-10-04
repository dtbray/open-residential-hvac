// SPDX-License-Identifier: AGPL-3.0-only
// Package project owns versioned strict serialization, independently of the engine.
package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/building"
	"go.yaml.in/yaml/v3"
)

const Version = 1

type Project struct {
	Version  int               `json:"version" yaml:"version"`
	Building building.Building `json:"building" yaml:"building"`
}

func Decode(data []byte, format string) (*Project, error) {
	var p Project
	switch strings.ToLower(format) {
	case "json":
		if err := rejectDuplicateKeys(data); err != nil {
			return nil, fmt.Errorf("project JSON: %w", err)
		}
		d := json.NewDecoder(bytes.NewReader(data))
		d.DisallowUnknownFields()
		if err := d.Decode(&p); err != nil {
			return nil, fmt.Errorf("project JSON: %w", err)
		}
		var extra any
		if err := d.Decode(&extra); err != io.EOF {
			return nil, fmt.Errorf("project must contain exactly one document")
		}
	case "yaml", "yml":
		d := yaml.NewDecoder(bytes.NewReader(data))
		d.KnownFields(true)
		if err := d.Decode(&p); err != nil {
			return nil, fmt.Errorf("project YAML: %w", err)
		}
		var extra any
		if err := d.Decode(&extra); err != io.EOF {
			return nil, fmt.Errorf("project must contain exactly one document")
		}
	default:
		return nil, fmt.Errorf("unsupported project format %q", format)
	}
	if p.Version != Version {
		return nil, fmt.Errorf("unsupported project version %d; supported: %d", p.Version, Version)
	}
	return &p, nil
}

// Ambiguous duplicate input keys must not silently overwrite engineering values.
func rejectDuplicateKeys(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	var value func() error
	value = func() error {
		t, err := d.Token()
		if err != nil {
			return err
		}
		delim, ok := t.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for d.More() {
				key, err := d.Token()
				if err != nil {
					return err
				}
				k, ok := key.(string)
				if !ok {
					return fmt.Errorf("object key must be a string")
				}
				if seen[k] {
					return fmt.Errorf("duplicate key %q", k)
				}
				seen[k] = true
				if err = value(); err != nil {
					return err
				}
			}
		case '[':
			for d.More() {
				if err = value(); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unexpected delimiter %q", delim)
		}
		_, err = d.Token()
		return err
	}
	return value()
}

func Encode(p Project, format string) ([]byte, error) {
	if p.Version != Version {
		return nil, fmt.Errorf("unsupported project version %d", p.Version)
	}
	switch strings.ToLower(format) {
	case "json":
		return json.MarshalIndent(p, "", "  ")
	case "yaml", "yml":
		return yaml.Marshal(p)
	default:
		return nil, fmt.Errorf("unsupported project format %q", format)
	}
}

func Open(path string) (*Project, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Decode(b, strings.TrimPrefix(filepath.Ext(path), "."))
}
