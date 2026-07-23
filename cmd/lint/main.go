package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type Violation struct {
	File     string
	Line     int
	Rule     string
	Message  string
	Severity string
}

func main() {
	fmt.Println("🔍 Running Golang Starter Kit Architecture Linter (gokit lint)...")
	fmt.Println("==========================================================================")

	moduleDir := "internal/module"
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		fmt.Printf("❌ Module directory '%s' not found.\n", moduleDir)
		os.Exit(1)
	}

	var violations []Violation
	fset := token.NewFileSet()

	err := filepath.WalkDir(moduleDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		node, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if parseErr != nil {
			fmt.Printf("⚠️  Failed to parse AST for %s: %v\n", path, parseErr)
			return nil
		}

		relPath, _ := filepath.Rel(".", path)
		relPath = filepath.ToSlash(relPath)
		fileBasename := filepath.Base(path)

		// Rule Check 1: Controller Isolation & DB/Repo checks
		if strings.HasSuffix(fileBasename, "controller.go") {
			violations = append(violations, checkControllerRules(fset, node, relPath)...)
		}

		// Rule Check 2: Service Context Isolation
		if strings.HasSuffix(fileBasename, "service.go") {
			violations = append(violations, checkServiceRules(fset, node, relPath)...)
		}

		// Rule Check 3: Repository DTO Isolation
		if strings.HasSuffix(fileBasename, "repository.go") {
			violations = append(violations, checkRepositoryRules(fset, node, relPath)...)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("❌ Error walking directory: %v\n", err)
		os.Exit(1)
	}

	if len(violations) == 0 {
		fmt.Println("✅ ARCHITECTURE LINTER PASSED: All architectural boundary rules respected!")
		fmt.Println("==========================================================================")
		os.Exit(0)
	}

	fmt.Printf("❌ FOUND %d ARCHITECTURE VIOLATION(S):\n\n", len(violations))
	for i, v := range violations {
		fmt.Printf("   %d) [%s] %s:%d\n", i+1, v.Rule, v.File, v.Line)
		fmt.Printf("      -> %s\n\n", v.Message)
	}
	fmt.Println("==========================================================================")
	fmt.Println("💡 Refactor your code to adhere to starter kit clean architecture boundaries.")
	os.Exit(1)
}

func checkControllerRules(fset *token.FileSet, file *ast.File, filePath string) []Violation {
	var violations []Violation

	for _, imp := range file.Imports {
		pathVal := strings.Trim(imp.Path.Value, `"`)
		if strings.Contains(pathVal, "gorm.io/gorm") {
			pos := fset.Position(imp.Pos())
			violations = append(violations, Violation{
				File:     filePath,
				Line:     pos.Line,
				Rule:     "Rule 4 (No DB in Controller)",
				Message:  "Controllers must NOT import 'gorm.io/gorm'. Database access must be encapsulated in Repositories.",
				Severity: "ERROR",
			})
		}
	}

	ast.Inspect(file, func(n ast.Node) bool {
		field, ok := n.(*ast.Field)
		if !ok {
			return true
		}

		// Check struct field types in Controller for direct Repository dependency
		if starExpr, isStar := field.Type.(*ast.StarExpr); isStar {
			if ident, isIdent := starExpr.X.(*ast.Ident); isIdent {
				if strings.HasSuffix(ident.Name, "Repository") {
					pos := fset.Position(field.Pos())
					violations = append(violations, Violation{
						File:     filePath,
						Line:     pos.Line,
						Rule:     "Rule 1 (Controller Repository Bypass)",
						Message:  fmt.Sprintf("Controller struct contains field '%s' directly. Controllers must depend on Service layer, not Repository directly.", ident.Name),
						Severity: "ERROR",
					})
				}
			}
		}
		return true
	})

	return violations
}

func checkServiceRules(fset *token.FileSet, file *ast.File, filePath string) []Violation {
	var violations []Violation

	for _, imp := range file.Imports {
		pathVal := strings.Trim(imp.Path.Value, `"`)
		if strings.Contains(pathVal, "github.com/gofiber/fiber") {
			pos := fset.Position(imp.Pos())
			violations = append(violations, Violation{
				File:     filePath,
				Line:     pos.Line,
				Rule:     "Rule 2 (No Fiber in Service)",
				Message:  "Services must NOT import 'github.com/gofiber/fiber'. Web framework context (fiber.Ctx) must not leak into Service layer.",
				Severity: "ERROR",
			})
		}
	}

	return violations
}

func checkRepositoryRules(fset *token.FileSet, file *ast.File, filePath string) []Violation {
	var violations []Violation

	for _, imp := range file.Imports {
		pathVal := strings.Trim(imp.Path.Value, `"`)
		if strings.HasSuffix(pathVal, "/dto") {
			pos := fset.Position(imp.Pos())
			violations = append(violations, Violation{
				File:     filePath,
				Line:     pos.Line,
				Rule:     "Rule 3 (No DTO in Repository)",
				Message:  "Repositories must NOT import HTTP DTO packages. Repositories should only accept and return domain Entities or QueryFilter structs.",
				Severity: "ERROR",
			})
		}
	}

	return violations
}
