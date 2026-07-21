package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

func printRmHelp() {
	fmt.Print(`
╔══════════════════════════════════════════════════════════════╗
║              golang-starter-kit Remover CLI                  ║
╚══════════════════════════════════════════════════════════════╝

USAGE:
  go run cmd/rm/main.go [flags]
  task rm [flags]

REQUIRED FLAGS:
  --name <PascalCase>   Module name to remove in PascalCase (e.g. Product, OrderItem)

OPTIONAL FLAGS:
  --dry-run             Preview all files that would be deleted/modified
                        without writing any changes to disk.

  --help, -h            Show this help message.

EXAMPLES:
  task rm name=Product
  task rm name=OrderItem dry=true
`)
}

func main() {
	var name string
	var dryRun bool

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--help" || arg == "-h" {
			printRmHelp()
			os.Exit(0)
		} else if arg == "--dry-run" {
			dryRun = true
		} else if arg == "--name" && i+1 < len(args) {
			name = args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--name=") {
			name = strings.TrimPrefix(arg, "--name=")
		} else if !strings.HasPrefix(arg, "-") && name == "" {
			name = arg
		}
	}

	if name == "" {
		fmt.Println("Error: --name wajib diisi. Contoh: task rm name=Product")
		fmt.Println("Jalankan dengan --help untuk melihat petunjuk penggunaan.")
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Gagal mendapatkan working directory: %v\n", err)
		os.Exit(1)
	}

	pascal := strings.ToUpper((name)[:1]) + (name)[1:]
	snake := toSnakeCase(pascal)
	camel := strings.ToLower((name)[:1]) + (name)[1:]
	plural := toPlural(snake)

	if dryRun {
		fmt.Printf("🔍 [DRY-RUN] Mensimulasikan penghapusan modul %s...\n", pascal)
	} else {
		fmt.Printf("🗑️ Menghapus modul %s...\n", pascal)
	}

	// 1. Delete Module Directory
	moduleDir := filepath.Join(cwd, "internal", "module", snake)
	if dryRun {
		fmt.Printf("[DRY-RUN] Would delete module directory: %s\n", moduleDir)
	} else {
		if err := os.RemoveAll(moduleDir); err == nil {
			fmt.Printf("✓ Terhapus: %s\n", moduleDir)
		} else {
			fmt.Printf("⚠ Gagal menghapus direktori modul %s: %v\n", moduleDir, err)
		}
	}

	// 2. Remove Migrations
	removeMigrations(cwd, plural, dryRun)

	// 3. Remove from App.go
	removeFromAppGo(cwd, pascal, camel, snake, dryRun)

	// 4. Run gofmt to clean up injected files if not dry-run
	if !dryRun {
		runGofmt(cwd)
		fmt.Println("✅ Modul berhasil dihapus, dibersihkan, dan diformat dengan gofmt!")
	} else {
		fmt.Println("✅ [DRY-RUN] Simulasi penghapusan modul selesai!")
	}
}

func removeMigrations(cwd, pluralSnake string, dryRun bool) {
	dir := filepath.Join(cwd, "db", "migration")
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	suffixUp := fmt.Sprintf("_create_%s_table.up.sql", pluralSnake)
	suffixDown := fmt.Sprintf("_create_%s_table.down.sql", pluralSnake)

	for _, file := range files {
		if strings.HasSuffix(file.Name(), suffixUp) || strings.HasSuffix(file.Name(), suffixDown) {
			path := filepath.Join(dir, file.Name())
			if dryRun {
				fmt.Printf("[DRY-RUN] Would delete migration: %s\n", path)
				continue
			}
			if err := os.Remove(path); err == nil {
				fmt.Printf("✓ Terhapus: %s\n", path)
			}
		}
	}
}

func removeFromAppGo(cwd, pascal, camel, snake string, dryRun bool) {
	path := filepath.Join(cwd, "internal", "config", "app.go")
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return
	}
	content := string(contentBytes)

	lines := strings.Split(content, "\n")
	var newLines []string

	// Patterns based on generator injection in cmd/gen/main.go
	patternRepo := fmt.Sprintf(`\b%sRepo\s*:=\s*%s\.New%sRepository`, camel, snake, pascal)
	patternUsecase := fmt.Sprintf(`\b%sService\s*:=\s*%s\.New%sService`, camel, snake, pascal)
	patternController := fmt.Sprintf(`\b%sController\s*:=\s*%s\.New%sController`, camel, snake, pascal)
	patternRouteConfig := fmt.Sprintf(`\b%s\.SetupRoutes\(api,\s*%sController`, snake, camel)
	patternImport := fmt.Sprintf(`"github.com/mkhsnw/golang-starter-kit/internal/module/%s"`, snake)

	reRepo := regexp.MustCompile(patternRepo)
	reUsecase := regexp.MustCompile(patternUsecase)
	reController := regexp.MustCompile(patternController)
	reRouteConfig := regexp.MustCompile(patternRouteConfig)
	reImport := regexp.MustCompile(patternImport)

	for _, line := range lines {
		if reRepo.MatchString(line) {
			continue
		}
		if reUsecase.MatchString(line) {
			continue
		}
		if reController.MatchString(line) {
			continue
		}
		if reRouteConfig.MatchString(line) {
			continue
		}
		if reImport.MatchString(line) {
			continue
		}
		newLines = append(newLines, line)
	}

	newContent := strings.Join(newLines, "\n")
	if content != newContent {
		if dryRun {
			fmt.Printf("[DRY-RUN] Would un-inject %s dependencies from app.go\n", pascal)
			return
		}
		os.WriteFile(path, []byte(newContent), 0644)
		fmt.Println("✓ Un-injected from app.go")
	}
}

func runGofmt(cwd string) {
	files := []string{
		filepath.Join(cwd, "internal", "config", "app.go"),
	}

	for _, f := range files {
		if _, err := os.Stat(f); err == nil {
			cmd := exec.Command("gofmt", "-w", f)
			_ = cmd.Run()
		}
	}
}

// -- Helpers --
func toSnakeCase(s string) string {
	var res []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				res = append(res, '_')
			}
			res = append(res, unicode.ToLower(r))
		} else {
			res = append(res, r)
		}
	}
	return string(res)
}

func toPlural(s string) string {
	if strings.HasSuffix(s, "y") && len(s) > 1 {
		switch s[len(s)-2] {
		case 'a', 'e', 'i', 'o', 'u':
			return s + "s"
		default:
			return s[:len(s)-1] + "ies"
		}
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "sh") || strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "x") || strings.HasSuffix(s, "z") {
		return s + "es"
	}
	return s + "s"
}
