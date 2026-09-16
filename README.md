# 🔗 Укротитель ссылок 
Pet project to switch from java to go with production-ready architecture

## URL shortner

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev)
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen)](./)
--------------
+ shortens entered URL using murmur3 hash
+ input - long URL, output - short link with alias. Stored in PostgreSQL. Short link redirects to original URL
+ The goal is a **production-ready** application with clean architecture, tests, and proper error handling.

### API

### Create a short link

```http
POST http://localhost:8080/api/v1/shorten
Content-Type: application/json

{
  "link": "https://www.google.com/search?q=golang"
}
```

**Response:**

```json
{
  "shortLink": "http://localhost:8080/3XqGtZ"
}
```

### Follow a short link

```http
GET http://localhost:8080/{alias}
```

**Response:**

```http
HTTP/1.1 301 Moved Permanently
Location: https://www.google.com/search?q=golang
```



### Tech Stack


| Category | Technology                                                    |
|----------|---------------------------------------------------------------|
| **Language** | Go 1.26                                                       |
| **HTTP Router** | [`go-chi/chi`](https://github.com/go-chi/chi)                 |
| **Database** | PostgreSQL 16                                                 |
| **DB Driver** | [`jackc/pgx`](https://github.com/jackc/pgx)                   |
| **Migrations** | [`golang-migrate`](https://github.com/golang-migrate/migrate) |
| **Logging** | `log/slog` (standard library)                                 |
| **Testing** | `testify`, `gomock`, `testcontainers-go`                      |
| **Configuration** | `godotenv` + environment variables                            |

### Roadmap

- [x] Core functionality (create and redirect)
- [x] Graceful shutdown
- [x] Structured logging
- [x] Database migrations
- [x] Unit and integration tests
- [ ] Click statistics
- [ ] Custom aliases
- [ ] TTL for links
- [ ] Rate limiting
- [ ] Redis cache
- [ ] CI/CD with GitHub Actions

---