# Authentication

Dashforge supports OAuth 2.0 authentication with GitHub and Google providers, using JWT tokens for session management.

## Overview

```
┌─────────┐     ┌───────────┐     ┌──────────────┐
│ Browser │────▶│ Dashforge │────▶│ GitHub/Google│
│         │◀────│  Server   │◀────│    OAuth     │
└─────────┘     └───────────┘     └──────────────┘
     │                │
     │  JWT Tokens    │
     └────────────────┘
```

1. User clicks "Login with GitHub/Google"
2. Dashforge redirects to OAuth provider
3. User authenticates with provider
4. Provider redirects back with authorization code
5. Dashforge exchanges code for user info
6. Dashforge creates/updates user and returns JWT tokens

## Setting Up OAuth

### GitHub

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click "New OAuth App"
3. Fill in:
   - **Application name**: Dashforge
   - **Homepage URL**: `https://your-domain.com`
   - **Authorization callback URL**: `https://your-domain.com/api/v1/auth/github/callback`
4. Save and note the Client ID and Client Secret

```bash
export GITHUB_CLIENT_ID="your-client-id"
export GITHUB_CLIENT_SECRET="your-client-secret"
```

### Google

1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Create a new project or select existing
3. Go to "Credentials" → "Create Credentials" → "OAuth client ID"
4. Configure consent screen if prompted
5. Select "Web application"
6. Add authorized redirect URI: `https://your-domain.com/api/v1/auth/google/callback`
7. Note the Client ID and Client Secret

```bash
export GOOGLE_CLIENT_ID="your-client-id"
export GOOGLE_CLIENT_SECRET="your-client-secret"
```

## Authentication Endpoints

### Initiate Login

Redirect users to start the OAuth flow:

```
GET /api/v1/auth/github
GET /api/v1/auth/google
```

Optional query parameter:

- `redirect`: URL to redirect after successful login

Example:

```html
<a href="/api/v1/auth/github?redirect=/dashboard">Login with GitHub</a>
```

### OAuth Callbacks

These are called by the OAuth provider (not directly by users):

```
GET /api/v1/auth/github/callback
GET /api/v1/auth/google/callback
```

### Get Current User

```
GET /api/v1/auth/me
Authorization: Bearer <access_token>
```

Response:

```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "role": "viewer",
  "active": true,
  "lastLoginAt": "2024-01-15T10:30:00Z",
  "createdAt": "2024-01-01T00:00:00Z"
}
```

### Refresh Tokens

```
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refreshToken": "your-refresh-token"
}
```

Response:

```json
{
  "accessToken": "new-access-token",
  "refreshToken": "new-refresh-token",
  "expiresIn": 900,
  "tokenType": "Bearer"
}
```

### Logout

```
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
```

Returns `204 No Content`.

## JWT Tokens

### Token Structure

Access tokens contain:

```json
{
  "iss": "dashforge",
  "sub": "1",
  "exp": 1705312200,
  "iat": 1705311300,
  "uid": 1,
  "email": "user@example.com",
  "role": "admin",
  "tid": 1
}
```

| Claim | Description |
|-------|-------------|
| iss | Issuer (always "dashforge") |
| sub | Subject (user ID as string) |
| exp | Expiration timestamp |
| iat | Issued at timestamp |
| uid | User ID (number) |
| email | User email |
| role | User role |
| tid | Tenant ID (for multi-tenancy) |

### Using Tokens

Include the access token in the Authorization header:

```bash
curl -H "Authorization: Bearer eyJhbGc..." \
  https://dashforge.example.com/api/v1/dashboards
```

### Token Expiration

| Token Type | Default Lifetime |
|------------|-----------------|
| Access Token | 15 minutes |
| Refresh Token | 7 days |

## User Roles

Dashforge uses role-based access control:

| Role | Permissions |
|------|-------------|
| viewer | View dashboards, run saved queries |
| editor | viewer + create/edit dashboards |
| admin | editor + manage users, data sources |
| owner | admin + tenant settings, billing |

### Role Hierarchy

```
owner > admin > editor > viewer
```

## Protecting API Routes

Authenticated routes require a valid JWT:

```go
// Protected route
mux.Handle("/api/v1/dashboards",
    jwtService.Middleware(dashboardHandler))

// Role-restricted route
mux.Handle("/api/v1/admin/users",
    jwtService.Middleware(
        auth.RequireJWTRole("admin", "owner")(userHandler)))
```

## Security Best Practices

### JWT Secret

- Use a cryptographically random secret (minimum 32 bytes)
- Never commit secrets to version control
- Rotate secrets periodically
- Use environment variables or secrets manager

```bash
# Generate secure secret
openssl rand -base64 32
```

### HTTPS

Always use HTTPS in production:

```bash
export BASE_URL="https://dashforge.example.com"
```

### Cookie Security

OAuth state cookies are configured with:

- `HttpOnly`: Prevents JavaScript access
- `Secure`: Only sent over HTTPS
- `SameSite=Lax`: CSRF protection
- `MaxAge=600`: 10-minute expiration

### CORS

Configure CORS for your frontend domain:

```yaml
# config.yaml
cors:
  allowed_origins:
    - https://app.example.com
  allowed_methods:
    - GET
    - POST
    - PUT
    - DELETE
  allowed_headers:
    - Authorization
    - Content-Type
```

## Disabling Authentication

For development only:

```bash
./dashforge-server serve --disable-auth
```

!!! danger "Warning"
    Never disable authentication in production. This flag is for local development only.
