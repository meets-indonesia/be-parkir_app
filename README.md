# Parking Digital API - Palembang

A comprehensive REST API backend for the Palembang Digital Parking Application built with Go, featuring role-based access control for Customers, Jukirs (parking attendants), and Admins.

## 🚀 Features

- **Multi-role Authentication**: Customer, Jukir, and Admin roles with JWT-based authentication
- **QR Code Integration**: Check-in/check-out system using QR codes
- **GPS Verification**: Location validation within 50m radius of parking areas
- **Real-time Payment Processing**: Cash payment confirmation by Jukirs
- **Comprehensive Dashboard**: Statistics and reporting for all user types
- **RESTful API**: Well-documented endpoints with Swagger integration
- **Docker Support**: Complete containerization with multi-stage builds
- **Clean Architecture**: Domain-driven design with separation of concerns

## 🏗️ Architecture

```
be-parkir/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── domain/entities/ # Domain models and DTOs
│   ├── repository/      # Data access layer
│   ├── usecase/         # Business logic layer
│   └── delivery/http/   # HTTP handlers and middleware
├── migrations/          # Database migration files
├── docs/               # Swagger documentation
├── Dockerfile          # Multi-stage Docker build
├── docker-compose.yml  # Development environment
└── docker-compose.prod.yml # Production environment
```

## 🛠️ Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin HTTP framework
- **Database**: PostgreSQL with GORM ORM
- **Cache**: Redis for session management
- **Authentication**: JWT tokens
- **Validation**: go-playground/validator
- **Configuration**: Viper
- **Logging**: Logrus
- **Testing**: Testify
- **Documentation**: Swagger with swaggo
- **Containerization**: Docker & Docker Compose

## 📋 Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 15+ (if running locally)
- Redis 7+ (if running locally)

## 🚀 Quick Start

### Using Docker Compose (Recommended)

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd be-parkir
   ```

2. **Copy environment file**

   ```bash
   cp env.example .env
   ```

3. **Update environment variables**
   Edit `.env` file with your configuration:

   ```env
   JWT_SECRET=your-super-secret-jwt-key
   DB_PASSWORD=your-secure-password
   REDIS_PASSWORD=your-redis-password
   ```

4. **Start services**

   ```bash
   # Development
   make docker-up

   # Production
   make docker-up-prod
   ```

5. **Access the application**
   - API: http://localhost:8080
   - Swagger Docs: http://localhost:8080/swagger/index.html
   - Health Check: http://localhost:8080/health

### Manual Setup

1. **Install dependencies**

   ```bash
   make deps
   ```

2. **Set up environment**

   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

3. **Start PostgreSQL and Redis**

   ```bash
   # Using Docker
   docker run -d --name postgres -e POSTGRES_DB=parking_app -e POSTGRES_USER=parking_user -e POSTGRES_PASSWORD=parking_pass -p 5432:5432 postgres:15-alpine
   docker run -d --name redis -p 6379:6379 redis:7-alpine
   ```

4. **Run the application**
   ```bash
   make run
   ```

## 📚 API Documentation

### Authentication Endpoints

| Method | Endpoint                | Description       | Auth Required |
| ------ | ----------------------- | ----------------- | ------------- |
| POST   | `/api/v1/auth/register` | Register new user | No            |
| POST   | `/api/v1/auth/login`    | User login        | No            |
| POST   | `/api/v1/auth/refresh`  | Refresh JWT token | No            |
| POST   | `/api/v1/auth/logout`   | User logout       | Yes           |

### Customer Endpoints

| Method | Endpoint                    | Description           | Auth Required  |
| ------ | --------------------------- | --------------------- | -------------- |
| GET    | `/api/v1/profile`           | Get user profile      | Yes (Customer) |
| PUT    | `/api/v1/profile`           | Update profile        | Yes (Customer) |
| GET    | `/api/v1/parking/locations` | Get nearby areas      | Yes (Customer) |
| POST   | `/api/v1/parking/checkin`   | Start parking session | Yes (Customer) |
| POST   | `/api/v1/parking/checkout`  | End parking session   | Yes (Customer) |
| GET    | `/api/v1/parking/active`    | Get active session    | Yes (Customer) |
| GET    | `/api/v1/parking/history`   | Get parking history   | Yes (Customer) |

### Jukir Endpoints

| Method | Endpoint                         | Description          | Auth Required |
| ------ | -------------------------------- | -------------------- | ------------- |
| GET    | `/api/v1/jukir/dashboard`        | Get dashboard stats  | Yes (Jukir)   |
| GET    | `/api/v1/jukir/pending-payments` | Get pending payments | Yes (Jukir)   |
| GET    | `/api/v1/jukir/active-sessions`  | Get active sessions  | Yes (Jukir)   |
| POST   | `/api/v1/jukir/confirm-payment`  | Confirm cash payment | Yes (Jukir)   |
| GET    | `/api/v1/jukir/qr-code`          | Get QR code info     | Yes (Jukir)   |
| GET    | `/api/v1/jukir/daily-report`     | Get daily report     | Yes (Jukir)   |

### Admin Endpoints

| Method | Endpoint                            | Description         | Auth Required |
| ------ | ----------------------------------- | ------------------- | ------------- |
| GET    | `/api/v1/admin/overview`            | System overview     | Yes (Admin)   |
| GET    | `/api/v1/admin/jukirs`              | List all jukirs     | Yes (Admin)   |
| POST   | `/api/v1/admin/jukirs`              | Create jukir        | Yes (Admin)   |
| PUT    | `/api/v1/admin/jukirs/{id}/approve` | Approve jukir       | Yes (Admin)   |
| GET    | `/api/v1/admin/reports`             | Generate reports    | Yes (Admin)   |
| GET    | `/api/v1/admin/sessions`            | All sessions        | Yes (Admin)   |
| POST   | `/api/v1/admin/areas`               | Create parking area | Yes (Admin)   |
| PUT    | `/api/v1/admin/areas/{id}`          | Update parking area | Yes (Admin)   |

## 🔧 Configuration

### Environment Variables

| Variable             | Description          | Default      | Required |
| -------------------- | -------------------- | ------------ | -------- |
| `DB_HOST`            | Database host        | localhost    | Yes      |
| `DB_PORT`            | Database port        | 5432         | Yes      |
| `DB_USER`            | Database user        | parking_user | Yes      |
| `DB_PASSWORD`        | Database password    | -            | Yes      |
| `DB_NAME`            | Database name        | parking_app  | Yes      |
| `DB_SSLMODE`         | SSL mode             | disable      | No       |
| `REDIS_HOST`         | Redis host           | localhost    | Yes      |
| `REDIS_PORT`         | Redis port           | 6379         | Yes      |
| `REDIS_PASSWORD`     | Redis password       | -            | No       |
| `REDIS_DB`           | Redis database       | 0            | No       |
| `JWT_SECRET`         | JWT secret key       | -            | Yes      |
| `JWT_ACCESS_EXPIRY`  | Access token expiry  | 15m          | No       |
| `JWT_REFRESH_EXPIRY` | Refresh token expiry | 7d           | No       |
| `SERVER_PORT`        | Server port          | 8080         | No       |
| `SERVER_ENVIRONMENT` | Environment          | development  | No       |

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -cover ./...

# Run specific test package
go test -v ./internal/usecase
```

## 🐳 Docker Commands

```bash
# Build Docker image
make docker-build

# Start development environment
make docker-up

# Start production environment
make docker-up-prod

# Stop services
make docker-down

# View logs
make docker-logs

# View production logs
make docker-logs-prod
```

## 📊 Database Schema

### Users Table

- `id` (Primary Key)
- `name` (VARCHAR)
- `email` (VARCHAR, Unique)
- `phone` (VARCHAR)
- `password` (VARCHAR, Hashed)
- `role` (ENUM: customer, jukir, admin)
- `status` (ENUM: active, inactive, pending)
- `created_at`, `updated_at`, `deleted_at`

### Jukirs Table

- `id` (Primary Key)
- `user_id` (Foreign Key to Users)
- `jukir_code` (VARCHAR, Unique)
- `area_id` (Foreign Key to Parking Areas)
- `qr_token` (VARCHAR, Unique)
- `status` (ENUM: active, inactive, pending)
- `created_at`, `updated_at`, `deleted_at`

### Parking Areas Table

- `id` (Primary Key)
- `name` (VARCHAR)
- `address` (VARCHAR)
- `latitude` (DECIMAL)
- `longitude` (DECIMAL)
- `hourly_rate` (DECIMAL)
- `status` (ENUM: active, inactive, maintenance)
- `created_at`, `updated_at`, `deleted_at`

### Parking Sessions Table

- `id` (Primary Key)
- `user_id` (Foreign Key to Users)
- `jukir_id` (Foreign Key to Jukirs)
- `area_id` (Foreign Key to Parking Areas)
- `checkin_time` (TIMESTAMP)
- `checkout_time` (TIMESTAMP, Nullable)
- `duration` (INTEGER, in minutes)
- `total_cost` (DECIMAL)
- `payment_status` (ENUM: pending, paid, failed)
- `session_status` (ENUM: active, pending_payment, completed, cancelled, timeout)
- `created_at`, `updated_at`, `deleted_at`

### Payments Table

- `id` (Primary Key)
- `session_id` (Foreign Key to Parking Sessions)
- `amount` (DECIMAL)
- `payment_method` (ENUM: cash, qris, bank_transfer)
- `confirmed_by` (Foreign Key to Jukirs)
- `confirmed_at` (TIMESTAMP, Nullable)
- `status` (ENUM: pending, paid, failed, refunded)
- `created_at`, `updated_at`, `deleted_at`

## 🔒 Security Features

- **JWT Authentication**: Secure token-based authentication
- **Role-based Access Control**: Granular permissions for different user types
- **Password Hashing**: Bcrypt for secure password storage
- **Input Validation**: Comprehensive request validation
- **Rate Limiting**: Protection against spam and abuse
- **CORS Support**: Configurable cross-origin resource sharing
- **SQL Injection Protection**: GORM ORM with parameterized queries

## 🚀 Deployment

### Production Deployment

1. **Set up production environment**

   ```bash
   cp env.example .env.prod
   # Configure production values
   ```

2. **Deploy with Docker Compose**

   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

3. **Set up reverse proxy (Nginx)**

   ```nginx
   server {
       listen 80;
       server_name your-domain.com;

       location / {
           proxy_pass http://localhost:8080;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
       }
   }
   ```

4. **Set up SSL (Let's Encrypt)**

   ```bash
   # Install certbot
   sudo apt install certbot python3-certbot-nginx

   # Get SSL certificate
   sudo certbot --nginx -d your-domain.com
   ```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:

- Create an issue in the repository
- Contact the development team
- Check the [API documentation](http://localhost:8080/swagger/index.html)

## 🗺️ Roadmap

- [ ] Mobile app integration
- [ ] Real-time notifications
- [ ] Payment gateway integration
- [ ] Advanced analytics dashboard
- [ ] Multi-language support
- [ ] API rate limiting
- [ ] Automated testing pipeline
- [ ] Performance monitoring

---

**Built with ❤️ for Palembang Digital Parking System**
