package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratorManifestLoading(t *testing.T) {
	tmpDir := t.TempDir()

	manifestContent := `name: Product
type: standard
transactions: true
fields:
  - name: name
    type: string
    required: true
  - name: price
    type: float64
    required: true
`
	manifestPath := filepath.Join(tmpDir, "product.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to write mock manifest: %v", err)
	}

	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load mock manifest: %v", err)
	}

	if manifest.Name != "Product" {
		t.Errorf("Expected manifest name 'Product', got '%s'", manifest.Name)
	}

	mod := getModuleNamesForName(manifest.Name)
	if mod.Pascal != "Product" || mod.Snake != "product" || mod.Plural != "products" {
		t.Errorf("Module name mapping error: %+v", mod)
	}
}

func TestGeneratorPluralize(t *testing.T) {
	tests := map[string]string{
		"category": "categories",
		"product":  "products",
		"user":     "users",
		"city":     "cities",
	}

	for input, expected := range tests {
		got := pluralize(input)
		if got != expected {
			t.Errorf("pluralize(%s) = %s; expected %s", input, got, expected)
		}
	}
}
