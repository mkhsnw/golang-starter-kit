.PHONY: dev build test lint clean init-dx

# Menjalankan aplikasi dengan hot-reload (pastikan Air terinstal)
dev:
	air

# Membangun aplikasi menjadi binary
build:
	go build -o tmp/main.exe cmd/main.go

# Menjalankan unit tests
test:
	go test -v ./...

# Menjalankan linter (pastikan golangci-lint terinstal)
lint:
	golangci-lint run

# Menghapus folder sementara
clean:
	rm -rf tmp
	rm -f main.exe

# Command untuk menginstall dependency DX jika belum ada
init:
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
