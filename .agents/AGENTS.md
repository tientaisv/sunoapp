# Project Rules

## Build & Execution Rules
- **Biên dịch & Chạy dự án bằng Docker**: Ứng dụng được thiết lập chạy trong Docker container (`golang:1.22-alpine`). Khi cần kiểm tra build/biên dịch mã Go, BẮT BUỘC sử dụng `docker build -t suno-composer:test .` hoặc `docker-compose build`. Tuyệt đối KHÔNG chạy `go build` trực tiếp ngoài host machine.
