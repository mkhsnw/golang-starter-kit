package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func printHelp() {
	fmt.Print(`
╔══════════════════════════════════════════════════════════════╗
║              🚂 RelGo Framework CLI (rel)                     ║
╚══════════════════════════════════════════════════════════════╝

RelGo — Framework Go opini yang bikin kamu ngoding REST API ngebut, rapi, dan anti-ribet, mulus kayak kereta berjalan di atas rel.

USAGE:
  rel <command> [arguments]

SCAFFOLDING & GENERATOR:
  rel new <project-path>     Scaffold a new RelGo project
  rel gen [module-name]      Generate a CRUD or business module (or interactive wizard)
  rel rm <module-name>       Remove a generated module cleanly
  rel make-seeder <name>     Generate a database seeder file
  rel make-factory <name>    Generate a data factory file
  rel make-migration <name>  Generate new SQL migration files (.up.sql & .down.sql)

DATABASE & MIGRATIONS:
  rel migrate [action]       Run migrations (up, down, fresh, version, force)
  rel seed [count=N]         Populate database with seed data

QUALITY & DIAGNOSTICS:
  rel doctor                 Run dev environment diagnostic health check
  rel lint                   Run Architecture Linter (AST parser rule check)
  rel version                Show RelGo version information
  rel help                   Show this help documentation

EXAMPLES:
  rel new github.com/username/my-awesome-api
  rel gen Product
  rel migrate up
  rel seed count=50
  rel doctor
  rel lint
`)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "new":
		if _, err := os.Stat("cmd/new/main.go"); err == nil {
			runSubcommand("cmd/new/main.go", args...)
		} else {
			scaffoldNewProject(args)
		}
	case "gen":
		runSubcommand("cmd/gen/main.go", args...)
	case "rm":
		runSubcommand("cmd/rm/main.go", args...)
	case "migrate":
		runSubcommand("cmd/migrate/main.go", args...)
	case "seed":
		runSubcommand("cmd/seed/main.go", args...)
	case "make-seeder":
		runSubcommand("cmd/gen/main.go", append([]string{"--make-seeder"}, args...)...)
	case "make-factory":
		runSubcommand("cmd/gen/main.go", append([]string{"--make-factory"}, args...)...)
	case "make-migration", "migrate-create":
		runSubcommand("cmd/gen/main.go", append([]string{"--make-migration"}, args...)...)
	case "doctor":
		runSubcommand("cmd/doctor/main.go", args...)
	case "lint", "lint-arch":
		runSubcommand("cmd/lint/main.go", args...)
	case "version", "-v", "--version":
		fmt.Println("🚂 RelGo Framework v1.0.2 (Core Frozen & Production Ready)")
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("❌ Unknown command 'rel %s'. Run 'rel help' for usage.\n", cmd)
		os.Exit(1)
	}
}

func scaffoldNewProject(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: rel new <github.com/username/project-name>")
		os.Exit(1)
	}

	newModule := strings.TrimPrefix(args[0], "name=")
	parts := strings.Split(newModule, "/")
	projectName := parts[len(parts)-1]

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	targetDir := filepath.Join(cwd, projectName)

	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		fmt.Printf("❌ Error: Target directory '%s' already exists.\n", targetDir)
		os.Exit(1)
	}

	fmt.Printf("🚀 Scaffolding new RelGo project: %s\n", newModule)
	fmt.Printf("📂 Target directory: %s\n\n", targetDir)

	// Step 1: Clone RelGo framework template from GitHub
	fmt.Println("1/4 📦 Downloading RelGo core framework template...")
	cloneCmd := exec.Command("git", "clone", "--depth", "1", "https://github.com/mkhsnw/rel.git", targetDir)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		fmt.Printf("❌ Failed to clone RelGo framework template: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Remove .git directory
	_ = os.RemoveAll(filepath.Join(targetDir, ".git"))
	fmt.Println("2/4 🧹 Initialized fresh project directory structure...")

	// Step 3: Rename Go module references across files
	fmt.Println("3/4 🔄 Renaming Go module imports...")
	oldModule := "github.com/mkhsnw/rel"
	if err := replaceInFiles(targetDir, oldModule, newModule); err != nil {
		fmt.Printf("❌ Failed to rename module imports: %v\n", err)
		os.Exit(1)
	}

	// Step 4: Run go mod tidy
	fmt.Println("4/4 📦 Running go mod tidy...")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = targetDir
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	_ = tidyCmd.Run()

	fmt.Println("\n==========================================================================")
	fmt.Println("🎉 RelGo project scaffolded successfully!")
	fmt.Println("==========================================================================")
	fmt.Printf("Next steps:\n")
	fmt.Printf("  1. cd %s\n", projectName)
	fmt.Printf("  2. cp env.example.json env.json\n")
	fmt.Printf("  3. docker-compose up -d\n")
	fmt.Printf("  4. rel migrate up\n")
	fmt.Printf("  5. task dev\n")
	fmt.Println("==========================================================================")
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
		if ext == ".go" || ext == ".mod" || ext == ".yml" || ext == ".yaml" || ext == ".md" || ext == ".tmpl" || ext == ".json" {
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

func runSubcommand(scriptPath string, args ...string) {
	// Case 1: Execute local script if present in current working directory
	if _, err := os.Stat(scriptPath); err == nil {
		cmdArgs := append([]string{"run", scriptPath}, args...)
		cmd := exec.Command("go", cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			os.Exit(1)
		}
		return
	}

	// Case 2: Fallback to remote tool execution from GitHub framework repo
	toolPkg := strings.TrimPrefix(scriptPath, "cmd/")
	toolPkg = strings.TrimSuffix(toolPkg, "/main.go")
	remotePkg := fmt.Sprintf("github.com/mkhsnw/rel/cmd/%s@latest", toolPkg)

	cmdArgs := append([]string{"run", remotePkg}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}
