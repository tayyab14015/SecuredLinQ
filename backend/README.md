# SecuredLinQ Backend

Go backend for the SecuredLinQ video call and load management application.

## Architecture

This backend follows clean architecture principles with the following layers:

- **Handlers** (`internal/handler/`): HTTP request handlers
- **Services** (`internal/service/`): Business logic
- **Repositories** (`internal/repository/`): Data access layer
- **Models** (`internal/models/`): GORM models

## Features

- RESTful API with Gin framework
- Cookie-based session authentication
- GORM for database access with MySQL
- AWS S3 integration for media storage
- Agora integration for video calls and recording
- Email notifications for meeting links

## Project Structure

```
backend/
├── cmd/
│   └── api/           # Main API server
├── internal/
│   ├── config/        # Configuration
│   ├── database/      # Database connection
│   ├── handler/       # HTTP handlers
│   ├── middleware/    # HTTP middleware
│   ├── models/        # GORM models
│   ├── repository/    # Repository layer
│   └── service/       # Service layer
├── pkg/
│   ├── agora/         # Agora SDK integration
│   └── s3/            # AWS S3 integration
├── go.mod
└── go.sum
```

## Getting Started

### Prerequisites

- Go 1.21+
- MySQL
- AWS S3 bucket
- Agora account
- Email service (SMTP)

### Configuration

1. Copy the example environment file:
   ```bash
   cp env.example.txt .env
   ```

2. Update the `.env` file with your configuration.

### Running the Server

```bash
# Run the API server
go run cmd/api/main.go
```

### Building

```bash
# Build the API server
go build -o bin/api cmd/api/main.go
```

## API Endpoints

### Authentication
- `POST /api/auth/login` - Admin login
- `POST /api/auth/logout` - Logout
- `GET /api/auth/validate` - Validate session
- `POST /api/auth/driver/register` - Driver registration
- `POST /api/auth/driver/login` - Driver login

### Agora (Video Calls)
- `POST /api/agora/token` - Generate RTC token
- `POST /api/agora/recording/start` - Start recording
- `POST /api/agora/recording/stop` - Stop recording
- `GET /api/agora/recording/query` - Query recording status

### Meetings
- `POST /api/meetings` - Create/get meeting room
- `GET /api/meetings` - Get meeting by room ID
- `DELETE /api/meetings` - End meeting

### Admin Routes
- `GET /api/admin/loads` - Get all loads
- `POST /api/admin/loads` - Create new load
- `GET /api/admin/loads/:id` - Get load by ID
- `GET /api/admin/loads/by-status` - Get loads by status
- `POST /api/admin/loads/:id/assign` - Assign driver to load
- `POST /api/admin/loads/:id/start-meeting` - Start meeting for load
- `DELETE /api/admin/loads/:id` - Delete load

### Driver Management (Admin)
- `GET /api/admin/drivers` - Get all drivers
- `GET /api/admin/drivers/:id` - Get driver by ID

### Driver Routes
- `GET /api/driver/loads` - Get driver's assigned loads
- `GET /api/driver/loads/:id` - Get load details
- `POST /api/driver/loads/:id/complete` - Mark load as completed
- `PUT /api/driver/loads/:id/status` - Update load status

### Email (Admin)
- `POST /api/email/send-meeting-link` - Send meeting link via email

### Media (Admin)
- `GET /api/media` - Get load media
- `POST /api/media/screenshot` - Save screenshot
- `GET /api/media/screenshots` - Get screenshots by load

