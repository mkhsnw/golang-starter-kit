package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type EnvConfigCheck struct {
	App struct {
		Name string `json:"name"`
		Port int    `json:"port"`
	} `json:"app"`
	Database struct {
		Host string `json:"host"`
		Port int    `json:"port"`
		Name string `json:"name"`
	} `json:"database"`
	JWT struct {
		Secret string `json:"secret"`
	} `json:"jwt"`
	Redis struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"redis"`
}

func main() {
	fmt.Println("🩺 Running Starter Kit Doctor (gokit doctor)...")
	fmt.Println("==========================================================================")

	passedCount := 0
	failedCount := 0

	// 1. Check Go Version
	fmt.Printf("[1/6] Checking Go Compiler Environment...\n")
	goVer := runtime.Version()
	fmt.Printf("      ✔ Installed Go Version: %s (%s/%s)\n", goVer, runtime.GOOS, runtime.GOARCH)
	passedCount++

	// 2. Check DX Tool Dependencies
	fmt.Printf("\n[2/6] Checking Developer Experience (DX) Command Tools...\n")
	tools := []string{"air", "golangci-lint", "swag", "gotestsum"}
	for _, tool := range tools {
		if checkCommandExists(tool) {
			fmt.Printf("      ✔ Tool '%s': Available in PATH\n", tool)
			passedCount++
		} else {
			fmt.Printf("      ⚠️ Tool '%s': Not found in PATH (Run 'task init' to install)\n", tool)
			failedCount++
		}
	}

	// 3. Check Configuration (env.json)
	fmt.Printf("\n[3/6] Checking Environment Configuration (env.json)...\n")
	envPath := "env.json"
	envData, err := os.ReadFile(envPath)
	if err != nil {
		fmt.Printf("      ❌ File 'env.json' not found! (Copy env.example.json to env.json)\n")
		failedCount++
	} else {
		var cfg EnvConfigCheck
		if err := json.Unmarshal(envData, &cfg); err != nil {
			fmt.Printf("      ❌ Invalid JSON format in 'env.json': %v\n", err)
			failedCount++
		} else {
			fmt.Printf("      ✔ File 'env.json': Parsed successfully (App: %s, Port: %d)\n", cfg.App.Name, cfg.App.Port)
			passedCount++

			if cfg.JWT.Secret == "your-secret-key-here" || len(cfg.JWT.Secret) < 16 {
				fmt.Printf("      ⚠️ Security Warning: JWT Secret is default or under 16 characters!\n")
				failedCount++
			} else {
				fmt.Printf("      ✔ Security: JWT Secret configured safely\n")
				passedCount++
			}
		}
	}

	// 4. Check Database Connectivity (TCP Dial)
	fmt.Printf("\n[4/6] Checking Database Connection (MySQL)...\n")
	mysqlAddr := "127.0.0.1:3306"
	if conn, err := net.DialTimeout("tcp", mysqlAddr, 2*time.Second); err == nil {
		conn.Close()
		fmt.Printf("      ✔ MySQL Database: Host %s reachable\n", mysqlAddr)
		passedCount++
	} else {
		fmt.Printf("      ⚠️ MySQL Database: Host %s unreachable (%v). Check docker-compose or local MySQL server.\n", mysqlAddr, err)
		failedCount++
	}

	// 5. Check Redis Connectivity (TCP Dial)
	fmt.Printf("\n[5/6] Checking Cache Connection (Redis)...\n")
	redisAddr := "127.0.0.1:6379"
	if conn, err := net.DialTimeout("tcp", redisAddr, 2*time.Second); err == nil {
		conn.Close()
		fmt.Printf("      ✔ Redis Cache: Host %s reachable\n", redisAddr)
		passedCount++
	} else {
		fmt.Printf("      ⚠️ Redis Cache: Host %s unreachable (%v). (Optional, rate limiter falls back gracefully)\n", redisAddr, err)
	}

	// 6. Check Project Structure & Generator Templates
	fmt.Printf("\n[6/6] Checking Project Structure & Generator Integrity...\n")
	dirs := []string{"cmd/gen/templates", "manifests", "db/seed", "db/factory", "internal/foundation"}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err == nil {
			fmt.Printf("      ✔ Directory '%s': Intact\n", dir)
			passedCount++
		} else {
			fmt.Printf("      ❌ Directory '%s': Missing!\n", dir)
			failedCount++
		}
	}

	fmt.Println("\n==========================================================================")
	fmt.Printf("🩺 DOCTOR SUMMARY: %d Checks Passed, %d Warnings/Failures Detected.\n", passedCount, failedCount)
	if failedCount == 0 {
		fmt.Println("🚀 System is 100% HEALTHY and ready for development!")
	} else {
		fmt.Println("💡 Address any warnings above to ensure optimal development experience.")
	}
	fmt.Println("==========================================================================")
}

func checkCommandExists(cmd string) bool {
	if runtime.GOOS == "windows" {
		_, err := exec.LookPath(cmd + ".exe")
		if err == nil {
			return true
		}
	}
	_, err := exec.LookPath(cmd)
	return err == nil
}
