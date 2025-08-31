# Communication_LTD (Secure Version) - Frontend

This repository contains the frontend for the **Communication_LTD (Secure Version)** project. It is a modern, secure, and user-friendly Single Page Application (SPA) built with React and Vite.

## 1. Project Overview

The frontend provides the client-side interface for user registration, login, and customer data management. It is designed to communicate securely with the backend API, ensuring that all data transmission is encrypted and protected.

Key features include:
-   User registration and authentication.
-   The frontend includes a comprehensive email verification flow, requiring users to verify their email address before they can log in.
-   A secure dashboard for managing customer information.
-   A responsive and intuitive user interface.
-   Support for two-factor authentication (2FA) is planned for a future release.

---

## 2. Requirements / Prerequisites

To run this project in a local development environment or deploy it using Docker, you will need the following tools installed:

-   **Node.js**: `v18.0` or higher.
-   **NPM** or **Yarn**: For managing project dependencies.
-   **Docker**: `v20.10` or higher (for containerized deployment).
-   **Docker Compose**: For simplified container orchestration.

---

## 3. Tech Stack

This project is built with a modern and robust technology stack:

-   **React**: A JavaScript library for building user interfaces.
-   **Vite**: A fast and lightweight build tool for modern web development.
-   **CSS**: Global stylesheets in `App.css` for consistent styling.
-   **Nginx**: A high-performance web server used to serve the production build inside a Docker container.

---




## 4. Folder Structure

The actual directory layout for the frontend is:

```
frontend/
├── public/
│   └── vite.svg
├── src/
│   ├── assets/
│   │   └── react.svg
│   ├── lib/
│   │   └── api.js
│   ├── pages/
│   │   ├── Login.jsx
│   │   └── Register.jsx
│   ├── App.css
│   ├── App.jsx
│   ├── index.css
│   └── main.jsx
├── .dockerignore
├── .env
├── .env.example
├── .gitignore
├── Dockerfile
├── README.md
├── eslint.config.js
├── index.html
├── nginx.conf
├── node_modules/
├── package-lock.json
├── package.json
└── vite.config.js
```

-   `public/`: Contains static assets that are not processed by the build tool (e.g., `vite.svg`).
-   `src/`: Contains all the application source code, including components, styles, and assets.
-   `src/assets/`: Static assets for the app (e.g., `react.svg`).
-   `src/lib/`: Utility libraries and API logic (e.g., `api.js`).
-   `src/pages/`: React page components (e.g., `Login.jsx`, `Register.jsx`). Planned additions include `Dashboard.jsx` and customer management pages.
-   `.env`, `.env.example`: Environment variable files for configuration.
-   `.dockerignore`, `.gitignore`: Ignore files for Docker and Git.
-   `Dockerfile`: Defines the multi-stage Docker build process.
-   `nginx.conf`: Nginx configuration for production deployment.
-   `index.html`: The main HTML template for the application.
-   `eslint.config.js`: ESLint configuration for code linting.
-   `package.json`, `package-lock.json`: Project dependencies and scripts.
-   `vite.config.js`: Vite configuration file.
-   `README.md`: Project documentation.

---

## 5. Security Considerations

Security is a top priority for this project. The following measures have been implemented on the frontend:

-   **Client-Side Input Validation**: The frontend performs validation to check for correct input format (e.g., email format). However, the backend is responsible for enforcing the full password policy and other security-critical validation.
-   **XSS Protection**: React automatically escapes JSX content, providing protection against Cross-Site Scripting (XSS) attacks.
-   **HTTPS Usage**: Recommended for all production deployments to ensure secure data transmission.
-   **No Secrets in Client Code**: Sensitive information, such as JWT secrets, is never stored or exposed in the client-side code.
-   **Secure Cookie Management**: Authentication tokens are stored in HttpOnly cookies, which are managed entirely by the backend. The frontend never stores tokens in `localStorage` or other client-side storage.

---

## 6. Available Scripts

The following scripts are available to run in the project directory:

-   `npm run dev`: Starts the Vite development server with Hot Module Replacement (HMR).
-   `npm run build`: Creates a production-ready build of the application.
-   `npm run preview`: Serves the production build locally for previewing.

---

## 7. Docker Usage

The `Dockerfile` uses a multi-stage build to create a lightweight and secure production image:

1.  **Build Stage**: Uses a Node.js image to build the React application.
2.  **Run Stage**: Copies the build artifacts to a lightweight Nginx image.

To build and run the frontend container:

```bash
# Build the Docker image
docker build -t frontend-secure ./frontend

# Run the container, mapping port 3000 on the host to port 80 in the container (where Nginx is listening).
docker run -p 3000:80 frontend-secure
```

Alternatively, you can use Docker Compose:

```bash
# Start the frontend service in detached mode
docker-compose up -d frontend
```

---

## 8. Compliance with Project Requirements

This frontend application meets the following project requirements:

-   **Client-Side Validation**: Implemented to ensure data integrity.
-   **User-Friendly Interface**: Provides a seamless experience for login and registration.
-   **Secure Communication**: No insecure requests are sent to the backend.
-   **Email Verification**: A complete UI for email verification is supported.
-   **Two-Factor Authentication (2FA)**: The UI for 2FA is planned but not yet implemented.
-   **Customer Management**: Pages for searching and adding customers are planned. The backend will handle SQL injection prevention.

---

## 9. Development Notes

-   **SPA Routing**: The `nginx.conf` file is configured to handle SPA routing by redirecting all requests to `index.html`.
-   **CORS**: The backend is configured to allow requests from the frontend's domain.
-   **Testing**: The application runs on `localhost:5173` in development (Vite) and `localhost:3000` when deployed with Docker.

---

## 10. Environment Variables
Copy `frontend/.env.example` to `frontend/.env` and update values:

```bash
cp frontend/.env.example frontend/.env
```

Required variables:

- `VITE_API_URL`: The URL of the backend API (e.g., `http://localhost:8080`).

---


## 11. License / Authors

This project is licensed under the **MIT License**.

**Author:** [Eliran Malka](https://github.com/EliranMalka1)
