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

RelGo — Framework Go berbasis rel opini yang menjaga arsitektur backend
Anda tetap lurus di jalurnya tanpa anjlok (architectural erosion).

USAGE:
  rel <command> [arguments]

AVAILABLE COMMANDS:
  new <project-path>     Scaffold a new RelGo project
  gen [module-name]      Generate a CRUD or business module (or interactive wizard)
  rm <module-name>       Remove a generated module cleanly
  doctor                 Run dev environment diagnostic health check
  lint                   Run Architecture Linter (AST parser rule check)
  version                Show RelGo version information

EXAMPLES:
  rel new github.com/username/my-awesome-api
  rel gen Product
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
	case "doctor":
		runSubcommand("cmd/doctor/main.go", args...)
	case "lint":
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
