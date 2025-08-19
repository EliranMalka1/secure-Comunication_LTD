# Secure Communication LTD

## System Architecture Diagram

```mermaid
graph TD
    subgraph "User Interface"
        A[<img src="https://www.vectorlogo.zone/logos/reactjs/reactjs-icon.svg" width="40" height="40" /><br/><strong>React Frontend</strong><br/>(localhost:3000)]
    end

    subgraph "Backend Services"
        B[<img src="https://www.vectorlogo.zone/logos/golang/golang-icon.svg" width="40" height="40" /><br/><strong>Go Backend API</strong><br/>(localhost:8080)]
    end

    subgraph "Data Storage"
        C[<img src="https://www.vectorlogo.zone/logos/mysql/mysql-icon.svg" width="40" height="40" /><br/><strong>MySQL Database</strong>]
    end

    subgraph "Email Service"
        D[<img src="https://raw.githubusercontent.com/mailhog/MailHog/master/docs/MailHog-Logo.png" width="40" height="40" /><br/><strong>MailHog SMTP/Web UI</strong><br/>(localhost:8025)]
    end

    A -- "HTTP/REST" --> B
    B -- "SQL" --> C
    B -- "SMTP" --> D

    style A fill:#282c34,stroke:#61DAFB,stroke-width:2px,color:#fff
    style B fill:#00ADD8,stroke:#fff,stroke-width:2px,color:#fff
    style C fill:#4479A1,stroke:#fff,stroke-width:2px,color:#fff
    style D fill:#c7c7c7,stroke:#000,stroke-width:2px,color:#000
```

# Communication_LTD (Secure Version)

## Project Overview

Welcome to **Communication_LTD (Secure Version)**, a web application developed as a final project for a cybersecurity course. This application provides a platform for managing users and customers, with a strong emphasis on security best practices.

The core functionalities of the system include:
-   **User Management**: Secure registration and login functionalities.
-   **Password Policy**: Enforces strong password creation and management rules.
-   **Customer Management**: Basic CRUD operations for managing customer data.
-   **Forgot/Reset Password**: A secure flow for users to recover their accounts.

This version of the project has been specifically hardened against common web vulnerabilities, serving as a practical example of securing a modern web application.

---

## Security Features Implemented

This project incorporates several security measures to protect user data and prevent attacks:

| Feature | Implementation | Protection Against |
| :--- | :--- | :--- |
| **Password Hashing** | HMAC+SHA512 with a unique Salt for each user. | Rainbow table attacks, dictionary attacks. |
| **Reset Token Hashing** | SHA-1 for generating password reset tokens. | Token guessing. |
| **Password Policy** | - Minimum length (e.g., 12 characters).<br>- Complexity (uppercase, lowercase, numbers, symbols).<br>- Password history to prevent reuse.<br>- Account lockout after multiple failed login attempts. | Brute-force attacks, weak passwords. |
| **SQL Injection (SQLi)** | Use of prepared statements in all database queries. | Injection of malicious SQL code. |
| **Cross-Site Scripting (XSS)** | - **Backend**: Strict output encoding on all data rendered in templates.<br>- **Frontend**: React's inherent data binding and JSX escaping. | Injection of malicious scripts into the web application. |
| **Cross-Site Request Forgery (CSRF)** | Implementation of Anti-CSRF tokens in sensitive forms. | Unauthorized commands being performed on behalf of an authenticated user. |
| **Rate Limiting** | Throttling login attempts and other sensitive endpoints. | Brute-force attacks, denial-of-service (DoS). |

---

## Tech Stack

| Component | Technology | Description |
| :--- | :--- | :--- |
| **Frontend** | [React](https://reactjs.org/) (with Node.js) | A JavaScript library for building user interfaces. |
| **Backend** | [Go](https://golang.org/) with [Echo Framework](https://echo.labstack.com/) | A high-performance, minimalist Go web framework. |
| **Database** | [MySQL](https://www.mysql.com/) | A popular open-source relational database. |
| **Mail Server** | [MailHog](https://github.com/mailhog/MailHog) | An email testing tool for developers (for Forgot Password flow). |
| **Containerization** | [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/) | For creating and managing isolated application environments. |

---

## Prerequisites / Requirements

Before you begin, ensure you have the following tools installed on your system:

-   **Go**: Version 1.24 or higher
-   **Node.js**: Version 18 or higher (with `npm` or `yarn`)
-   **Docker**: Version 20.10 or higher
-   **Docker Compose**: Version 2.x
-   **Git**: For cloning the repository
-   **Optional**: `golang-migrate` for database migrations if you prefer to run them outside of Docker.

---


## Folder Structure

The actual directory layout for this project is:

```
secure-Comunication_LTD/
├── backend/           # Go backend application
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
├── frontend/          # React frontend application
│   ├── public/
│   │   └── index.html
│   ├── src/
│   │   └── App.js
│   ├── package.json
│   └── Dockerfile
├── db/
│   └── init.sql       # SQL initialization file
├── docker-compose.yml # Docker Compose configuration
├── .gitignore         # Git ignore file
└── README.md          # Project documentation
```

---

## Setup Instructions

1.  **Clone the Repository**
    ```bash
    git clone https://github.com/your-username/communication_ltd.git
    cd communication_ltd
    ```

2.  **Configure Environment Variables**
    Copy the example environment file and customize it with your local settings.
    ```bash
    cp .env.example .env
    ```
    *Fill in the required values in the `.env` file as described in the Environment Variables section.*

3.  **Build and Run Containers**
    Use Docker Compose to build the images and start all services in detached mode.
    ```bash
    docker compose up -d --build
    ```

4.  **Run Database Migrations & Seed**
    Execute the migration files to set up the database schema and optionally seed it with initial data.
    ```bash
    # Example command (adjust if using a different migration tool)
    docker compose exec backend ./migrate -path /migrations -database "mysql://user:password@tcp(db:3306)/dbname" up
    ```

---

## Access Points

Once the services are running, you can access them at the following locations:

| Service | URL | Port |
| :--- | :--- | :--- |
| **React Frontend** | `http://localhost:3000` | `3000` |
| **Go Backend API** | `http://localhost:8080` | `8080` |
| **MailHog Web UI** | `http://localhost:8025` | `8025` |

---

## Environment Variables

The `.env` file is crucial for configuring the application.

| Variable | Description | Example |
| :--- | :--- | :--- |
| `DB_HOST` | Database host name. | `db` |
| `DB_PORT` | Database port. | `3306` |
| `DB_USER` | Database username. | `user` |
| `DB_PASSWORD` | Database password. | `secret` |
| `DB_NAME` | Database name. | `communication_ltd` |
| `JWT_SECRET` | Secret key for signing JWTs. | `a-very-strong-secret-key` |
| `SMTP_HOST` | MailHog SMTP server host. | `mailhog` |
| `SMTP_PORT` | MailHog SMTP server port. | `1025` |
| `API_PORT` | Port for the Go backend API. | `8080` |
| `APP_PORT` | Port for the React frontend. | `3000` |

---

## Running Migrations & Seed Data

To run database migrations manually:
```bash
docker compose exec backend ./migrate -path /migrations -database "mysql://user:password@tcp(db:3306)/dbname" up
```

To seed the database with initial data:
```bash
docker compose exec db mysql -u<user> -p<password> <database_name> < /db/seed/seed.sql
```

---

## How to Run Tests

*(This section should be updated if tests are provided)*

To run backend tests:
```bash
cd backend/
go test ./...
```

To run frontend tests:
```bash
cd frontend/
npm test
```

---

## Usage Examples

1.  **Register a new user**: Navigate to `http://localhost:3000/register` and fill out the form.
2.  **Login**: Go to `http://localhost:3000/login` and enter your credentials.
3.  **Change Password**: Once logged in, go to your profile page to change your password.
4.  **Reset Password**: On the login page, click "Forgot Password", enter your email, and follow the instructions sent to your inbox (viewable in the MailHog UI at `http://localhost:8025`).

---

## Notes for Development

-   **MailHog**: The included MailHog service is for **development and testing only**. It captures all outgoing emails for easy inspection without sending them to actual recipients. For production, you would replace this with a real SMTP service like SendGrid or Amazon SES.
-   **Password Hashing**: HMAC+Salt is used here for educational purposes to demonstrate the principles of hashing and salting. In a **production environment**, it is strongly recommended to use a more robust and battle-tested adaptive hashing algorithm like **bcrypt** or **Argon2**.

---

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

## Authors

-   Eliran Malka(https://github.com/EliranMalka1)
