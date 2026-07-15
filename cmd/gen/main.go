package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode"
)

func runModuleGenerator(args []string) {
	var name, fieldsStr string
	var force bool

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--force" {
			force = true
		} else if arg == "--name" && i+1 < len(args) {
			name = args[i+1]
			i++
		} else if arg == "--fields" && i+1 < len(args) {
			fieldsStr = args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--name=") {
			name = strings.TrimPrefix(arg, "--name=")
		} else if strings.HasPrefix(arg, "--fields=") {
			fieldsStr = strings.TrimPrefix(arg, "--fields=")
		} else if !strings.HasPrefix(arg, "-") && name == "" {
			name = arg
		}
	}

	if name == "" {
		fmt.Println("Error: --name wajib diisi. Contoh: --name Product")
		os.Exit(1)
	}

	if strings.Contains(name, "_") || strings.ToLower((name)[:1]) == (name)[:1] {
		fmt.Println("⚠️  Gunakan PascalCase: --name OrderItem, bukan order_item atau orderItem")
		os.Exit(1)
	}

	mod := buildModuleNames(name, fieldsStr)
	generateModule(mod, force)
}

type Field struct {
	Name       string // "price"
	Type       string // "float64"
	PascalName string // "Price"
	SnakeName  string // "price"
	Nullable   bool   // true jika ada '?' di type
}

type ModuleNames struct {
	Pascal string  // "Product"      -> nama struct, tipe
	Snake  string  // "product"      -> nama file, nama tabel
	Camel  string  // "product"      -> nama variabel lokal (kalau 1 kata, sama dgn snake)
	Plural string  // "products"     -> nama tabel, nama route group
	Fields []Field // Daftar field dinamis
}

func buildModuleNames(input, fieldsStr string) ModuleNames {
	pascal := strings.ToUpper(input[:1]) + input[1:]
	snake := toSnakeCase(input)
	camel := strings.ToLower(input[:1]) + input[1:]

	var fields []Field
	if fieldsStr != "" {
		parts := strings.Split(fieldsStr, ",")
		for _, part := range parts {
			kv := strings.Split(part, ":")
			if len(kv) == 2 {
				name := strings.TrimSpace(kv[0])
				typ := strings.TrimSpace(kv[1])
				nullable := false
				if strings.HasSuffix(typ, "?") {
					nullable = true
					typ = strings.TrimSuffix(typ, "?")
				}
				fields = append(fields, Field{
					Name:       name,
					Type:       typ,
					PascalName: toPascalCase(name),
					SnakeName:  toSnakeCase(name),
					Nullable:   nullable,
				})
			}
		}
	}

	return ModuleNames{
		Pascal: pascal,
		Snake:  snake,
		Camel:  camel,
		Plural: pluralize(snake),
		Fields: fields,
	}
}

func pluralize(s string) string {
	if strings.HasSuffix(s, "y") && len(s) > 1 {
		prev := s[len(s)-2]
		if prev != 'a' && prev != 'e' && prev != 'i' && prev != 'o' && prev != 'u' {
			return s[:len(s)-1] + "ies"
		}
	}
	return s + "s"
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

//go:embed templates/*.tmpl
var templateFS embed.FS

type fileToGenerate struct {
	TemplateName string // nama file template, misal "entity.go.tmpl"
	OutputPath   string // path tujuan, misal "internal/entity/product_entity.go"
}

func generateModule(mod ModuleNames, force bool) {
	files := []fileToGenerate{
		{"entity.go.tmpl", fmt.Sprintf("internal/entity/%s_entity.go", mod.Snake)},
		{"model.go.tmpl", fmt.Sprintf("internal/model/%s_model.go", mod.Snake)},
		{"repository.go.tmpl", fmt.Sprintf("internal/repository/%s_repository.go", mod.Snake)},
		{"usecase.go.tmpl", fmt.Sprintf("internal/usecase/%s_usecase.go", mod.Snake)},
		{"controller.go.tmpl", fmt.Sprintf("internal/delivery/http/controller/%s_controller.go", mod.Snake)},
		{"usecase_test.go.tmpl", fmt.Sprintf("internal/usecase/%s_usecase_test.go", mod.Snake)},
		{"controller_test.go.tmpl", fmt.Sprintf("internal/delivery/http/controller/%s_controller_test.go", mod.Snake)},
	}

	for _, f := range files {
		if err := renderFile(f, mod, force); err != nil {
			fmt.Printf("Gagal generate %s: %v\n", f.OutputPath, err)
			os.Exit(1)
		}
		fmt.Printf("✓ Created %s\n", f.OutputPath)
	}

	injectToAppGo(mod)
	injectToRouteGo(mod)
	injectToRepositoryInterfaces(mod)
	injectToUsecaseInterfaces(mod)
	generateMigration(mod)

	fmt.Println("\nModule berhasil dibuat dan di-inject otomatis ke app.go, route.go & interfaces.go!")
}

func goTypeToSQLType(goType string) string {
	switch goType {
	case "string":
		return "VARCHAR(255)"
	case "int", "int32":
		return "INT"
	case "int64":
		return "BIGINT"
	case "uint", "uint32":
		return "INT UNSIGNED"
	case "uint64":
		return "BIGINT UNSIGNED"
	case "float32", "float64":
		return "DOUBLE PRECISION"
	case "bool":
		return "BOOLEAN"
	case "time.Time":
		return "TIMESTAMP"
	default:
		return "VARCHAR(255)" // Default
	}
}

func generateMigration(mod ModuleNames) {
	timestamp := time.Now().Format("20060102150405")
	baseName := fmt.Sprintf("create_%s_table", mod.Plural)

	upFile := fmt.Sprintf("db/migration/%s_%s.up.sql", timestamp, baseName)
	downFile := fmt.Sprintf("db/migration/%s_%s.down.sql", timestamp, baseName)

	var upSQL strings.Builder
	upSQL.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", mod.Plural))
	upSQL.WriteString("    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,\n")

	for _, f := range mod.Fields {
		nullConstraint := "NOT NULL"
		if f.Nullable {
			nullConstraint = "NULL"
		}
		upSQL.WriteString(fmt.Sprintf("    %s %s %s,\n", f.SnakeName, goTypeToSQLType(f.Type), nullConstraint))
	}

	upSQL.WriteString("    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,\n")
	upSQL.WriteString("    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP\n")
	upSQL.WriteString(");\n")

	// Generate indexes for foreign keys (ending with _id)
	for _, f := range mod.Fields {
		if strings.HasSuffix(f.SnakeName, "_id") {
			upSQL.WriteString(fmt.Sprintf("\nCREATE INDEX idx_%s_%s ON %s(%s);", mod.Plural, f.SnakeName, mod.Plural, f.SnakeName))
		}
	}

	downSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", mod.Plural)

	os.MkdirAll("db/migration", 0755)

	if err := os.WriteFile(upFile, []byte(upSQL.String()), 0644); err != nil {
		fmt.Printf("Gagal membuat migration up: %v\n", err)
	} else {
		fmt.Printf("✓ Created %s\n", upFile)
	}

	if err := os.WriteFile(downFile, []byte(downSQL), 0644); err != nil {
		fmt.Printf("Gagal membuat migration down: %v\n", err)
	} else {
		fmt.Printf("✓ Created %s\n", downFile)
	}
}

func injectToAppGo(mod ModuleNames) {
	appPath := "internal/config/app.go"
	contentBytes, err := os.ReadFile(appPath)
	if err != nil {
		fmt.Printf("Gagal membaca app.go: %v\n", err)
		return
	}
	content := string(contentBytes)

	if !strings.Contains(content, "// @InjectRepo") {
		fmt.Println("⚠️  WARNING: Marker // @InjectRepo tidak ditemukan di app.go!")
		fmt.Println("   Kode TIDAK ter-inject. Tambahkan marker secara manual.")
		return
	}

	// Guard: skip if already injected
	if strings.Contains(content, mod.Camel+"Repo := repository.New"+mod.Pascal+"Repository") {
		fmt.Printf("⚠ %sRepository sudah ter-inject di app.go, dilewati.\n", mod.Pascal)
		return
	}

	repoCode := fmt.Sprintf("%sRepo := repository.New%sRepository(config.Database)\n\t// @InjectRepo", mod.Camel, mod.Pascal)
	usecaseCode := fmt.Sprintf("%sUsecase := usecase.New%sUsecase(%sRepo)\n\t// @InjectUsecase", mod.Camel, mod.Pascal, mod.Camel)
	controllerCode := fmt.Sprintf("%sController := controller.New%sController(%sUsecase, config.Validator)\n\t// @InjectController", mod.Camel, mod.Pascal, mod.Camel)
	routeConfigCode := fmt.Sprintf("%sController: %sController,\n\t\t// @InjectRouteConfig", mod.Pascal, mod.Camel)

	content = strings.Replace(content, "// @InjectRepo", repoCode, 1)
	content = strings.Replace(content, "// @InjectUsecase", usecaseCode, 1)
	content = strings.Replace(content, "// @InjectController", controllerCode, 1)
	content = strings.Replace(content, "// @InjectRouteConfig", routeConfigCode, 1)

	os.WriteFile(appPath, []byte(content), 0644)
	fmt.Println("✓ Injected to app.go")
}

func injectToRouteGo(mod ModuleNames) {
	routePath := "internal/delivery/http/route/route.go"
	contentBytes, err := os.ReadFile(routePath)
	if err != nil {
		fmt.Printf("Gagal membaca route.go: %v\n", err)
		return
	}
	content := string(contentBytes)

	// Guard: skip if already injected
	if strings.Contains(content, mod.Pascal+"Controller *controller."+mod.Pascal+"Controller") {
		fmt.Printf("⚠ %sController sudah ada di route.go, dilewati.\n", mod.Pascal)
		return
	}

	if !strings.Contains(content, "// @InjectRouteStruct") {
		fmt.Println("⚠️  WARNING: Marker // @InjectRouteStruct tidak ditemukan di route.go!")
		fmt.Println("   Kode TIDAK ter-inject. Tambahkan marker secara manual.")
		return
	}

	routeStructCode := fmt.Sprintf("%sController *controller.%sController\n\t// @InjectRouteStruct", mod.Pascal, mod.Pascal)
	routeSetupCode := fmt.Sprintf("c.setup%sRoutes(apiAuth)\n\t// @InjectRouteSetup", mod.Pascal)

	content = strings.Replace(content, "// @InjectRouteStruct", routeStructCode, 1)
	content = strings.Replace(content, "// @InjectRouteSetup", routeSetupCode, 1)

	// Add the setup function at the end of the file
	setupFunc := fmt.Sprintf(`

func (c *RouteConfig) setup%sRoutes(api fiber.Router) {
	%s := api.Group("/%s")
	
	%s.Post("/", c.%sController.Create)
	%s.Get("/", c.%sController.GetAll)
	%s.Get("/:id", c.%sController.GetByID)
	%s.Put("/:id", c.%sController.Update)
	%s.Delete("/:id", c.%sController.Delete)
}
`, mod.Pascal, mod.Plural, mod.Plural, mod.Plural, mod.Pascal, mod.Plural, mod.Pascal, mod.Plural, mod.Pascal, mod.Plural, mod.Pascal, mod.Plural, mod.Pascal)

	content += setupFunc

	os.WriteFile(routePath, []byte(content), 0644)
	fmt.Println("✓ Injected to route.go")
}

func renderFile(f fileToGenerate, mod ModuleNames, force bool) error {
	// Jangan timpa file yang sudah ada kecuali flag --force diberikan
	if _, err := os.Stat(f.OutputPath); err == nil && !force {
		return fmt.Errorf("file sudah ada, dilewati (gunakan --force untuk menimpa): %s", f.OutputPath)
	}

	tmpl, err := template.ParseFS(templateFS, "templates/"+f.TemplateName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(f.OutputPath), 0755); err != nil {
		return err
	}

	out, err := os.Create(f.OutputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	return tmpl.Execute(out, mod)
}

func main() {
	runModuleGenerator(os.Args[1:])
}

func injectToRepositoryInterfaces(mod ModuleNames) {
	filePath := "internal/repository/interfaces.go"
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Gagal membaca interfaces.go (repository): %v\n", err)
		return
	}
	content := string(contentBytes)

	// Guard: skip if interface already exists
	if strings.Contains(content, "type "+mod.Pascal+"RepositoryInterface interface") {
		fmt.Printf("⚠ %sRepositoryInterface sudah ada di repository/interfaces.go, dilewati.\n", mod.Pascal)
		return
	}

	if !strings.Contains(content, "// @InjectRepositoryInterface") {
		fmt.Println("⚠️  WARNING: Marker // @InjectRepositoryInterface tidak ditemukan di repository/interfaces.go!")
		return
	}

	interfaceCode := fmt.Sprintf("type %sRepositoryInterface interface {\n\tRepositoryInterface[entity.%s]\n}\n\n// @InjectRepositoryInterface", mod.Pascal, mod.Pascal)
	content = strings.Replace(content, "// @InjectRepositoryInterface", interfaceCode, 1)

	os.WriteFile(filePath, []byte(content), 0644)
	fmt.Println("✓ Injected to repository/interfaces.go")
}

func injectToUsecaseInterfaces(mod ModuleNames) {
	filePath := "internal/usecase/interfaces.go"
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Gagal membaca interfaces.go (usecase): %v\n", err)
		return
	}
	content := string(contentBytes)

	// Guard: skip if interface already exists
	if strings.Contains(content, "type "+mod.Pascal+"UsecaseInterface interface") {
		fmt.Printf("⚠ %sUsecaseInterface sudah ada di usecase/interfaces.go, dilewati.\n", mod.Pascal)
		return
	}

	if !strings.Contains(content, "// @InjectUsecaseInterface") {
		fmt.Println("⚠️  WARNING: Marker // @InjectUsecaseInterface tidak ditemukan di usecase/interfaces.go!")
		return
	}

	interfaceCode := fmt.Sprintf("type %sUsecaseInterface interface {\n\tCreate(ctx context.Context, req *model.Create%sRequest) (*model.%sResponse, error)\n\tGetByID(ctx context.Context, id uint64) (*model.%sResponse, error)\n\tGetAll(ctx context.Context, page, size int) ([]model.%sResponse, int64, error)\n\tUpdate(ctx context.Context, id uint64, req *model.Update%sRequest) (*model.%sResponse, error)\n\tDelete(ctx context.Context, id uint64) error\n}\n\n// @InjectUsecaseInterface", mod.Pascal, mod.Pascal, mod.Pascal, mod.Pascal, mod.Pascal, mod.Pascal, mod.Pascal)
	content = strings.Replace(content, "// @InjectUsecaseInterface", interfaceCode, 1)

	os.WriteFile(filePath, []byte(content), 0644)
	fmt.Println("✓ Injected to usecase/interfaces.go")
}
