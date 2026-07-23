package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func runInteractiveWizard() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🧙‍♂️ Welcome to Golang Starter Kit Interactive Module Generator!")
	fmt.Println("==========================================================================")

	// 1. Module Name
	fmt.Print("👉 Enter Module Name (e.g. Product, Category, Invoice): ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		fmt.Println("❌ Module name cannot be empty!")
		os.Exit(1)
	}

	// 2. Module Type
	fmt.Print("👉 Choose Module Type [1] Standard CRUD, [2] Custom Business Logic (default 1): ")
	typeChoice, _ := reader.ReadString('\n')
	typeChoice = strings.TrimSpace(typeChoice)
	moduleType := "crud"
	if typeChoice == "2" {
		moduleType = "business"
	}

	// 3. Transactions
	fmt.Print("👉 Enable Transaction Manager support? (y/n, default y): ")
	txChoice, _ := reader.ReadString('\n')
	txChoice = strings.TrimSpace(strings.ToLower(txChoice))
	isTx := txChoice != "n"

	// 4. Fields (if CRUD)
	var fields []string
	if moduleType == "crud" {
		fmt.Println("\n📝 Add Domain Entity Fields (press Enter with empty name when finished):")
		for {
			fmt.Print("   Field Name (e.g. title, price, stock, category_id): ")
			fieldName, _ := reader.ReadString('\n')
			fieldName = strings.TrimSpace(fieldName)
			if fieldName == "" {
				break
			}

			fmt.Print("   Field Go Type (string, int, float64, bool, time.Time, default string): ")
			fieldGoType, _ := reader.ReadString('\n')
			fieldGoType = strings.TrimSpace(fieldGoType)
			if fieldGoType == "" {
				fieldGoType = "string"
			}

			fmt.Print("   Is Nullable? (y/n, default n): ")
			nullChoice, _ := reader.ReadString('\n')
			nullChoice = strings.TrimSpace(strings.ToLower(nullChoice))
			if nullChoice == "y" {
				fields = append(fields, fmt.Sprintf("%s:%s?", fieldName, fieldGoType))
			} else {
				fields = append(fields, fmt.Sprintf("%s:%s", fieldName, fieldGoType))
			}
		}
	}

	// Build YAML Manifest
	manifestPath := fmt.Sprintf("manifests/%s.yaml", strings.ToLower(name))
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("name: %s\n", name))
	sb.WriteString(fmt.Sprintf("type: %s\n", moduleType))
	sb.WriteString(fmt.Sprintf("transactions: %t\n", isTx))

	if len(fields) > 0 {
		sb.WriteString("fields:\n")
		for _, f := range fields {
			parts := strings.Split(f, ":")
			fName := parts[0]
			fType := parts[1]
			nullable := strings.HasSuffix(fType, "?")
			cleanType := strings.TrimSuffix(fType, "?")

			sb.WriteString(fmt.Sprintf("  - name: %s\n", fName))
			sb.WriteString(fmt.Sprintf("    type: %s\n", cleanType))
			if nullable {
				sb.WriteString("    nullable: true\n")
			}
		}
	}

	_ = os.MkdirAll("manifests", 0755)
	err := os.WriteFile(manifestPath, []byte(sb.String()), 0644)
	if err != nil {
		fmt.Printf("❌ Failed to write manifest %s: %v\n", manifestPath, err)
		os.Exit(1)
	}

	fmt.Printf("\n✨ Manifest successfully created at '%s'\n", manifestPath)
	fmt.Println("==========================================================================")
	return name, manifestPath
}
