# ZenAuth - OAuth 2.0 Authorization Server in Go

ZenAuth is a lightweight yet complete OAuth 2.0 authorization server implemented in Go. It supports multiple OAuth 2.0 flows and provides a simple and secure solution for managing authentication and authorization for your applications.

![ZenAuth Logo](assets/logo.png)

## Table of Contents
- [ZenAuth - OAuth 2.0 Authorization Server in Go](#zenauth---oauth-20-authorization-server-in-go)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Installation and Execution](#installation-and-execution)
  - [Environment Variables](#environment-variables)
  - [API Endpoints](#api-endpoints)
  - [OAuth Flow Implementation](#oauth-flow-implementation)
    - [Authorization Code Flow with PKCE](#authorization-code-flow-with-pkce)
    - [Client Credentials Flow](#client-credentials-flow)
    - [Refresh Token Flow](#refresh-token-flow)
    - [External Authentication Flow](#external-authentication-flow)
  - [Testing Flows (via curl)](#testing-flows-via-curl)
    - [Client Credentials](#client-credentials)
    - [Authorization Code (with PKCE plain)](#authorization-code-with-pkce-plain)
    - [Refresh Token](#refresh-token)
  - [Operation Modes](#operation-modes)
    - [Standalone Mode](#standalone-mode)
    - [Hybrid Mode](#hybrid-mode)
    - [Integration with External Database](#integration-with-external-database)
  - [External Authentication Providers](#external-authentication-providers)
    - [Configuration](#configuration)
  - [Development Commands](#development-commands)
  - [License](#license)
  - [Contribution](#contribution)

---

## Features

- **Multiple OAuth 2.0 Flows**:
  - Authorization Code flow with PKCE support
  - Client Credentials flow
  - Refresh Token flow

- **External Authentication**:
  - Support for popular identity providers (Microsoft, Google, GitHub)
  - Streamlined login experience with social sign-in buttons

- **Security Features**:
  - PKCE (Proof Key for Code Exchange) support
  - JWT-based access tokens
  - Secure password hashing with bcrypt
  - CORS protection
  - Single-use authorization codes

- **PostgreSQL Storage**:
  - Persistent storage for users, clients, authorization codes, and refresh tokens

- **Role-Based Access Control**:
  - User roles and groups management
  - Role inclusion in JWT tokens

- **Integration Flexibility**:
  - Standalone mode with internal database
  - Hybrid mode with external user authentication and roles
  - Support for custom database schemas

- **Admin Console**:
  - Web-based admin interface
  - Manage users, clients, and auth providers
  - Role and group assignment

---

## Getting Started

### Prerequisites

- Go 1.20+
- Docker and Docker Compose (for PostgreSQL)
- Make (optional, for running commands)

### Installation and Execution

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

6. Visit [http://localhost:3000](http://localhost:3000) and click on "Login with Mini OAuth"
   - Use the test credentials:
     - Username: `demo`
     - Password: `demo123`
   - Or use one of the configured external authentication providers
      - Microsoft
      - Google
      - GitHub
  
7. Access the admin console at http://localhost:8080/admin/

---

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```
# Base Configuration
SERVER_PORT=8080
DATABASE_URL=postgres://oauth_user:oauth_pass@localhost:5432/oauth?sslmode=disable
JWT_SECRET=supersecretkey
FRONTEND_ORIGIN=http://localhost:3000

# User Provider Configuration
USER_PROVIDER_TYPE=default  # options: default, sql
USER_PROVIDER_SQL_CONN=postgres://user:pass@host:port/db?sslmode=disable
USER_PROVIDER_SQL_TABLE=users
USER_PROVIDER_SQL_ID_FIELD=id
USER_PROVIDER_SQL_USER_FIELD=username
USER_PROVIDER_SQL_PASS_FIELD=password_hash
USER_PROVIDER_SQL_EMAIL_FIELD=email

# Role Manager Configuration
ROLE_MANAGER_TYPE=default  # options: default, external
ROLE_MANAGER_INCLUDE_IN_JWT=true
ROLE_MANAGER_EXTERNAL_CONN=postgres://user:pass@host:port/db?sslmode=disable
ROLE_MANAGER_ROLE_TABLE=roles
ROLE_MANAGER_GROUP_TABLE=groups
```

---

## API Endpoints

| Method | Path             | Description                                                          |
| ------ | ---------------- | -------------------------------------------------------------------- |
| GET    | `/authorize`     | Starts the Authorization Code flow                                   |
| POST   | `/token`         | Exchanges code or client credentials                                 |
| GET    | `/userinfo`      | Returns user information from token                                  |
| GET    | `/admin/`        | Admin console to manage users, clients, and authentication providers |
| GET    | `/auth/external` | Starts external authentication flow                                  |
| GET    | `/auth/callback` | Callback URL for external authentication providers                   |

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

### External Authentication Flow

```mermaid
graph TD;
  A[Client App] -->|Redirect to| B[ZenAuth /authorize];
  B --> C[Login Form with External Providers];
  C -->|Select External Provider| D[/auth/external];
  D -->|Redirect to Provider| E[Identity Provider];
  E -->|Authentication| F[/auth/callback/];
  F -->|Authorization Code| A;
  A -->|POST /token| G[Access Token + Refresh Token];
```

---

## Testing Flows (via curl)

### Client Credentials
```bash
curl -X POST http://localhost:8080/token   -u demo-client:demo-secret   -d "grant_type=client_credentials"
```

### Authorization Code (with PKCE plain)
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
curl -X POST http://localhost:8080/token   -d "grant_type=refresh_token"   -d "refresh_token=xxx"   -d "client_id=demo-client"   -d "client_secret=demo-secret"
```

---

## Operation Modes

ZenAuth can be used in different modes depending on your needs.

### Standalone Mode

In this mode, ZenAuth manages both user authentication and OAuth data in its own PostgreSQL database.

```
USER_PROVIDER_TYPE=default
ROLE_MANAGER_TYPE=default
```

### Hybrid Mode

ZenAuth manages OAuth data (clients, authorization codes, tokens) in its own database, but uses an external database for user authentication and roles.

```
DATABASE_URL=postgres://oauth_user:oauth_pass@localhost:5432/oauth?sslmode=disable
USER_PROVIDER_TYPE=sql
USER_PROVIDER_SQL_CONN=postgres://user:pass@external-db:5432/app_db?sslmode=disable
ROLE_MANAGER_TYPE=external
ROLE_MANAGER_EXTERNAL_CONN=postgres://user:pass@external-db:5432/app_db?sslmode=disable
```

### Integration with External Database

Example configuration for using an existing database:

```
# Connection to an external database for users
USER_PROVIDER_TYPE=sql
USER_PROVIDER_SQL_CONN=postgres://postgres:password@192.168.1.10:5432/app_db?sslmode=disable
USER_PROVIDER_SQL_TABLE=users
USER_PROVIDER_SQL_ID_FIELD=uuid
USER_PROVIDER_SQL_USER_FIELD=username
USER_PROVIDER_SQL_PASS_FIELD=password
USER_PROVIDER_SQL_EMAIL_FIELD=email

# Connection to an external database for roles
ROLE_MANAGER_TYPE=external
ROLE_MANAGER_INCLUDE_IN_JWT=true
ROLE_MANAGER_EXTERNAL_CONN=postgres://postgres:password@192.168.1.10:5432/app_db?sslmode=disable
ROLE_MANAGER_ROLE_TABLE=roles
ROLE_MANAGER_GROUP_TABLE=groups
```

## External Authentication Providers
ZenAuth supports integration with popular identity providers to enable social login and enterprise authentication.

- **Microsoft**: Use Microsoft accounts for authentication.
- **Google**: Authenticate users with their Google accounts.
- **GitHub**: Allow users to sign in using their GitHub accounts.

### Configuration
External authentication providers can be configured via the admin console at /admin/ in the "Auth Providers" section.

For each provider, you will need:

- Provider Name (display name)
- Provider Type (microsoft, google, github)
- Client ID (from the provider's developer console)
- Client Secret (from the provider's developer console)
- Tenant ID (only for Microsoft, optional)

---

## Development Commands

- `make up` - Start the PostgreSQL container
- `make init-db` - Initialize the database schema
- `make seed` - Create a test user
- `make run` - Start the ZenAuth server
- `make client` - Start the test client
- `make logs` - Show PostgreSQL logs
- `make down` - Stop all containers

---

## License
MIT License

## Contribution
Contributions are welcome! Feel free to submit a Pull Request.