package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Manifest defines the structure for a module generation manifest file (e.g. product.yaml)
type Manifest struct {
	Name         string          `yaml:"name"`
	Type         string          `yaml:"type"`
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

type rawManifest struct {
	Name          string `yaml:"name"`
	Module        string `yaml:"module"`
	Type          string `yaml:"type"`
	Table         string `yaml:"table"`
	Transactions  *bool  `yaml:"transactions"`
	Transactional *bool  `yaml:"transactional"`
	Tests         *bool  `yaml:"tests"`
	GenerateTest  *bool  `yaml:"generate_test"`
	Options       struct {
		Transactional *bool `yaml:"transactional"`
		GenerateTest  *bool `yaml:"generate_test"`
		Transactions  *bool `yaml:"transactions"`
		Tests         *bool `yaml:"tests"`
	} `yaml:"options"`
	Fields []ManifestField `yaml:"fields"`
}

// LoadManifest reads and parses a YAML manifest file with backward and flexible key compatibility
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw rawManifest
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	m := &Manifest{
		Name:   raw.Name,
		Type:   raw.Type,
		Table:  raw.Table,
		Fields: raw.Fields,
	}

	if m.Name == "" {
		m.Name = raw.Module
	}

	// Handle transactions / transactional flag
	if raw.Transactions != nil {
		m.Transactions = *raw.Transactions
	} else if raw.Transactional != nil {
		m.Transactions = *raw.Transactional
	} else if raw.Options.Transactional != nil {
		m.Transactions = *raw.Options.Transactional
	} else if raw.Options.Transactions != nil {
		m.Transactions = *raw.Options.Transactions
	} else if raw.Type == "transactional" {
		m.Transactions = true
	}

	// Handle tests flag
	if raw.Tests != nil {
		m.Tests = *raw.Tests
	} else if raw.GenerateTest != nil {
		m.Tests = *raw.GenerateTest
	} else if raw.Options.GenerateTest != nil {
		m.Tests = *raw.Options.GenerateTest
	} else if raw.Options.Tests != nil {
		m.Tests = *raw.Options.Tests
	}

	return m, nil
}
