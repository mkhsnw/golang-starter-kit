package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode"
)

func printHelp() {
	fmt.Print(`
╔══════════════════════════════════════════════════════════════╗
║              				Generator CLI                      ║
╚══════════════════════════════════════════════════════════════╝

USAGE:
  go run cmd/gen/main.go [flags]
  task gen [flags]

REQUIRED FLAGS:
  --name <PascalCase>   Module name in PascalCase (e.g. Product, OrderItem)

OPTIONAL FLAGS:
  --fields <list>       Comma-separated field definitions.
                        Format: name:go_type[:sql_type] or name:go_type? (nullable)
                        Example 1: --fields "note:text?" (becomes NULL)
                        Example 2: --fields "status:string:ENUM('IN','OUT')?" (becomes NULL)

  --tx                  Generate the Usecase module to run database writes (Create,
                        Update, Delete) inside a Transaction. Recommended for
                        transactional data (e.g., Orders, Payments).

  --force               Overwrite existing files without asking.

  --dry-run             Preview all files that would be generated/modified
                        without writing any changes to disk.

  --migrate             Automatically run database migrations after generation
                        without prompting.

  --test                Generate test files (usecase_test.go, controller_test.go).
                        Automatically enabled when --tx is used, since
                        transactional logic is exactly what benefits most
                        from tests.

  --help, -h            Show this help message.

SUPPORTED FIELD TYPES:
  string, text, bool, []byte
  int, int8, int16, int32, int64
  uint, uint8, uint16, uint32, uint64
  float32, float64, time.Time

EXAMPLES:
  task gen name=Product fields="name:string,price:float64,stock:int"
  task gen name=Invoice fields="amount:float64,note:text?" tx=true
  task gen file=module-product.yml
  task rm name=Product
`)
}

func runModuleGenerator(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Masukkan nama modul. Contoh: task gen product")
		os.Exit(1)
	}

	arg := args[0]
	if arg == "--help" || arg == "-h" {
		printHelp()
		os.Exit(0)
	}

	// Assuming arg is the module name (e.g. product)
	manifestPath := fmt.Sprintf("manifests/%s.yaml", strings.ToLower(arg))
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		fmt.Printf("Gagal membaca manifest file %s: %v\n", manifestPath, err)
		os.Exit(1)
	}

	if manifest.Name == "" {
		fmt.Println("Error: 'name' wajib diisi di dalam manifest.")
		os.Exit(1)
	}

	mod := ModuleNames{
		Pascal:     toPascalCase(manifest.Name),
		Snake:      toSnakeCase(manifest.Name),
		Camel:      strings.ToLower(string(manifest.Name[0])) + toPascalCase(manifest.Name)[1:],
		Plural:     pluralize(toSnakeCase(manifest.Name)),
		IsTx:       manifest.Transactions,
		ModuleName: getModuleName(),
		IsBusiness: manifest.Type == "business",
	}

	for _, f := range manifest.Fields {
		if strings.ToLower(f.Name) == "id" || strings.ToLower(f.Name) == "created_at" || strings.ToLower(f.Name) == "updated_at" {
			continue // Skip implicit fields
		}

		sqlType := f.SqlType
		if sqlType == "" {
			sqlType = goTypeToSQLType(f.Type)
		}
		
		isForeignKey := strings.HasSuffix(f.Name, "_id")
		refTable := ""
		if isForeignKey {
			refTable = pluralize(strings.TrimSuffix(f.Name, "_id"))
		}

		goType := f.Type
		if goType == "text" {
			goType = "string"
		}
		
		mod.Fields = append(mod.Fields, Field{
			Name:            f.Name,
			Type:            goType,
			RawType:         f.Type,
			PascalName:      toPascalCase(f.Name),
			SnakeName:       f.Name,
			Nullable:        !f.Required,
			SQLType:         sqlType,
			IsForeignKey:    isForeignKey,
			ReferencedTable: refTable,
			Validate:        buildValidatorTag(f),
		})
	}

	// For manifest mode, we run migrations only if it's not a business module
	generateModule(mod, true, false, !mod.IsBusiness, manifest.Tests)
}

func buildValidatorTag(f ManifestField) string {
	if f.Required {
		return "required"
	}
	return ""
}

type Field struct {
	Name            string // "price"
	Type            string // "float64" (Go type, "text" is remapped to "string")
	RawType         string // "text" (original declared type, used for SQL mapping)
	PascalName      string // "Price"
	SnakeName       string // "price"
	Nullable        bool   // true jika tidak required
	SQLType         string // "TEXT"
	IsForeignKey    bool   // true jika suffix _id
	ReferencedTable string // "users" jika user_id
	Validate        string // "required,min=3"
}

type ModuleNames struct {
	Pascal        string // "Product"
	Snake         string // "product"
	Camel         string // "product"
	Plural        string // "products"
	Fields        []Field
	IsTx          bool   // true if transaction enabled
	ModuleName    string // e.g., "github.com/username/project"
	IsBusiness    bool   // true if business module
}

func getModuleName() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "github.com/mkhsnw/golang-starter-kit"
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return "github.com/mkhsnw/golang-starter-kit"
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

var commonInitialisms = map[string]string{
	"Id":   "ID",
	"Url":  "URL",
	"Api":  "API",
	"Json": "JSON",
	"Html": "HTML",
	"Xml":  "XML",
	"Http": "HTTP",
	"Uuid": "UUID",
	"Uri":  "URI",
	"Ip":   "IP",
}

func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			word := strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
			if upper, ok := commonInitialisms[word]; ok {
				word = upper
			}
			parts[i] = word
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

func generateModule(mod ModuleNames, force, dryRun, runMigrate, genTest bool) {
	// Validate field types early — fail fast before touching the filesystem
	validateFields(mod.Fields)

	timestamp := time.Now().Format("20060102150405")
	baseName := fmt.Sprintf("create_%s_table", mod.Plural)

	// Check if migration already exists to reuse its timestamp
	matches, _ := filepath.Glob(fmt.Sprintf("db/migration/*_%s.up.sql", baseName))
	if len(matches) > 0 {
		filename := filepath.Base(matches[0])
		parts := strings.SplitN(filename, "_", 2)
		if len(parts) == 2 {
			timestamp = parts[0]
		}
	}

	var files []fileToGenerate
	if mod.IsBusiness {
		files = []fileToGenerate{
			{"business_service.go.tmpl", fmt.Sprintf("internal/module/%s/service.go", mod.Snake)},
			{"business_controller.go.tmpl", fmt.Sprintf("internal/module/%s/controller.go", mod.Snake)},
			{"business_route.go.tmpl", fmt.Sprintf("internal/module/%s/route.go", mod.Snake)},
			{"business_dto_request.go.tmpl", fmt.Sprintf("internal/module/%s/dto/request.go", mod.Snake)},
			{"business_dto_response.go.tmpl", fmt.Sprintf("internal/module/%s/dto/response.go", mod.Snake)},
		}
	} else {
		files = []fileToGenerate{
			{"entity.go.tmpl", fmt.Sprintf("internal/module/%s/entity.go", mod.Snake)},
			{"repository.go.tmpl", fmt.Sprintf("internal/module/%s/repository.go", mod.Snake)},
			{"service.go.tmpl", fmt.Sprintf("internal/module/%s/service.go", mod.Snake)},
			{"controller.go.tmpl", fmt.Sprintf("internal/module/%s/controller.go", mod.Snake)},
			{"route.go.tmpl", fmt.Sprintf("internal/module/%s/route.go", mod.Snake)},
			{"dto_request.go.tmpl", fmt.Sprintf("internal/module/%s/dto/request.go", mod.Snake)},
			{"dto_response.go.tmpl", fmt.Sprintf("internal/module/%s/dto/response.go", mod.Snake)},
			{"mapper.go.tmpl", fmt.Sprintf("internal/module/%s/mapper.go", mod.Snake)},
			{"migration_up.sql.tmpl", fmt.Sprintf("db/migration/%s_%s.up.sql", timestamp, baseName)},
			{"migration_down.sql.tmpl", fmt.Sprintf("db/migration/%s_%s.down.sql", timestamp, baseName)},
		}
	}
	for _, f := range files {
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create %s\n", f.OutputPath)
			continue
		}
		if err := renderFile(f, mod, force); err != nil {
			fmt.Printf("Gagal generate %s: %v\n", f.OutputPath, err)
			os.Exit(1)
		}
		fmt.Printf("✓ Created %s\n", f.OutputPath)
	}

	injectToAppGo(mod, dryRun)

	if dryRun {
		fmt.Println("\n[DRY-RUN] Module generation simulated successfully!")
		return
	}

	// Run gofmt to format generated/injected Go files
	runGofmtOnGen(mod)

	fmt.Println("\nModule berhasil dibuat dan di-inject otomatis ke app.go, route.go & interfaces.go!")

	// Interactive auto-migrate prompt
	if !mod.IsBusiness && !runMigrate {
		fmt.Print("\nApakah Anda ingin menjalankan migrasi ke database sekarang? (y/N): ")
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(strings.TrimSpace(input)) == "y" {
			runMigrate = true
		}
	}

	if runMigrate {
		fmt.Println("🚀 Menjalankan migrasi database (go run cmd/migrate/main.go up)...")
		cmd := exec.Command("go", "run", "cmd/migrate/main.go", "up")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ Gagal menjalankan migrasi: %v\n", err)
		} else {
			fmt.Println("✅ Migrasi database berhasil diterapkan!")
		}
	}
}

func runGofmtOnGen(mod ModuleNames) {
	files := []string{
		"internal/config/app.go",
		"internal/delivery/http/route/route.go",
		"internal/repository/interfaces.go",
		"internal/usecase/interfaces.go",
		fmt.Sprintf("internal/module/%s/dto/request.go", mod.Snake),
		fmt.Sprintf("internal/module/%s/dto/response.go", mod.Snake),
		fmt.Sprintf("internal/module/%s/service.go", mod.Snake),
		fmt.Sprintf("internal/module/%s/controller.go", mod.Snake),
		fmt.Sprintf("internal/module/%s/route.go", mod.Snake),
	}
	if !mod.IsBusiness {
		files = append(files,
			fmt.Sprintf("internal/module/%s/entity.go", mod.Snake),
			fmt.Sprintf("internal/module/%s/mapper.go", mod.Snake),
			fmt.Sprintf("internal/module/%s/repository.go", mod.Snake),
		)
	}
	for _, f := range files {
		if _, err := os.Stat(f); err == nil {
			cmd := exec.Command("gofmt", "-w", f)
			_ = cmd.Run()
		}
	}
}

func goTypeToSQLType(goType string) string {
	switch goType {
	case "string":
		return "VARCHAR(255)"
	case "text":
		return "TEXT"
	case "int", "int32":
		return "INT"
	case "int8":
		return "TINYINT"
	case "int16":
		return "SMALLINT"
	case "int64":
		return "BIGINT"
	case "uint", "uint32":
		return "INT UNSIGNED"
	case "uint8":
		return "TINYINT UNSIGNED"
	case "uint16":
		return "SMALLINT UNSIGNED"
	case "uint64":
		return "BIGINT UNSIGNED"
	case "float32":
		return "FLOAT"
	case "float64":
		return "DOUBLE PRECISION"
	case "bool":
		return "TINYINT(1)"
	case "time.Time":
		return "TIMESTAMP"
	case "[]byte":
		return "BLOB"
	default:
		return "" // signal for unknown type
	}
}

// supportedGoTypes lists all types that can be safely generated.
var supportedGoTypes = map[string]bool{
	"string": true, "text": true,
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"float32": true, "float64": true,
	"bool": true, "time.Time": true, "[]byte": true,
}

func validateFields(fields []Field) {
	for _, f := range fields {
		rawType := f.RawType
		if rawType == "" {
			rawType = f.Type
		}
		if !supportedGoTypes[rawType] {
			fmt.Printf("❌ ERROR: Field '%s' memiliki tipe '%s' yang tidak didukung.\n", f.Name, rawType)
			fmt.Println("   Tipe yang didukung: string, text, int, int8, int16, int32, int64,")
			fmt.Println("                       uint, uint8, uint16, uint32, uint64,")
			fmt.Println("                       float32, float64, bool, time.Time, []byte")
			os.Exit(1)
		}
	}
}

func injectToAppGo(mod ModuleNames, dryRun bool) {
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
	if !mod.IsBusiness && strings.Contains(content, mod.Camel+"Repo := "+mod.Snake+".New"+mod.Pascal+"Repository") {
		fmt.Printf("⚠ %sRepository sudah ter-inject di app.go, dilewati.\n", mod.Pascal)
		return
	}
	if mod.IsBusiness && strings.Contains(content, mod.Camel+"Service := "+mod.Snake+".New"+mod.Pascal+"Service") {
		fmt.Printf("⚠ %sService sudah ter-inject di app.go, dilewati.\n", mod.Pascal)
		return
	}

	var usecaseCode string
	if mod.IsBusiness {
		if mod.IsTx {
			usecaseCode = fmt.Sprintf("%sService := %s.New%sService(config.Logger, txManager)\n\t// @InjectUsecase", mod.Camel, mod.Snake, mod.Pascal)
		} else {
			usecaseCode = fmt.Sprintf("%sService := %s.New%sService(config.Logger)\n\t// @InjectUsecase", mod.Camel, mod.Snake, mod.Pascal)
		}
	} else {
		repoCode := fmt.Sprintf("%sRepo := %s.New%sRepository(config.Database)\n\t// @InjectRepo", mod.Camel, mod.Snake, mod.Pascal)
		content = strings.Replace(content, "// @InjectRepo", repoCode, 1)

		if mod.IsTx {
			usecaseCode = fmt.Sprintf("%sService := %s.New%sService(config.Logger, txManager, %sRepo)\n\t// @InjectUsecase", mod.Camel, mod.Snake, mod.Pascal, mod.Camel)
		} else {
			usecaseCode = fmt.Sprintf("%sService := %s.New%sService(config.Logger, %sRepo)\n\t// @InjectUsecase", mod.Camel, mod.Snake, mod.Pascal, mod.Camel)
		}
	}

	controllerCode := fmt.Sprintf("%sController := %s.New%sController(%sService, config.Validator)\n\t// @InjectController", mod.Camel, mod.Snake, mod.Pascal, mod.Camel)
	routeConfigCode := fmt.Sprintf("%s.SetupRoutes(api, %sController, authMiddleware)\n\t// @InjectRouteConfig", mod.Snake, mod.Camel)

	content = strings.Replace(content, "// @InjectUsecase", usecaseCode, 1)
	content = strings.Replace(content, "// @InjectController", controllerCode, 1)
	content = strings.Replace(content, "// @InjectRouteConfig", routeConfigCode, 1)

	// Inject import
	importPath := fmt.Sprintf(`"%s/internal/module/%s"`, mod.ModuleName, mod.Snake)
	if !strings.Contains(content, importPath) {
		content = strings.Replace(content, "import (", "import (\n\t"+importPath, 1)
	}

	if dryRun {
		fmt.Println("[DRY-RUN] Would inject to app.go")
		return
	}

	os.WriteFile(appPath, []byte(content), 0644)
	fmt.Println("✓ Injected to app.go")
}
func renderFile(f fileToGenerate, mod ModuleNames, force bool) error {
	// Jangan timpa file yang sudah ada kecuali flag --force diberikan
	if _, err := os.Stat(f.OutputPath); err == nil && !force {
		return fmt.Errorf("file sudah ada, dilewati (gunakan --force untuk menimpa): %s", f.OutputPath)
	}

	tmpl, err := template.New(f.TemplateName).Funcs(template.FuncMap{
		"hasSuffix": strings.HasSuffix,
	}).ParseFS(templateFS, "templates/"+f.TemplateName)
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

