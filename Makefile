.PHONY: help dev build run migrate-up migrate-down migrate-version migrate-reset seed test fmt tidy check

help: ## Liệt kê các lệnh có sẵn
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

dev: ## Chạy server ở chế độ phát triển
	go run ./cmd/server

build: ## Biên dịch binary vào ./bin/ai-staffing
	go build -o bin/ai-staffing ./cmd/server

migrate-up: ## Áp dụng toàn bộ migration còn thiếu
	go run ./cmd/migrate up

migrate-down: ## Lùi lại một bước migration
	go run ./cmd/migrate down

migrate-version: ## Xem phiên bản schema hiện tại
	go run ./cmd/migrate version

migrate-reset: ## Xóa sạch schema rồi tạo lại (chỉ dùng khi phát triển)
	go run ./cmd/migrate drop && go run ./cmd/migrate up

seed: ## Nạp dữ liệu mẫu
	go run ./cmd/seed

test: ## Chạy toàn bộ test
	go test ./... -count=1

fmt: ## Định dạng mã nguồn
	go fmt ./...

tidy: ## Dọn dependency
	go mod tidy

check: fmt tidy ## Kiểm tra trước khi commit
	go vet ./...
	go build ./...
	go test ./... -count=1
