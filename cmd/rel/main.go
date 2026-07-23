package main

import (
	"fmt"
	"os"
	"os/exec"
)

func printHelp() {
	fmt.Print(`
╔══════════════════════════════════════════════════════════════╗
║              🚂 RelGo Framework CLI (rel)                     ║
╚══════════════════════════════════════════════════════════════╝

RelGo — Starter-kit Go yang buat kamu membuat API dengan lancar dan mulu seperti kamu berjalan di rel.

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
		runSubcommand("cmd/new/main.go", args...)
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
		fmt.Println("🚂 RelGo Framework v1.0.0 (Core Frozen & Production Ready)")
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("❌ Unknown command 'rel %s'. Run 'rel help' for usage.\n", cmd)
		os.Exit(1)
	}
}

func runSubcommand(scriptPath string, args ...string) {
	if _, err := os.Stat(scriptPath); err != nil {
		fmt.Printf("⚠️  RelGo Notice: File '%s' not found in current directory.\n", scriptPath)
		fmt.Printf("   Please make sure you are executing 'rel' inside a RelGo project directory.\n")
		os.Exit(1)
	}

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
}
