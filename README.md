# Go Banking Application

A banking transaction processing application built with Go for handling incoming requests from 3rd-party providers. This application processes balance updates and provides current balance information for users with built-in concurrency support.

## Architecture

- **Backend**: Go with Gorilla Mux router
- **Database**: PostgreSQL 16
- **Containerization**: Docker
- **Migration**: golang-migrate

## Features

- **Transaction Processing**: Handle `win`/`lose` transactions from 3rd-party providers
- **Balance Management**: Real-time balance calculation and retrieval
- **Idempotency**: Duplicate transaction prevention using `transactionId`
- **Source Type Support**: Handle requests from `game`, `server`, and `payment` sources
- **Concurrency Safe**: Process multiple transactions simultaneously
- **Negative Balance Protection**: Prevent account balance from going negative
- **Predefined Users**: Users with IDs 1, 2, and 3 ready for testing

## 📁 Project Structure

```
go-banking/
├── app/
│   ├── api/                 # API handlers
│   │   ├── accounts.go
│   │   ├── transactions.go
│   │   └── users.go
│   ├── database/            # Database layer
│   │   ├── migration/       # SQL migration files
│   │   ├── query/          # SQL queries
│   │   └── sqlc/           # Generated SQL code
│   ├── helpers/            # Helper functions
│   ├── middleware/         # HTTP middleware
│   └── models/             # Data models
├── docker-compose.yml      # Production docker setup
├── docker-compose.dev.yml  # Development docker setup
├── Dockerfile             # Container definition
├── Makefile              # Build and deployment commands
├── .env             # Environment configuration
├── migrate.sh            # Migration helper script
└── main.go              # Application entry point
```

## Prerequisites
- [Docker](https://www.docker.com/get-started) 
- [Docker Compose](https://docs.docker.com/compose/install/)

## Quick Start

```bash
git clone https://github.com/rathorevk/go-banking.git
cd go-banking
docker-compose up -d
```

The application will be available at `http://localhost:8000` with predefined users (ID: 1, 2, 3) ready for testing.

## 📊 Database Schema

The application uses PostgreSQL with the following schema:

### Users Table

| Column      | Type      | Description               |
|-------------|-----------|---------------------------|
| id          | BIGSERIAL | Primary key (user ID)     |
| username    | VARCHAR   | User name (UNIQUE)        |
| full_name   | VARCHAR   | User Full Name            |
| email       | VARCHAR   | User Email (UNIQUE)       |
| inserted_at | TIMESTAMP | User insertion time       |

### Accounts Table

| Column      | Type          | Description                        |
|-------------|---------------|------------------------------------|
| id          | BIGSERIAL     | Primary key (account ID)           |
| user_id     | INTEGER       | Foreign key to users table         |
| balance     | NUMERIC(10,2) | Current account balance            |
| currency    | VARCHAR       | DEFAULT 'EUR'(optional)            |
| status      | VARCHAR       | DEFAULT 'active'(optional)         |
| inserted_at | TIMESTAMP     | Account insertion time             |

### Transactions Table

| Column         | Type           | Description                          |
|----------------|----------------|--------------------------------------|
| id             | TEXT           | Primary key (transaction ID)         |
| account_id     | INTEGER        | Foreign key to accounts table        |
| type           | VARCHAR        | Transaction type: 'win' or 'lose'    |
| amount         | NUMERIC(10,2)  | Transaction amount                   |
| source         | VARCHAR        | Source: 'game', 'server', 'payment'  |
| inserted_at    | TIMESTAMP      | Transaction insertion time           |

**Relationships**:
- Each account is linked to a user (`accounts.user_id` → `users.id`)
- Each transaction is linked to an account (`transactions.user_id` → `accounts.id`)
- Idempotency is enforced via unique `transaction_id` in transactions

### Predefined Users
- User ID: 1 (Account ID: 1 and balance: 0.00)
- User ID: 2 (Account ID: 2 and balance: 0.00)
- User ID: 3 (Account ID: 3 and balance: 0.00)

### Migration
The application uses: [golang-migrate](https://github.com/golang-migrate/migrate)

Migration files are located in `app/database/migrations/` and follow the naming convention:
- `000001_add_users.up.sql` - Creates users table with predefined users
- `000002_add_accounts.up.sql` - Creates accounts table with initial balances
- `000003_add_transactions.up.sql` - Creates transactions table with idempotency support

### Write and Generate Queries

All SQL queries for the application are defined in `app/database/query/`.
To generate Go code for these queries, the project uses [sqlc](https://github.com/kyleconroy/sqlc) and
config: `sqlc.yaml`.

**Generate Go code from SQL queries:**
```bash
sqlc generate
```

This command reads the SQL files and produces type-safe Go code in `app/database/sqlc/` for database access.

## 🌐 API Endpoints

The application provides the exact endpoints required by the test task:

| Method | Endpoint | Description | Headers Required |
|--------|----------|-------------|------------------|
| POST | `/user/{userId}/transaction` | Process transaction (win/lose) | `Source-Type:`, `Content-Type: application/json` |
| GET | `/user/{userId}/balance` | Get current user balance | None |

### Transaction Endpoint

**Endpoint**: `POST /user/{userId}/transaction`

**Required Headers**:
- `Source-Type`: `game`, `server`, or `payment`
- `Content-Type`: `application/json`

**Request Body**:
```json
{
  "state": "win|lose",
  "amount": "10.15",
  "transactionId": "unique-transaction-id"
}
```

**Field Specifications**:
- `state`: String - either "win" (increases balance) or "lose" (decreases balance)
- `amount`: String - monetary amount with up to 2 decimal places
- `transactionId`: String - unique identifier for idempotency

**Example Requests**:

**Win Transaction (Increase Balance)**:
```bash
curl -X POST http://localhost:8000/user/1/transaction \
  -H "Source-Type: game" \
  -H "Content-Type: application/json" \
  -d '{"state": "win", "amount": "10.15", "transactionId": "win-001"}'
```

**Lose Transaction (Decrease Balance)**:
```bash
curl -X POST http://localhost:8000/user/1/transaction \
  -H "Source-Type: game" \
  -H "Content-Type: application/json" \
  -d '{"state": "lose", "amount": "5.50", "transactionId": "lose-001"}'
```

**Response**: `200 OK` on success, error status codes on failure

### Balance Endpoint

**Endpoint**: `GET /user/{userId}/balance`

**Example Request**:
```bash
curl http://localhost:8000/user/1/balance
```

**Response Format**:
```json
{
  "userId": 1,
  "balance": "104.65"
}
```

**Field Specifications**:
- `userId`: uint64 - The user identifier
- `balance`: string - Current balance rounded to 2 decimal places

## Configuration

Environment variables are configured in `.env`:

```env
# Database Configuration
DATABASE_URL=postgres://postgres:postgres@postgres:5432/go_banking?sslmode=disable
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_banking

# Server Configuration
SERVER_ADDRESS=0.0.0.0
SERVER_PORT=8000
```

## Testing

### Automated Testing

The application is designed to work with automated testing tools.

**Run application tests**:
```bash
go test -v ./...
```

**Test with coverage**:
```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Manual Testing Examples

**Test predefined users** (should work immediately after startup):
```bash
# Check initial balances
curl http://localhost:8000/user/1/balance  # Should return {"userId": 1, "balance": "100.00"}
curl http://localhost:8000/user/2/balance  # Should return {"userId": 2, "balance": "50.00"}
curl http://localhost:8000/user/3/balance  # Should return {"userId": 3, "balance": "25.00"}
```

**Test transaction processing**:
```bash
# Win transaction (increase balance)
curl -X POST http://localhost:8000/user/1/transaction \
  -H "Source-Type: game" \
  -H "Content-Type: application/json" \
  -d '{"state": "win", "amount": "15.25", "transactionId": "test-win-001"}'

# Check updated balance
curl http://localhost:8000/user/1/balance  # Should return {"userId": 1, "balance": "115.25"}

# Lose transaction (decrease balance)
curl -X POST http://localhost:8000/user/1/transaction \
  -H "Source-Type: payment" \
  -H "Content-Type: application/json" \
  -d '{"state": "lose", "amount": "10.00", "transactionId": "test-lose-001"}'

# Check updated balance
curl http://localhost:8000/user/1/balance  # Should return {"userId": 1, "balance": "105.25"}
```

**Test idempotency** (duplicate transaction handling):
```bash
# First request (should succeed)
curl -X POST http://localhost:8000/user/1/transaction \
  -H "Source-Type: game" \
  -H "Content-Type: application/json" \
  -d '{"state": "win", "amount": "5.00", "transactionId": "duplicate-test"}'

# Duplicate request (should be ignored, same transactionId)
curl -X POST http://localhost:8000/user/1/transaction \
  -H "Source-Type: game" \
  -H "Content-Type: application/json" \
  -d '{"state": "win", "amount": "5.00", "transactionId": "duplicate-test"}'
```

**Test negative balance protection**:
```bash
# Try to lose more than available balance (should fail)
curl -X POST http://localhost:8000/user/3/transaction \
  -H "Source-Type: game" \
  -H "Content-Type: application/json" \
  -d '{"state": "lose", "amount": "100.00", "transactionId": "negative-test"}'
```

### Creating New Users

To create new users via the API, use the following endpoint:

**Endpoint**: `POST /user`

**Request Body**:
```json
{
    "username": "newuser",
    "full_name": "New User",
    "email": "newuser@example.com"
}
```

**Response**:
```json
{
    "userId": 4,
    "username": "newuser",
    "full_name": "New User",
    "email": "newuser@example.com"
}
```

### Performance Testing

The application is designed to handle **20-30 RPS** as specified in the requirements.

**Load testing with curl** (basic):
```bash
# Simple load test (25 parallel requests)
seq 1 250 | xargs -n1 -P25 -I{} curl -X POST http://localhost:8000/user/1/transaction \
  -H "Source-Type: game" \
  -H "Content-Type: application/json" \
  -d '{"state": "win", "amount": "0.01", "transactionId": "load-test-{}"}'
```

## 👨‍💻 Author

**Vikram Rathore** - [@rathorevk](https://github.com/rathorevk)

## 🔗 Links

- [Repository](https://github.com/rathorevk/go-banking)
