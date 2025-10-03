# Greenlight API

Greenlight is a RESTful API server for managing movies and users, built with Go and PostgreSQL.

## Features

- **Movies CRUD**: Create, read, update, and delete movies.
- **User Registration**: Register new users with email verification.
- **Authentication**: Secure token-based authentication for users.
- **Token Management**: Issue, store, and validate authentication tokens.
- **Role-Based Access Control**: Assign and check permissions for users.
- **Permissions Management**: Add and retrieve permissions for users.
- **Email Sending**: Welcome emails sent via SMTP.
- **Rate Limiting**: Per-IP rate limiting for API endpoints.
- **Validation**: Input validation for movies and users.
- **Pagination & Filtering**: List movies with pagination, sorting, and filtering.
- **Database Migrations**: SQL migration scripts for schema management.

## Tech Stack

- **Go** (1.24+)
- **PostgreSQL**
- **SMTP** (for email)
- **Libraries**:
  - [github.com/julienschmidt/httprouter](https://github.com/julienschmidt/httprouter) (routing)
  - [github.com/lib/pq](https://github.com/lib/pq) (PostgreSQL driver)
  - [github.com/wneessen/go-mail](https://github.com/wneessen/go-mail) (SMTP client)
  - [github.com/tomasen/realip](https://github.com/tomasen/realip) (real IP extraction)
  - [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) (password hashing)
  - [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) (rate limiting)

## Getting Started

### Prerequisites

- Go 1.24+
- PostgreSQL database

### Setup

1. **Clone the repository:**

   ```sh
   git clone https://github.com/solomonsitotaw23/greenlight.git
   cd greenlight
   ```

2. **Configure the database:**

   - Create a PostgreSQL database.
   - Set the `GREENLIGHT_DB_DSN` environment variable:
     ```
     export GREENLIGHT_DB_DSN="postgres://user:password@localhost:5432/greenlight?sslmode=disable"
     ```

3. **Run migrations:**

   - Use a migration tool (e.g., [golang-migrate](https://github.com/golang-migrate/migrate)) to apply migrations in the `migrations/` folder:
     ```sh
     migrate -path migrations -database "$GREENLIGHT_DB_DSN" up
     ```

4. **Build and run the server:**
   ```sh
   go build -o bin/greenlight ./cmd/api
   ./bin/greenlight -port 4000
   ```

### Configuration

You can configure the server using command-line flags or environment variables. See [`cmd/api/main.go`](cmd/api/main.go) for all options.

### API Endpoints

- `GET /v1/healthcheck` – Health check
- `GET /v1/movies` – List movies
- `POST /v1/movies` – Create movie
- `GET /v1/movies/:id` – Get movie details
- `PATCH /v1/movies/:id` – Update movie
- `DELETE /v1/movies/:id` – Delete movie
- `POST /v1/users` – Register user
- `POST /v1/tokens/authentication` – Obtain authentication token
- **Permissions Endpoints** (example):
  - `GET /v1/users/:id/permissions` – Get all permissions for a user
  - `POST /v1/users/:id/permissions` – Add permissions to a user

## License

MIT
