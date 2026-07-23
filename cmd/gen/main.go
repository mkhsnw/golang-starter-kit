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

COMMANDS & FLAGS:
  task gen name=<Name>             Generate complete CRUD module + Migration + Seeder + Factory
  task make-seeder name=<Name>     Generate Seeder (db/seed/<name>_seeder.go) and register in seeder.go
  task make-factory name=<Name>    Generate Data Factory (db/factory/<name>_factory.go)
  task make-migration name=<Name>  Generate custom SQL migration files (.up.sql & .down.sql)
  task rm name=<Name>              Remove a generated module

OPTIONAL FLAGS FOR GEN:
  --fields <list>       Comma-separated field definitions.
                        Format: name:go_type[:sql_type] or name:go_type? (nullable)
  --tx                  Generate Usecase with Database Transaction support
  --force               Overwrite existing files without asking
  --dry-run             Preview generated files without writing to disk
  --migrate             Run database migrations automatically
  --help, -h            Show this help message

EXAMPLES:
  task gen name=Product fields="name:string,price:float64,stock:int"
  task make-seeder name=Product
  task make-factory name=Product
  task make-migration name=change_publish_year_type
`)
}

func cleanModuleNameArg(arg string) string {
	if strings.HasPrefix(arg, "name=") {
		return strings.TrimPrefix(arg, "name=")
	}
	return arg
}

func runModuleGenerator(args []string) {
	if len(args) == 0 || args[0] == "--interactive" || args[0] == "-i" {
		moduleName, _ := runInteractiveWizard()
		args = []string{moduleName}
	}

	arg := args[0]
	if arg == "--help" || arg == "-h" {
		printHelp()
		os.Exit(0)
	}

	if arg == "--make-seeder" || arg == "make-seeder" {
		if len(args) < 2 {
			fmt.Println("Error: Masukkan nama modul. Contoh: task make-seeder name=Product")
			os.Exit(1)
		}
		makeSeeder(cleanModuleNameArg(args[1]))
		return
	}

	if arg == "--make-factory" || arg == "make-factory" {
		if len(args) < 2 {
			fmt.Println("Error: Masukkan nama modul. Contoh: task make-factory name=Product")
			os.Exit(1)
		}
		makeFactory(cleanModuleNameArg(args[1]))
		return
	}

	if arg == "--make-migration" || arg == "make-migration" || arg == "--migrate-create" || arg == "migrate-create" {
		if len(args) < 2 {
			fmt.Println("Error: Masukkan nama migrasi. Contoh: task make-migration name=add_phone_to_users")
			os.Exit(1)
		}
		makeMigration(cleanModuleNameArg(args[1]))
		return
	}

	moduleArg := cleanModuleNameArg(arg)

	// Assuming arg is the module name (e.g. product)
	manifestPath := fmt.Sprintf("manifests/%s.yaml", strings.ToLower(moduleArg))
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
		} else if goType == "time" {
			goType = "time.Time"
		}

		if goType == "time.Time" {
			mod.HasTime = true
		}

		if isForeignKey {
			mod.HasForeignKey = true
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

	generateModule(mod, true, false, false, manifest.Tests)
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
	HasTime       bool   // true if any field uses time.Time
	HasForeignKey bool   // true if any field is a foreign key
}

func getModuleName() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "github.com/mkhsnw/rel"
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return "github.com/mkhsnw/rel"
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
			{"factory.go.tmpl", fmt.Sprintf("db/factory/%s_factory.go", mod.Snake)},
			{"seeder.go.tmpl", fmt.Sprintf("db/seed/%s_seeder.go", mod.Snake)},
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
	injectSeederToRegistry(mod, dryRun)

	if dryRun {
		fmt.Println("\n[DRY-RUN] Module generation simulated successfully!")
		return
	}

	// Run gofmt to format generated/injected Go files
	runGofmtOnGen(mod)

	fmt.Printf("\n✨ Modul '%s' berhasil dibuat dan di-inject otomatis!\n", mod.Pascal)
	fmt.Println("📌 Langkah Selanjutnya:")
	if !mod.IsBusiness {
		fmt.Printf("   1. Peninjauan SQL : Periksa file migrasi di db/migration/%s_%s.up.sql\n", timestamp, baseName)
		fmt.Println("   2. Jalankan Migrasi: task migrate-up")
		fmt.Println("   3. Isi Data Dummy : task seed")
	} else {
		fmt.Println("   1. Modul bisnis siap digunakan di internal/config/app.go")
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
			fmt.Sprintf("db/factory/%s_factory.go", mod.Snake),
			fmt.Sprintf("db/seed/%s_seeder.go", mod.Snake),
			"db/seed/seeder.go",
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
	case "time", "time.Time":
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
	"bool": true, "time": true, "time.Time": true, "[]byte": true,
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
			fmt.Println("                       float32, float64, bool, time, time.Time, []byte")
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

func injectSeederToRegistry(mod ModuleNames, dryRun bool) {
	if mod.IsBusiness {
		return
	}
	seederPath := "db/seed/seeder.go"
	contentBytes, err := os.ReadFile(seederPath)
	if err != nil {
		return
	}
	content := string(contentBytes)
	seederRegister := fmt.Sprintf("registry.Register(New%sSeeder())\n\t// @InjectSeeder", mod.Pascal)
	if strings.Contains(content, fmt.Sprintf("New%sSeeder()", mod.Pascal)) {
		return
	}
	if !strings.Contains(content, "// @InjectSeeder") {
		fmt.Println("⚠️  WARNING: Marker // @InjectSeeder tidak ditemukan di db/seed/seeder.go!")
		return
	}
	content = strings.Replace(content, "// @InjectSeeder", seederRegister, 1)

	if dryRun {
		fmt.Println("[DRY-RUN] Would inject seeder to db/seed/seeder.go")
		return
	}

	if err := os.WriteFile(seederPath, []byte(content), 0644); err == nil {
		fmt.Printf("✓ Injected %sSeeder to db/seed/seeder.go\n", mod.Pascal)
	}
}

func getModuleNamesForName(name string) ModuleNames {
	manifestPath := fmt.Sprintf("manifests/%s.yaml", strings.ToLower(name))
	manifest, err := LoadManifest(manifestPath)
	if err == nil && manifest.Name != "" {
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
				continue
			}
			goType := f.Type
			if goType == "text" {
				goType = "string"
			} else if goType == "time" {
				goType = "time.Time"
			}
			isForeignKey := strings.HasSuffix(f.Name, "_id")
			refTable := ""
			if isForeignKey {
				refTable = pluralize(strings.TrimSuffix(f.Name, "_id"))
				mod.HasForeignKey = true
			}
			mod.Fields = append(mod.Fields, Field{
				Name:            f.Name,
				Type:            goType,
				PascalName:      toPascalCase(f.Name),
				SnakeName:       f.Name,
				IsForeignKey:    isForeignKey,
				ReferencedTable: refTable,
			})
		}
		return mod
	}

	pascal := toPascalCase(name)
	snake := toSnakeCase(name)
	return ModuleNames{
		Pascal:     pascal,
		Snake:      snake,
		Camel:      strings.ToLower(string(pascal[0])) + pascal[1:],
		Plural:     pluralize(snake),
		ModuleName: getModuleName(),
	}
}

func makeFactory(name string) {
	mod := getModuleNamesForName(name)
	f := fileToGenerate{
		TemplateName: "factory.go.tmpl",
		OutputPath:   fmt.Sprintf("db/factory/%s_factory.go", mod.Snake),
	}
	if err := renderFile(f, mod, true); err != nil {
		fmt.Printf("❌ Gagal generate factory: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Created %s\n", f.OutputPath)
	cmd := exec.Command("gofmt", "-w", f.OutputPath)
	_ = cmd.Run()
}

func makeSeeder(name string) {
	mod := getModuleNamesForName(name)
	factoryPath := fmt.Sprintf("db/factory/%s_factory.go", mod.Snake)
	if _, err := os.Stat(factoryPath); os.IsNotExist(err) {
		makeFactory(name)
	}

	f := fileToGenerate{
		TemplateName: "seeder.go.tmpl",
		OutputPath:   fmt.Sprintf("db/seed/%s_seeder.go", mod.Snake),
	}
	if err := renderFile(f, mod, true); err != nil {
		fmt.Printf("❌ Gagal generate seeder: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Created %s\n", f.OutputPath)
	injectSeederToRegistry(mod, false)

	cmd := exec.Command("gofmt", "-w", f.OutputPath, "db/seed/seeder.go")
	_ = cmd.Run()
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

type MigrationData struct {
	Name      string
	Timestamp string
}

func makeMigration(name string) {
	snakeName := toSnakeCase(name)
	timestamp := time.Now().Format("20060102150405")
	data := MigrationData{
		Name:      snakeName,
		Timestamp: timestamp,
	}

	upPath := fmt.Sprintf("db/migration/%s_%s.up.sql", timestamp, snakeName)
	downPath := fmt.Sprintf("db/migration/%s_%s.down.sql", timestamp, snakeName)

	if err := renderCustomTemplate("migration_custom_up.sql.tmpl", upPath, data); err != nil {
		fmt.Printf("❌ Gagal generate up migration: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Created %s\n", upPath)

	if err := renderCustomTemplate("migration_custom_down.sql.tmpl", downPath, data); err != nil {
		fmt.Printf("❌ Gagal generate down migration: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Created %s\n", downPath)
}

func renderCustomTemplate(tmplName, outputPath string, data interface{}) error {
	tmpl, err := template.New(tmplName).ParseFS(templateFS, "templates/"+tmplName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	return tmpl.Execute(out, data)
}

func main() {
	runModuleGenerator(os.Args[1:])
}
