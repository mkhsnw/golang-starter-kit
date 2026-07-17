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
║              Generator CLI                                   ║
╚══════════════════════════════════════════════════════════════╝

USAGE:
  go run cmd/gen/main.go [flags]
  task gen [flags]

REQUIRED FLAGS:
  --name <PascalCase>   Module name in PascalCase (e.g. Product, OrderItem)

OPTIONAL FLAGS:
  --fields <list>       Comma-separated field definitions.
                        Format: name:type or name:type? (nullable)
                        Example: --fields "title:string,price:float64,note:text?"

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
  task gen name=Tag dry=true
  task rm name=Product
`)
}

func runModuleGenerator(args []string) {
	var name, fieldsStr string
	var force, dryRun, isTx, runMigrate, genTest bool

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--help" || arg == "-h" {
			printHelp()
			os.Exit(0)
		} else if arg == "--force" {
			force = true
		} else if arg == "--dry-run" {
			dryRun = true
		} else if arg == "--tx" {
			isTx = true
		} else if arg == "--test" {
			genTest = true
		} else if arg == "--migrate" {
			runMigrate = true
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
		fmt.Println("Jalankan dengan --help untuk melihat semua opsi yang tersedia.")
		os.Exit(1)
	}

	if strings.Contains(name, "_") || strings.ToLower((name)[:1]) == (name)[:1] {
		fmt.Println("⚠️  Gunakan PascalCase: --name OrderItem, bukan order_item atau orderItem")
		os.Exit(1)
	}

	mod := buildModuleNames(name, fieldsStr, isTx)

	if isTx {
		genTest = true
	}

	generateModule(mod, force, dryRun, runMigrate, genTest)
}

type Field struct {
	Name            string // "price"
	Type            string // "float64" (Go type, "text" is remapped to "string")
	RawType         string // "text" (original declared type, used for SQL mapping)
	PascalName      string // "Price"
	SnakeName       string // "price"
	Nullable        bool   // true jika ada '?' di type
	SQLType         string // "TEXT"
	IsForeignKey    bool   // true jika suffix _id
	ReferencedTable string // "users" jika user_id
}

type ModuleNames struct {
	Pascal     string // "Product"
	Snake      string // "product"
	Camel      string // "product"
	Plural     string // "products"
	Fields     []Field
	IsTx       bool   // true if --tx flag is passed
	ModuleName string // e.g., "github.com/username/project"
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

func buildModuleNames(input, fieldsStr string, isTx bool) ModuleNames {
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
				goType := typ
				// "text" is a SQL hint — Go type stays string
				if typ == "text" {
					goType = "string"
				}
				isFK := strings.HasSuffix(toSnakeCase(name), "_id")
				refTable := ""
				if isFK {
					refTable = pluralize(strings.TrimSuffix(toSnakeCase(name), "_id"))
				}
				sqlType := goTypeToSQLType(typ)
				if isFK {
					sqlType = "VARCHAR(36)"
				}
				fields = append(fields, Field{
					Name:            name,
					Type:            goType,
					RawType:         typ,
					PascalName:      toPascalCase(name),
					SnakeName:       toSnakeCase(name),
					Nullable:        nullable,
					SQLType:         sqlType,
					IsForeignKey:    isFK,
					ReferencedTable: refTable,
				})
			}
		}
	}

	return ModuleNames{
		Pascal:     pascal,
		Snake:      snake,
		Camel:      camel,
		Plural:     pluralize(snake),
		Fields:     fields,
		IsTx:       isTx,
		ModuleName: getModuleName(),
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

func generateModule(mod ModuleNames, force, dryRun, runMigrate, genTest bool) {
	// Validate field types early — fail fast before touching the filesystem
	validateFields(mod.Fields)

	timestamp := time.Now().Format("20060102150405")
	baseName := fmt.Sprintf("create_%s_table", mod.Plural)

	files := []fileToGenerate{
		{"entity.go.tmpl", fmt.Sprintf("internal/entity/%s_entity.go", mod.Snake)},
		{"model.go.tmpl", fmt.Sprintf("internal/model/%s_model.go", mod.Snake)},
		{"repository.go.tmpl", fmt.Sprintf("internal/repository/%s_repository.go", mod.Snake)},
		{"usecase.go.tmpl", fmt.Sprintf("internal/usecase/%s_usecase.go", mod.Snake)},
		{"controller.go.tmpl", fmt.Sprintf("internal/delivery/http/controller/%s_controller.go", mod.Snake)},
		{"migration_up.sql.tmpl", fmt.Sprintf("db/migration/%s_%s.up.sql", timestamp, baseName)},
		{"migration_down.sql.tmpl", fmt.Sprintf("db/migration/%s_%s.down.sql", timestamp, baseName)},
	}

	if genTest {
		files = append(files,
			fileToGenerate{"usecase_test.go.tmpl", fmt.Sprintf("internal/usecase/%s_usecase_test.go", mod.Snake)},
			fileToGenerate{"controller_test.go.tmpl", fmt.Sprintf("internal/delivery/http/controller/%s_controller_test.go", mod.Snake)},
		)
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
	injectToRouteGo(mod, dryRun)
	injectToRepositoryInterfaces(mod, dryRun)
	injectToUsecaseInterfaces(mod, dryRun)

	if dryRun {
		fmt.Println("\n[DRY-RUN] Module generation simulated successfully!")
		return
	}

	// Run gofmt to format generated/injected Go files
	runGofmtOnGen(mod)

	fmt.Println("\nModule berhasil dibuat dan di-inject otomatis ke app.go, route.go & interfaces.go!")

	// Interactive auto-migrate prompt
	if !runMigrate {
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
		fmt.Sprintf("internal/entity/%s_entity.go", mod.Snake),
		fmt.Sprintf("internal/model/%s_model.go", mod.Snake),
		fmt.Sprintf("internal/repository/%s_repository.go", mod.Snake),
		fmt.Sprintf("internal/usecase/%s_usecase.go", mod.Snake),
		fmt.Sprintf("internal/delivery/http/controller/%s_controller.go", mod.Snake),
		fmt.Sprintf("internal/usecase/%s_usecase_test.go", mod.Snake),
		fmt.Sprintf("internal/delivery/http/controller/%s_controller_test.go", mod.Snake),
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
	if strings.Contains(content, mod.Camel+"Repo := repository.New"+mod.Pascal+"Repository") {
		fmt.Printf("⚠ %sRepository sudah ter-inject di app.go, dilewati.\n", mod.Pascal)
		return
	}

	repoCode := fmt.Sprintf("%sRepo := repository.New%sRepository(config.Database)\n\t// @InjectRepo", mod.Camel, mod.Pascal)
	var usecaseCode string
	if mod.IsTx {
		usecaseCode = fmt.Sprintf("%sUsecase := usecase.New%sUsecase(config.Logger, txManager, %sRepo)\n\t// @InjectUsecase", mod.Camel, mod.Pascal, mod.Camel)
	} else {
		usecaseCode = fmt.Sprintf("%sUsecase := usecase.New%sUsecase(config.Logger, %sRepo)\n\t// @InjectUsecase", mod.Camel, mod.Pascal, mod.Camel)
	}
	controllerCode := fmt.Sprintf("%sController := controller.New%sController(%sUsecase, config.Validator)\n\t// @InjectController", mod.Camel, mod.Pascal, mod.Camel)
	routeConfigCode := fmt.Sprintf("%sController: %sController,\n\t\t// @InjectRouteConfig", mod.Pascal, mod.Camel)

	content = strings.Replace(content, "// @InjectRepo", repoCode, 1)
	content = strings.Replace(content, "// @InjectUsecase", usecaseCode, 1)
	content = strings.Replace(content, "// @InjectController", controllerCode, 1)
	content = strings.Replace(content, "// @InjectRouteConfig", routeConfigCode, 1)

	if dryRun {
		fmt.Println("[DRY-RUN] Would inject to app.go")
		return
	}

	os.WriteFile(appPath, []byte(content), 0644)
	fmt.Println("✓ Injected to app.go")
}

func injectToRouteGo(mod ModuleNames, dryRun bool) {
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

	if dryRun {
		fmt.Println("[DRY-RUN] Would inject to route.go")
		return
	}

	os.WriteFile(routePath, []byte(content), 0644)
	fmt.Println("✓ Injected to route.go")
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

func injectToRepositoryInterfaces(mod ModuleNames, dryRun bool) {
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

	if dryRun {
		fmt.Println("[DRY-RUN] Would inject to repository/interfaces.go")
		return
	}

	os.WriteFile(filePath, []byte(content), 0644)
	fmt.Println("✓ Injected to repository/interfaces.go")
}

func injectToUsecaseInterfaces(mod ModuleNames, dryRun bool) {
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

	interfaceCode := fmt.Sprintf("type %sUsecaseInterface interface {\n\tCreate(ctx context.Context, req *model.Create%sRequest) (*model.%sResponse, error)\n\tGetByID(ctx context.Context, id string) (*model.%sResponse, error)\n\tGetAll(ctx context.Context, cursor string, size int) ([]model.%sResponse, *string, error)\n\tUpdate(ctx context.Context, id string, req *model.Update%sRequest) (*model.%sResponse, error)\n\tDelete(ctx context.Context, id string) error\n}\n\n// @InjectUsecaseInterface", mod.Pascal, mod.Pascal, mod.Pascal, mod.Pascal, mod.Pascal, mod.Pascal, mod.Pascal)
	content = strings.Replace(content, "// @InjectUsecaseInterface", interfaceCode, 1)

	if dryRun {
		fmt.Println("[DRY-RUN] Would inject to usecase/interfaces.go")
		return
	}

	os.WriteFile(filePath, []byte(content), 0644)
	fmt.Println("✓ Injected to usecase/interfaces.go")
}
