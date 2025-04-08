# ZenAuth - OAuth 2.0 Authorization Server in Go

ZenAuth is a lightweight, yet fully-featured OAuth 2.0 authorization server implemented in Go. It supports multiple OAuth 2.0 flows and provides a simple, secure way to handle authentication and authorization for your applications.

![ZenAuth Logo](assets/logo.png)

## Table of Contents
- [ZenAuth - OAuth 2.0 Authorization Server in Go](#zenauth---oauth-20-authorization-server-in-go)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Project Structure](#project-structure)
  - [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Setup and Running](#setup-and-running)
  - [Environment Variables](#environment-variables)
  - [API Endpoints](#api-endpoints)
  - [OAuth Flow Implementation](#oauth-flow-implementation)
    - [Authorization Code Flow with PKCE](#authorization-code-flow-with-pkce)
    - [Client Credentials Flow](#client-credentials-flow)
    - [Refresh Token Flow](#refresh-token-flow)
  - [Testing the Flows (via curl)](#testing-the-flows-via-curl)
    - [Client Credentials](#client-credentials)
    - [Authorization Code Flow (with PKCE plain)](#authorization-code-flow-with-pkce-plain)
    - [Refresh Token](#refresh-token)
  - [Development Commands](#development-commands)
  - [License](#license)
  - [Contributing](#contributing)

---

## Features

- **Multiple OAuth 2.0 Flows**:
  - Authorization Code flow with PKCE support
  - Client Credentials flow
  - Refresh Token flow

- **Security Features**:
  - PKCE (Proof Key for Code Exchange) support
  - JWT-based access tokens
  - Secure password hashing with bcrypt
  - CORS protection
  - Single-use authorization codes

- **PostgreSQL Storage**:
  - Persistent storage for users, clients, authorization codes, and refresh tokens

---

## Project Structure

```
ZenAuth/
├── client/             # Test OAuth client application
├── cmd/                     # Application entry points
├── config/                  # Configuration handling
├── handlers/                # HTTP request handlers
├── middlewares/             # HTTP middleware functions
├── models/                  # Data models
├── oauth/                   # OAuth implementation
│   └── store/               # Database interactions
├── scripts/                 # Utility scripts
└── templates/               # HTML templates
```

---

## Getting Started

### Prerequisites

- Go 1.20+
- Docker and Docker Compose (for PostgreSQL)
- Make (optional, for running commands)

### Setup and Running

1. **Clone the repository**:
   ```bash
   git clone https://github.com/yourusername/zenauth.git
   cd zenauth
   ```

2. **Start the PostgreSQL database**:
   ```bash
   make up
   ```

3. **Initialize the database and create a test user**:
   ```bash
   make init-db seed
   ```

4. **Run the ZenAuth server**:
   ```bash
   make run
   ```

5. **In a separate terminal, start the test client**:
   ```bash
   make client
   ```

6. Visit [http://localhost:3000](http://localhost:3000) and click "Login with Mini OAuth"
   - Use the test credentials:
     - Username: `demo`
     - Password: `demo123`

---

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```
SERVER_PORT=8080
DATABASE_URL=postgres://oauth_user:oauth_pass@localhost:5432/oauth?sslmode=disable
JWT_SECRET=supersecretkey
FRONTEND_ORIGIN=http://localhost:3000
```

---

## API Endpoints

| Method | Path         | Description                          |
| ------ | ------------ | ------------------------------------ |
| GET    | `/authorize` | Starts the authorization code flow   |
| POST   | `/token`     | Exchanges code or client credentials |
| GET    | `/userinfo`  | Returns user info from access token  |

---

## OAuth Flow Implementation

### Authorization Code Flow with PKCE

```mermaid
graph TD;
  A[Client App] -->|Redirect with code_challenge| B[ZenAuth /authorize];
  B --> C[Login Form];
  C --> D[Authorization Code];
  D -->|Redirect to client| A;
  A -->|POST /token with code & verifier| E[ZenAuth /token];
  E --> F[Access Token + Refresh Token];
```

### Client Credentials Flow

```mermaid
graph TD;
  A[Client App] -->|POST /token with credentials| B[ZenAuth];
  B --> C[Access Token + Refresh Token];
```

### Refresh Token Flow

```mermaid
graph TD;
  A[Client App] -->|POST /token with refresh_token| B[ZenAuth];
  B --> C[New Access Token];
```

---

## Testing the Flows (via curl)

### Client Credentials
```bash
curl -X POST http://localhost:8080/token   -u demo-client:demo-secret   -d "grant_type=client_credentials"
```

### Authorization Code Flow (with PKCE plain)
```bash
export CODE_VERIFIER=demo123
export CODE_CHALLENGE=demo123
```
Visit:
```
http://localhost:8080/authorize?client_id=demo-client&redirect_uri=http://localhost:3000&response_type=code&code_challenge=$CODE_CHALLENGE&code_challenge_method=plain
```
Then exchange the code:
```bash
curl -X POST http://localhost:8080/token   -d "grant_type=authorization_code"   -d "code=xxx"   -d "redirect_uri=http://localhost:3000"   -d "code_verifier=$CODE_VERIFIER"
```

### Refresh Token
```bash
curl -X POST http://localhost:8080/token   -d "grant_type=refresh_token"   -d "refresh_token=xxx"
```

---

## Development Commands

- `make up` - Start PostgreSQL container
- `make init-db` - Initialize database schema
- `make seed` - Create test user
- `make run` - Start ZenAuth server
- `make client` - Start test client
- `make logs` - View PostgreSQL logs
- `make down` - Stop all containers

---

## License
MIT License

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.