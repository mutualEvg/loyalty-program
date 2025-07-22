# Gufermart Loyalty System

A hobby loyalty system HTTP API that manages user registration, order processing, and loyalty points accrual/withdrawal.

## Features

- **User Management**: Registration, authentication, and authorization
- **Order Processing**: Accept and validate order numbers using Luhn algorithm
- **Loyalty Points**: Automatic accrual from processed orders
- **Points Withdrawal**: Use loyalty points for partial/full payment of orders
- **Transaction History**: Track all orders and withdrawals
- **Security**: JWT-based authentication, password hashing
- **Database**: PostgreSQL with automatic schema migration

## API Endpoints

### Authentication
- `POST /api/user/register` - User registration
- `POST /api/user/login` - User authentication

### Order Management
- `POST /api/user/orders` - Submit order number for processing
- `GET /api/user/orders` - Get user's orders with status and accrual info

### Balance & Withdrawals
- `GET /api/user/balance` - Get current balance and total withdrawn
- `POST /api/user/balance/withdraw` - Withdraw points for order payment
- `GET /api/user/withdrawals` - Get withdrawal history

## Order Processing Statuses

- `NEW` - Order uploaded but not yet processed
- `PROCESSING` - Order being processed by loyalty calculation system
- `INVALID` - Order rejected by loyalty calculation system
- `PROCESSED` - Order processed successfully, points accrued if applicable

## Installation & Setup

### Prerequisites

- Go 1.24+
- PostgreSQL 12+

### 1. Install Dependencies

```bash
go mod tidy
```

### 2. Database Setup

Create a PostgreSQL database:

```sql
CREATE DATABASE gofemart;
CREATE USER gofemart_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE gofemart TO gofemart_user;
```

### 3. Configuration

Copy the example configuration:

```bash
cp config.example .env
```

Edit `.env` with your database credentials and settings:

```env
DATABASE_URL=postgres://gofemart_user:your_password@localhost:5432/gofemart?sslmode=disable
SERVER_ADDRESS=:8080
JWT_SECRET=your-very-secure-secret-key-change-this-in-production
```

### 4. Run the Application

```bash
go run main.go
```

The server will start on `http://localhost:8080`

## Usage Examples

### 1. Register a User

```bash
curl -X POST http://localhost:8080/api/user/register \
  -H "Content-Type: application/json" \
  -d '{"login": "user123", "password": "securepassword"}'
```

### 2. Login

```bash
curl -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"login": "user123", "password": "securepassword"}'
```

### 3. Submit an Order (with authentication)

```bash
curl -X POST http://localhost:8080/api/user/orders \
  -H "Content-Type: text/plain" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d "12345678903"
```

### 4. Check Balance

```bash
curl -X GET http://localhost:8080/api/user/balance \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 5. Withdraw Points

```bash
curl -X POST http://localhost:8080/api/user/balance/withdraw \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"order": "ORDER123", "sum": 100.50}'
```

## Project Structure

```
gofemart/
├── cmd/server/           # Alternative server entry point
├── internal/
│   ├── database/         # Database connection and schema
│   ├── handlers/         # HTTP request handlers
│   ├── middleware/       # Authentication middleware
│   ├── models/          # Data models and structs
│   ├── services/        # Business logic layer
│   └── utils/           # Utility functions (Luhn, JWT, etc.)
├── main.go              # Main application entry point
├── go.mod               # Go module definition
├── config.example       # Example configuration file
└── README.md           # This file
```

## Key Features Explained

### Luhn Algorithm Validation

Order numbers are validated using the Luhn algorithm to ensure they follow the correct format before processing.

### Loyalty Points Simulation

The system includes a mock loyalty calculation service that:
- Orders ending in 0-3: Award 10% of order value as points
- Orders ending in 4-6: Valid orders but no points awarded
- Orders ending in 7-9: Invalid orders

### Security

- Passwords are hashed using bcrypt
- JWT tokens for stateless authentication
- Protected routes require authentication
- SQL injection prevention with parameterized queries

### Database Schema

The system automatically creates the following tables:
- `users` - User accounts and authentication
- `orders` - Order tracking and processing status
- `user_balances` - Current and lifetime balance tracking
- `withdrawals` - Points withdrawal transaction history

## Response Codes

- `200` - Success
- `202` - Accepted (order submitted for processing)
- `204` - No Content (empty result sets)
- `400` - Bad Request (invalid format)
- `401` - Unauthorized (authentication required)
- `409` - Conflict (login taken, order already submitted by another user)
- `422` - Unprocessable Entity (invalid order number format)
- `500` - Internal Server Error

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o gofemart main.go
```

## Production Deployment

1. Change JWT_SECRET to a secure random value
2. Use a production PostgreSQL database
3. Enable HTTPS/TLS
4. Configure proper logging
5. Set up monitoring and health checks
6. Use environment variables for sensitive configuration

## License

This project is for educational purposes as part of a technical assignment. 