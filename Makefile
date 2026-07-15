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
	go install github.com/vektra/mockery/v2@v2.42.1

# Generate mocks
mock:
	mockery --dir internal/repository --all --output internal/repository/mocks
	mockery --dir internal/usecase --all --output internal/usecase/mocks

# Generate swagger docs
docs:
	swag init -g cmd/main.go --parseInternal --parseDependency
