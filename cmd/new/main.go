package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func printHelp() {
	fmt.Print(`
╔══════════════════════════════════════════════════════════════╗
║              gokit Scaffolding CLI                           ║
╚══════════════════════════════════════════════════════════════╝

USAGE:
  gokit new <github.com/username/project-name>

DESCRIPTION:
  Scaffolds a new project based on the golang-starter-kit in a sibling directory.
  It automatically copies the core framework, removes example modules (Product, Order),
  and renames the go module.

EXAMPLES:
  task gokit-new name=github.com/myuser/my-awesome-api
  go run cmd/new/main.go github.com/myuser/my-awesome-api
`)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	newModule := os.Args[1]
	if newModule == "--help" || newModule == "-h" {
		printHelp()
		os.Exit(0)
	}
	
	if strings.HasPrefix(newModule, "name=") {
		newModule = strings.TrimPrefix(newModule, "name=")
	}

	parts := strings.Split(newModule, "/")
	projectName := parts[len(parts)-1]

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	targetDir := filepath.Join(filepath.Dir(cwd), projectName)

	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		fmt.Printf("Error: Target directory '%s' already exists.\n", targetDir)
		os.Exit(1)
	}

	fmt.Printf("🚀 Scaffolding new project: %s\n", newModule)
	fmt.Printf("📂 Target directory: %s\n", targetDir)

	// 1. Copy files
	err = copyDir(cwd, targetDir)
	if err != nil {
		fmt.Printf("❌ Failed to copy files: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Core framework copied.")

	// 2. Remove examples using the copied rm command
	fmt.Println("🧹 Removing example modules (Product, Order)...")
	runCmd(targetDir, "go", "run", "cmd/rm/main.go", "--name", "Product")
	runCmd(targetDir, "go", "run", "cmd/rm/main.go", "--name", "Order")

	// 3. Rename module
	fmt.Println("🔄 Renaming Go module...")
	oldModule := "github.com/mkhsnw/golang-starter-kit"
	err = replaceInFiles(targetDir, oldModule, newModule)
	if err != nil {
		fmt.Printf("❌ Failed to rename module: %v\n", err)
		os.Exit(1)
	}

	// 4. Run go mod tidy
	fmt.Println("📦 Running go mod tidy...")
	runCmd(targetDir, "go", "mod", "tidy")

	fmt.Println("\n✅ Project scaffolded successfully!")
	fmt.Printf("Next steps:\n")
	fmt.Printf("  cd ../%s\n", projectName)
	fmt.Printf("  cp env.example.json env.json\n")
	fmt.Printf("  task install-tools\n")
}

func copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip .git, .github, env.json, and other unnecessary things for a fresh project
		if d.IsDir() && (d.Name() == ".git" || d.Name() == ".github" || d.Name() == "tmp" || d.Name() == ".idea" || d.Name() == ".vscode" || d.Name() == "bin") {
			return filepath.SkipDir
		}
		if !d.IsDir() && (d.Name() == "env.json" || d.Name() == "coverage.out" || strings.HasSuffix(d.Name(), ".exe")) {
			return nil
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		return copyFile(path, targetPath)
	})
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return nil
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func replaceInFiles(dir string, oldStr, newStr string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".go" || ext == ".mod" || ext == ".yml" || ext == ".yaml" || ext == ".md" {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			strContent := string(content)
			if strings.Contains(strContent, oldStr) {
				newContent := strings.ReplaceAll(strContent, oldStr, newStr)
				err = os.WriteFile(path, []byte(newContent), 0644)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func runCmd(dir string, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}
