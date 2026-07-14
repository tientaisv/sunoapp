# Stage 1: Biên dịch ứng dụng Go
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Sao chép các file cấu hình và mã nguồn
COPY go.mod ./
COPY main.go ./
COPY public/ ./public/
COPY internal/ ./internal/

# Biên dịch binary dạng tĩnh, tối ưu kích thước (-w -s để xóa debug info)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o suno-composer .

# Stage 2: Image chạy thực tế siêu nhẹ
FROM alpine:latest

# Cài đặt chứng chỉ SSL CA Certificates để gọi API HTTPS (Gemini API) thành công
RUN apk --no-cache add ca-certificates rclone

WORKDIR /root/

# Sao chép file chạy tĩnh từ builder stage
COPY --from=builder /app/suno-composer .

# Expose cổng kết nối mặc định của ứng dụng
EXPOSE 8080

# Chạy ứng dụng
CMD ["./suno-composer"]
