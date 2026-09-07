# Workout Tracker API Postman Collection

## Purpose
Professional Postman setup for testing the Workout Tracker API endpoints.

## Files Included

### Environment
- `Workout Tracker Environment.environment.yaml`
  - Configuration variables:
    - `base_url`: `http://localhost:8080` (development default)
    - `access_token`: Automatically populated after login
    - `test_email`: `test@example.com`
    - `test_password`: `TestPassword123`
    - `test_display_name`: `Test User`
    - `basic_auth_username`: (from BASIC_AUTH_USERNAME env var, default empty)
    - `basic_auth_password`: (from BASIC_AUTH_PASSWORD env var, default empty)

### Collection
- `Workout Tracker API` collection containing:
  - **Debug** folder with:
    - `Debug Vars.request.yaml` - GET /debug/vars (expvar metrics - requires basic auth)
    - `Statsviz Index.request.yaml` - GET /debug/statsviz/ (statsviz interface - requires basic auth)
    - `Statsviz Wildcard.request.yaml` - GET /debug/statsviz/* (statsviz resources - requires basic auth)
    - `Statsviz WebSocket.request.yaml` - GET /debug/statsviz/ws (statsviz websocket - websocket bypasses basic auth)
  - **Auth** folder with:
    - Positive test requests:
      - `Signup.request.yaml` - Create new user account
      - `Login.request.yaml` - Authenticate user (automatically saves token)
      - `Logout.request.yaml` - Logout (requires valid token)
    - Negative test requests:
      - Invalid signup scenarios (duplicate email, invalid email, missing fields)
      - Login failure scenarios (wrong password, nonexistent user)
      - Logout failure scenarios (missing auth header, invalid token)

### Flows
- `Authentication Flow.flow.yaml` - Complete test sequence:
  1. Debug endpoints (access with basic auth)
  2. Authentication sequence: Signup → Login → Logout
  3. Verify logout by attempting again (should fail with 401)

### Documentation
- `workout-tracker-api.yaml` - OpenAPI specification
  - Complete API documentation with:
    - Debug endpoints (expvar, statsviz)
    - Authentication flow
    - Error response format
    - Request/response examples
    - Security requirements (Bearer token, Basic Auth)
    - HTTP status codes

## Usage Instructions

### Import into Postman
1. Open Postman
2. Click "Import" → "Upload Files" or "Folder"
3. Select the entire `postman/` directory
4. Or import individually:
   - Environment: `Workout Tracker Environment.environment.yaml`
   - Collection: `Workout Tracker API` collection
   - Flows: `Authentication Flow.flow.yaml`
   - Spec: `workout-tracker-api.yaml`

### Run the Complete Test Flow
1. Select "Workout Tracker API" collection
2. Choose "Workout Tracker Environment" environment
3. In Collection Runner (or use Flows):
   - Execute in this order:
     1. `Debug Vars.request.yaml` - Returns expvar metrics (requires basic auth)
     2. `Statsviz Index.request.yaml` - Returns statsviz HTML interface (requires basic auth)
     3. `Statsviz Wildcard.request.yaml` - Returns statsviz resource (requires basic auth)
     4. `Statsviz WebSocket.request.yaml` - Attempts websocket connection (bypasses basic auth)
     5. `Signup.request.yaml` (creates user, gets token)
     6. `Login.request.yaml` (saves token to environment)
     7. `Logout.request.yaml` (uses token, returns 204)
     8. `Logout Missing Auth Header.request.yaml` (verify 401 after logout)

### Key Features
- ✅ Automatic token handling (login saves token to environment)
- ✅ Protected endpoint support (logout requires valid token)
- ✅ Basic auth support for debug endpoints
- ✅ WebSocket endpoint special handling (bypasses basic auth)
- ✅ Comprehensive negative testing
- ✅ Consistent error response validation
- ✅ OpenAPI specification for reference
- ✅ Complete documentation of all endpoints

## API Endpoints Reference

### Debug Endpoints (Require Basic Auth)
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/debug/vars` | Expvar metrics | ✅ Basic Auth |
| GET | `/debug/statsviz/` | Statsviz interface | ✅ Basic Auth |
| GET | `/debug/statsviz/*` | Statsviz resources | ✅ Basic Auth |
| GET | `/debug/statsviz/ws` | Statsviz WebSocket | ❌ (WebSocket bypass) |

### Authentication Endpoints

#### Signup
- **Method**: POST
- **URL**: `{{base_url}}/api/v1/auth/signup`
- **Headers**:
  - `Content-Type: application/json`
  - `Accept: application/json`
- **Body**:
  ```json
  {
    "email": "{{test_email}}",
    "password": "{{test_password}}",
    "display_name": "{{test_display_name}}"
  }
  ```
- **Success**: `201 Created`
- **Response**:
  ```json
  {
    "data": {
      "access_token": "...",
      "token_type": "Bearer",
      "expires_in": 900
    }
  }

#### Login
- **Method**: POST
- **URL**: `{{base_url}}/api/v1/auth/login`
- **Headers**:
  - `Content-Type: application/json`
  - `Accept: application/json`
- **Body**:
  ```json
  {
    "email": "{{test_email}}",
    "password": "{{test_password}}"
  }
  ```
- **Success**: `200 OK`
- **Response**: Same as signup with token in `access_token` environment variable

#### Logout
- **Method**: POST
- **URL**: `{{base_url}}/api/v1/auth/logout`
- **Headers**:
  - `Accept: application/json`
  - `Authorization: Bearer {{access_token}}`
- **Success**: `204 No Content`
- **Response**: Empty body

## Authentication Flow
1. **Debug Endpoints** → Test basic auth protected endpoints
2. **Signup** → Creates user and returns token
3. **Login** → Validates credentials, returns token (automatically saved)
4. **Logout** → Discards token (requires valid token)
5. **Verify** → Attempt logout again with same token (should fail with 401)

## Basic Authentication
The debug endpoints (`/debug/vars`, `/debug/statsviz/*`) require Basic Authentication.
Set these environment variables:
- `basic_auth_username`: From `BASIC_AUTH_USERNAME` env var
- `basic_auth_password`: From `BASIC_AUTH_PASSWORD` env var

Note: The WebSocket endpoint (`/debug/statsviz/ws`) bypasses basic auth per middleware configuration.

## Negative Test Cases Covered
- Signup: duplicate email, invalid email, missing fields, short password
- Login: wrong password, nonexistent user, invalid body, missing fields
- Logout: missing auth header, invalid token