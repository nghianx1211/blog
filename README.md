# Blog API - Hiệu năng cao với Go

API Blog được xây dựng bằng Go với PostgreSQL, Redis, và Elasticsearch để đảm bảo hiệu năng cao và khả năng mở rộng.

## Tính năng

- **PostgreSQL**: Lưu trữ dữ liệu bài viết với GIN index để tìm kiếm nhanh theo tag
- **Redis**: Cache dữ liệu để giảm tải database và tăng tốc độ phản hồi  
- **Elasticsearch**: Tìm kiếm full-text mạnh mẽ
- **Transaction**: Đảm bảo tính nhất quán dữ liệu
- **Activity Logging**: Ghi log mọi hoạt động của bài viết

## API Endpoints

### Posts
- `POST /api/v1/posts` - Tạo bài viết mới
- `GET /api/v1/posts/:id` - Lấy bài viết theo ID
- `PUT /api/v1/posts/:id` - Cập nhật bài viết
- `DELETE /api/v1/posts/:id` - Xóa bài viết

### Search
- `GET /api/v1/posts/search?q=<query>&tags=<tags>&limit=<limit>&page=<page>` - Tìm kiếm full-text
- `GET /api/v1/posts/search-by-tag?tag=<tag_name>` - Tìm kiếm theo tag

## Cài đặt và chạy

### Sử dụng Docker Compose

```bash
# Clone repository
git clone <repository-url>
cd blog

go mod init blog
go mod tidy

# Chạy tất cả services
docker compose up -d

# API sẽ có sẵn tại http://localhost:8080
```

### Chạy thủ công

1. **Cài đặt dependencies:**
```bash
go mod init blog 
go mod tidy
```

2. **Setup databases:**
```bash
# PostgreSQL
createdb blog_db

# Chạy migrations
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/blog_db?sslmode=disable" up

# Redis (port 6379)
redis-server

# Elasticsearch (port 9200)
# Tải và chạy Elasticsearch
```

3. **Cài đặt biến môi trường:**
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=blog_db
export REDIS_HOST=localhost
export REDIS_PORT=6379
export ELASTICSEARCH_URL=http://localhost:9200
export SERVER_PORT=8080
```

4. **Chạy ứng dụng:**
```bash
go run cmd/server/main.go
```

## Ví dụ sử dụng

### Tạo bài viết mới
```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Bài viết đầu tiên",
    "content": "Đây là nội dung của bài viết đầu tiên",
    "tags": ["golang", "api", "blog"]
  }'
```

### Tìm kiếm bài viết
```bash
curl "http://localhost:8080/api/v1/posts/search?q=golang&limit=10&page=1"
```

### Tìm kiếm theo tag
```bash
curl "http://localhost:8080/api/v1/posts/search-by-tag?tag=golang"
```

## Cấu trúc dự án

```
blog/
├── cmd/server/          # Entry point
├── internal/
│   ├── config/          # Configuration
│   ├── database/        # Database connections
│   ├── models/          # Data models
│   ├── handlers/        # HTTP handlers
│   ├── services/        # Business logic
│   ├── middleware/      # HTTP middleware
│   └── utils/          # Utilities
├── migrations/          # Database migrations
├── docker-compose.yml   # Docker setup
└── README.md
```

## Tối ưu hóa hiệu năng

1. **Cache-Aside Pattern**: Sử dụng Redis để cache bài viết được truy cập thường xuyên
2. **GIN Index**: PostgreSQL GIN index cho tìm kiếm tag nhanh chóng
3. **Connection Pooling**: Tối ưu kết nối database
4. **Elasticsearch**: Tìm kiếm full-text hiệu suất cao
5. **Graceful Shutdown**: Đảm bảo tắt ứng dụng an toàn

## Monitoring & Health Check

- Health check endpoint: `GET /health`
- Logging middleware cho tất cả requests
- Recovery middleware để xử lý panic

## License

MIT License