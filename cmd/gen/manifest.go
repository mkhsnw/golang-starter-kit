package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Manifest defines the structure for a module generation manifest file (e.g. product.yaml)
type Manifest struct {
	Name         string          `yaml:"name"`
	Table        string          `yaml:"table"`
	Transactions bool            `yaml:"transactions"`
	Tests        bool            `yaml:"tests"`
	Fields       []ManifestField `yaml:"fields"`
}

// ManifestField defines a single field within the module
type ManifestField struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	SqlType  string `yaml:"sql_type"`
	Required bool   `yaml:"required"`
	Primary  bool   `yaml:"primary"`
}

// LoadManifest reads and parses a YAML manifest file
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return &m, nil
}
