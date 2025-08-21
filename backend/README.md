# Backend (Secure Version) - Communication_LTD

## Project Overview

This document outlines the backend for the Communication_LTD project. This backend provides a secure API for user management, authentication, password policy enforcement, customer management, and a secure password reset mechanism via email.

This implementation focuses on security best practices to protect user data and ensure a robust and reliable system.

## Requirements / Prerequisites

To run this backend, you will need the following installed:

*   **Go:** Version 1.24 or higher
*   **MySQL:** Version 8 or higher
*   **Git:** For cloning the repository
*   **Docker:** Version 20.10 or higher (optional, for containerized setup)
*   **Docker Compose v2:** (optional, for containerized setup)

**Note:** Node.js is only required for the frontend and is not needed to run the backend.

## Tech Stack

*   **Backend:** Go with the [Echo](https://echo.labstack.com/) framework.
*   **Database Driver:** [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) with [sqlx](https://github.com/jmoiron/sqlx) for prepared statements.
*   **Configuration:** [godotenv](https://github.com/joho/godotenv) for loading environment variables.
*   **Mail:** SMTP integration with [MailHog](https://github.com/mailhog/MailHog) for development.

## Folder Structure

```
.
├── cmd/
│   └── main.go
├── internal/
│   ├── handlers/
│   ├── services/
│   └── repository/
├── config/
│   ├── password-policy.toml
│   └── .env.example
├── migrations/
│   ├── schema.sql
│   └── seed.sql
├── go.mod
├── go.sum
└── README.md
```

*   `cmd/`: Main entry point of the application.
*   `internal/handlers/`: Contains the Echo route handlers.
*   `internal/services/`: Implements the business logic (e.g., authentication, customer management).
*   `internal/repository/`: Handles database access using prepared statements.
*   `config/`: Contains configuration files.
*   `migrations/`: Database schema and seed data.

## Configuration

### Environment Variables

Copy the `.env.example` file to `.env` and update the values as needed.

```bash
cp config/.env.example .env
```

The following variables need to be set:

*   `DB_HOST`: The database host.
*   `DB_PORT`: The database port.
*   `DB_USER`: The database user.
*   `DB_PASS`: The database password.
*   `DB_NAME`: The database name.
*   `HMAC_SECRET`: A secret key for HMAC hashing.
*   `JWT_SECRET`: A secret key for signing JWTs.
*   `SMTP_HOST`: The SMTP server host.
*   `SMTP_PORT`: The SMTP server port.
*   `SMTP_FROM`: The email address to send emails from.

### Password Policy

The password policy is defined in `config/password-policy.toml`. The default values are:

```toml
min_length = 10
complexity_rules = ["has_upper", "has_lower", "has_digit", "has_special"]
history = 3
max_login_attempts = 3
lockout_minutes = 15
```

These values can be changed by an administrator.

## Database Schema

The database schema is defined in `migrations/schema.sql`. The following tables are used:

*   **users:** Stores user information, including the hashed password and salt.
*   **password_history:** Stores a history of the user's previous passwords to prevent reuse.
*   **password_reset_tokens:** Stores tokens for password reset requests.
*   **customers:** Stores customer information.
*   **login_attempts:** (Optional) Tracks login attempts to enforce the lockout policy.

## API Routes

### Authentication

*   **POST `/api/register`**
    *   **Description:** Registers a new user.
    *   **Input:** `{"username": "user", "password": "password"}`
    *   **Output:** `{"message": "User registered successfully"}`

*   **POST `/api/login`**
    *   **Description:** Authenticates a user and returns a JWT.
    *   **Input:** `{"username": "user", "password": "password"}`
    *   **Output:** `{"token": "jwt_token"}`

### Password Management

*   **POST `/api/password/change`**
    *   **Description:** Changes the user's password.
    *   **Input:** `{"old_password": "old", "new_password": "new"}`
    *   **Output:** `{"message": "Password changed successfully"}`

*   **POST `/api/password/forgot`**
    *   **Description:** Sends a password reset link to the user's email.
    *   **Input:** `{"email": "user@example.com"}`
    *   **Output:** `{"message": "Password reset email sent"}`

*   **POST `/api/password/reset`**
    *   **Description:** Resets the user's password using a token.
    *   **Input:** `{"token": "reset_token", "new_password": "new"}`
    *   **Output:** `{"message": "Password reset successfully"}`

### Customer Management

*   **POST `/api/customers`**
    *   **Description:** Creates a new customer.
    *   **Input:** `{"name": "Customer Name", "email": "customer@example.com"}`
    *   **Output:** `{"id": 1, "name": "Customer Name", "email": "customer@example.com"}`

*   **GET `/api/customers`**
    *   **Description:** Returns a list of all customers.
    *   **Output:** `[{"id": 1, "name": "Customer Name", "email": "customer@example.com"}]`

## Security Features Implemented

*   **Password Storage:** Passwords are not stored in plaintext. They are hashed using HMAC with a salt.
*   **Password Reset:** Password reset tokens are stored as SHA-1 hashes in the database. The raw token is sent to the user via email and is required to reset the password.
*   **Password Policy:** A strong password policy is enforced via the `config/password-policy.toml` file.
*   **SQL Injection:** All database queries are executed using prepared statements to prevent SQL injection attacks.
*   **XSS Protection:** All outputs are properly escaped. The React frontend also escapes HTML by default.
*   **Login Throttling:** Login attempts are limited to 3 by default, with a lockout period of 15 minutes.
*   **Error Messages:** Error messages are generic and do not reveal sensitive information.

## Running the Backend

### With Docker

```bash
docker build -t backend-secure .
docker run -p 8080:8080 --env-file .env backend-secure
```

### With Docker Compose

If this backend is part of a larger project with a `docker-compose.yml` file, you can run it with:

```bash
docker-compose up -d backend
```

### Directly

```bash
go run cmd/main.go
```

## Compliance with Project Requirements

*   **User Management:** Implemented via the `/api/register`, `/api/login`, `/api/password/change`, `/api/password/forgot`, and `/api/password/reset` routes.
*   **Password Storage:** Passwords are stored using HMAC with a salt.
*   **Password Reset:** Reset tokens are hashed with SHA-1.
*   **Password Policy:** The password policy is loaded from the `config/password-policy.toml` file.
*   **SQL Injection Protection:** Prepared statements are used for all database queries.
*   **XSS Protection:** Outputs are escaped, and the frontend is expected to handle HTML escaping.
*   **Login Throttling:** Login attempts are limited to 3 by default.

## Development Notes

*   **MailHog:** MailHog is used for development and testing of email functionality. It is not intended for production use.
*   **Password Hashing:** For production, it is recommended to use a stronger hashing algorithm like bcrypt or Argon2 instead of HMAC+Salt.
*   **Configuration:** The default lockout period and password rules can be changed in the `config/password-policy.toml` file.
*   **CORS:** If the frontend is running on a different domain or port, you may need to enable CORS in the backend. By default, only `localhost:3000` is allowed.

## License/Authors

This project is licensed under the MIT License.

**Authors:**
*   [Your Name](https://github.com/your-username)
