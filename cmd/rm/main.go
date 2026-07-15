package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

func main() {
	var name string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--name" && i+1 < len(args) {
			name = args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--name=") {
			name = strings.TrimPrefix(arg, "--name=")
		} else if !strings.HasPrefix(arg, "-") && name == "" {
			name = arg
		}
	}

	if name == "" {
		fmt.Println("Error: name wajib diisi. Contoh: task rm -- Product")
		os.Exit(1)
	}

	pascal := strings.ToUpper((name)[:1]) + (name)[1:]
	snake := toSnakeCase(pascal)
	camel := strings.ToLower((name)[:1]) + (name)[1:]
	plural := toPlural(snake)

	fmt.Printf("🗑️ Menghapus modul %s...\n", pascal)

	// 1. Delete Files
	filesToDelete := []string{
		fmt.Sprintf("internal/entity/%s_entity.go", snake),
		fmt.Sprintf("internal/model/%s_model.go", snake),
		fmt.Sprintf("internal/repository/%s_repository.go", snake),
		fmt.Sprintf("internal/usecase/%s_usecase.go", snake),
		fmt.Sprintf("internal/usecase/%s_usecase_test.go", snake),
		fmt.Sprintf("internal/delivery/http/controller/%s_controller.go", snake),
		fmt.Sprintf("internal/delivery/http/controller/%s_controller_test.go", snake),
	}

	for _, f := range filesToDelete {
		if err := os.Remove(f); err == nil {
			fmt.Printf("✓ Terhapus: %s\n", f)
		} else if !os.IsNotExist(err) {
			fmt.Printf("⚠ Gagal menghapus %s: %v\n", f, err)
		}
	}

	// Remove Migrations
	removeMigrations(plural)

	// 2. Remove from Interfaces
	removeFromRepoInterfaces(pascal)
	removeFromUsecaseInterfaces(pascal)

	// 3. Remove from App & Route
	removeFromAppGo(pascal, camel)
	removeFromRouteGo(pascal, plural)

	fmt.Println("✅ Modul berhasil dihapus dan dibersihkan!")
}

func removeMigrations(pluralSnake string) {
	dir := "db/migration"
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	suffixUp := fmt.Sprintf("_create_%s_table.up.sql", pluralSnake)
	suffixDown := fmt.Sprintf("_create_%s_table.down.sql", pluralSnake)

	for _, file := range files {
		if strings.HasSuffix(file.Name(), suffixUp) || strings.HasSuffix(file.Name(), suffixDown) {
			path := filepath.Join(dir, file.Name())
			if err := os.Remove(path); err == nil {
				fmt.Printf("✓ Terhapus: %s\n", path)
			}
		}
	}
}

func removeFromRepoInterfaces(pascal string) {
	path := "internal/repository/interfaces.go"
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return
	}
	content := string(contentBytes)

	pattern := fmt.Sprintf(`(?s)type %sRepositoryInterface interface \{.*?\n\}\n*`, pascal)
	re := regexp.MustCompile(pattern)
	newContent := re.ReplaceAllString(content, "")

	if content != newContent {
		os.WriteFile(path, []byte(newContent), 0644)
		fmt.Println("✓ Un-injected from repository/interfaces.go")
	}
}

func removeFromUsecaseInterfaces(pascal string) {
	path := "internal/usecase/interfaces.go"
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return
	}
	content := string(contentBytes)

	pattern := fmt.Sprintf(`(?s)type %sUsecaseInterface interface \{.*?\n\}\n*`, pascal)
	re := regexp.MustCompile(pattern)
	newContent := re.ReplaceAllString(content, "")

	if content != newContent {
		os.WriteFile(path, []byte(newContent), 0644)
		fmt.Println("✓ Un-injected from usecase/interfaces.go")
	}
}

func removeFromAppGo(pascal, camel string) {
	path := "internal/config/app.go"
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return
	}
	content := string(contentBytes)

	lines := strings.Split(content, "\n")
	var newLines []string

	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("%sRepo := repository.New%sRepository", camel, pascal)) {
			continue
		}
		if strings.Contains(line, fmt.Sprintf("%sUsecase := usecase.New%sUsecase(%sRepo)", camel, pascal, camel)) {
			continue
		}
		if strings.Contains(line, fmt.Sprintf("%sController := controller.New%sController", camel, pascal)) {
			continue
		}
		if strings.Contains(line, fmt.Sprintf("%sController: %sController,", pascal, camel)) {
			continue
		}
		newLines = append(newLines, line)
	}

	newContent := strings.Join(newLines, "\n")
	if content != newContent {
		os.WriteFile(path, []byte(newContent), 0644)
		fmt.Println("✓ Un-injected from app.go")
	}
}

func removeFromRouteGo(pascal, plural string) {
	path := "internal/delivery/http/route/route.go"
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return
	}
	content := string(contentBytes)

	// Remove struct field and setup call
	lines := strings.Split(content, "\n")
	var newLines []string

	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("%sController *controller.%sController", pascal, pascal)) {
			continue
		}
		if strings.Contains(line, fmt.Sprintf("c.setup%sRoutes(apiAuth)", pascal)) {
			continue
		}
		newLines = append(newLines, line)
	}

	newContent := strings.Join(newLines, "\n")

	// Remove setup function
	pattern := fmt.Sprintf(`(?s)func \(c \*RouteConfig\) setup%sRoutes\(api fiber\.Router\) \{.*?\n\}\n*`, pascal)
	re := regexp.MustCompile(pattern)
	newContent = re.ReplaceAllString(newContent, "")

	if content != newContent {
		os.WriteFile(path, []byte(newContent), 0644)
		fmt.Println("✓ Un-injected from route.go")
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
