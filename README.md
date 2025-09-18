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

2. **Chạy ứng dụng:**
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

### Lấy bài viết theo ID
```bash
curl -X GET http://localhost:8080/api/v1/posts/<post_id>
```

### Cập nhật bài viết
```bash
curl -X PUT http://localhost:8080/api/v1/posts/<post_id> \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Tiêu đề mới",
    "content": "Nội dung mới"
  }'
```

### Xóa bài viết
```bash
curl -X DELETE http://localhost:8080/api/v1/posts/<post_id>
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
├── docker-compose.yml   # Docker setup
└── README.md
```

## Tối ưu hóa hiệu năng

1. **Cache-Aside Pattern**: Sử dụng Redis để cache bài viết được truy cập thường xuyên
2. **GIN Index**: PostgreSQL GIN index cho tìm kiếm tag nhanh chóng
3. **Connection Pooling**: Tối ưu kết nối database
4. **Elasticsearch**: Tìm kiếm full-text hiệu suất cao
5. **Graceful Shutdown**: Đảm bảo tắt ứng dụng an toàn