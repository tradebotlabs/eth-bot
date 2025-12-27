# Production-Ready MVP Implementation Roadmap
## Ethereum Trading Bot - 14-Week Sprint Plan

**Generated:** December 26, 2025
**Target:** Production-ready, enterprise-grade trading platform
**Team:** 2 Full-Stack Engineers (Backend-focused + Frontend-focused)

---

## Overview

### Current State: 60% Complete
### Target State: 95% Production-Ready MVP
### Timeline: 14 weeks (3.5 months)
### Sprints: 7 x 2-week sprints

---

## Sprint Breakdown

```
Week 1-2:  Sprint 1 - Authentication Foundation
Week 3-4:  Sprint 2 - Security & Authorization
Week 5-6:  Sprint 3 - Audit & Monitoring
Week 7-8:  Sprint 4 - Testing Infrastructure
Week 9-10: Sprint 5 - Analytics & UI Polish
Week 11-12: Sprint 6 - Notifications & Alerts
Week 13-14: Sprint 7 - Final Integration & Deployment
```

---

## 🏃 Sprint 1: Authentication Foundation (Week 1-2)

### Goal
Implement complete user authentication system with secure session management.

### Backend Tasks (Engineer 1)

#### Task 1.1: Database Schema & Models (Day 1-2)
```sql
-- New tables to create

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'trader', -- admin, trader, viewer
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255),
    expires_at TIMESTAMP NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE email_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_users_email ON users(email);
```

**Go Models:**
```go
// internal/auth/models.go
package auth

import (
    "time"
    "github.com/google/uuid"
)

type User struct {
    ID            uuid.UUID  `json:"id" db:"id"`
    Email         string     `json:"email" db:"email"`
    PasswordHash  string     `json:"-" db:"password_hash"`
    FullName      string     `json:"full_name" db:"full_name"`
    Role          Role       `json:"role" db:"role"`
    IsActive      bool       `json:"is_active" db:"is_active"`
    EmailVerified bool       `json:"email_verified" db:"email_verified"`
    CreatedAt     time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
    LastLoginAt   *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
}

type Role string

const (
    RoleAdmin  Role = "admin"
    RoleTrader Role = "trader"
    RoleViewer Role = "viewer"
)

type Session struct {
    ID               uuid.UUID `json:"id" db:"id"`
    UserID           uuid.UUID `json:"user_id" db:"user_id"`
    TokenHash        string    `json:"-" db:"token_hash"`
    RefreshTokenHash string    `json:"-" db:"refresh_token_hash"`
    ExpiresAt        time.Time `json:"expires_at" db:"expires_at"`
    IPAddress        string    `json:"ip_address" db:"ip_address"`
    UserAgent        string    `json:"user_agent" db:"user_agent"`
    CreatedAt        time.Time `json:"created_at" db:"created_at"`
}
```

**Files to create:**
- `internal/auth/models.go`
- `internal/auth/repository.go`
- `internal/storage/migrations/004_create_auth_tables.sql`

**Estimated:** 1.5 days

---

#### Task 1.2: Authentication Service (Day 3-5)
```go
// internal/auth/service.go
package auth

import (
    "context"
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

type Service struct {
    repo         Repository
    jwtSecret    []byte
    tokenExpiry  time.Duration
    refreshExpiry time.Duration
}

type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    FullName string `json:"full_name" validate:"required"`
}

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
    User         *User  `json:"user"`
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*User, error) {
    // 1. Validate email uniqueness
    // 2. Hash password with bcrypt
    // 3. Create user
    // 4. Send verification email
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
    // 1. Find user by email
    // 2. Verify password
    // 3. Check if email verified
    // 4. Check if active
    // 5. Generate JWT tokens
    // 6. Create session
    // 7. Update last login
}

func (s *Service) Logout(ctx context.Context, token string) error {
    // 1. Parse token
    // 2. Delete session
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
    // 1. Validate refresh token
    // 2. Generate new access token
    // 3. Extend session
}

func (s *Service) VerifyToken(ctx context.Context, token string) (*User, error) {
    // 1. Parse JWT
    // 2. Validate expiry
    // 3. Fetch user
    // 4. Return user
}
```

**Dependencies to add:**
```go
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
go get github.com/go-playground/validator/v10
```

**Files to create:**
- `internal/auth/service.go`
- `internal/auth/jwt.go`
- `internal/auth/password.go`
- `internal/auth/validation.go`

**Estimated:** 3 days

---

#### Task 1.3: API Endpoints (Day 6-7)
```go
// internal/api/handlers/auth.go
package handlers

type AuthHandler struct {
    authService *auth.Service
}

// POST /api/v1/auth/register
func (h *AuthHandler) Register(c echo.Context) error {
    var req auth.RegisterRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(400, ErrorResponse{Message: "Invalid request"})
    }

    user, err := h.authService.Register(c.Request().Context(), req)
    if err != nil {
        return handleError(c, err)
    }

    return c.JSON(201, user)
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c echo.Context) error {}

// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c echo.Context) error {}

// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c echo.Context) error {}

// GET /api/v1/auth/me
func (h *AuthHandler) GetCurrentUser(c echo.Context) error {}

// POST /api/v1/auth/verify-email
func (h *AuthHandler) VerifyEmail(c echo.Context) error {}

// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c echo.Context) error {}

// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c echo.Context) error {}
```

**Middleware:**
```go
// internal/api/middleware/auth.go
package middleware

func AuthMiddleware(authService *auth.Service) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            token := extractToken(c.Request())
            if token == "" {
                return echo.NewHTTPError(401, "Missing token")
            }

            user, err := authService.VerifyToken(c.Request().Context(), token)
            if err != nil {
                return echo.NewHTTPError(401, "Invalid token")
            }

            c.Set("user", user)
            return next(c)
        }
    }
}

func RequireRole(roles ...auth.Role) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            user := c.Get("user").(*auth.User)

            for _, role := range roles {
                if user.Role == role {
                    return next(c)
                }
            }

            return echo.NewHTTPError(403, "Insufficient permissions")
        }
    }
}
```

**Files to create:**
- `internal/api/handlers/auth.go`
- `internal/api/middleware/auth.go`

**Estimated:** 2 days

---

### Frontend Tasks (Engineer 2)

#### Task 1.4: Authentication UI Pages (Day 1-7)

**Pages to create:**

1. **Login Page** (`web/src/pages/Login.tsx`)
```typescript
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import * as api from '../services/api';

export function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const response = await api.login({ email, password });
      localStorage.setItem('accessToken', response.data.access_token);
      localStorage.setItem('refreshToken', response.data.refresh_token);
      navigate('/');
    } catch (err) {
      setError('Invalid email or password');
    } finally {
      setLoading(false);
    }
  };

  return (
    // ... UI implementation as per design specs
  );
}
```

2. **Register Page** (`web/src/pages/Register.tsx`)
3. **Forgot Password Page** (`web/src/pages/ForgotPassword.tsx`)
4. **Reset Password Page** (`web/src/pages/ResetPassword.tsx`)
5. **Email Verification Page** (`web/src/pages/VerifyEmail.tsx`)

**API Service Updates:**
```typescript
// web/src/services/api.ts
export const register = (data: RegisterRequest) =>
  axios.post('/api/v1/auth/register', data);

export const login = (data: LoginRequest) =>
  axios.post('/api/v1/auth/login', data);

export const logout = () =>
  axios.post('/api/v1/auth/logout');

export const refreshToken = (refreshToken: string) =>
  axios.post('/api/v1/auth/refresh', { refresh_token: refreshToken });

export const getCurrentUser = () =>
  axios.get('/api/v1/auth/me');
```

**Axios Interceptor for Auth:**
```typescript
// web/src/services/axiosInterceptor.ts
axios.interceptors.request.use((config) => {
  const token = localStorage.getItem('accessToken');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

axios.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Try to refresh token
      const refreshToken = localStorage.getItem('refreshToken');
      if (refreshToken) {
        try {
          const response = await api.refreshToken(refreshToken);
          localStorage.setItem('accessToken', response.data.access_token);
          // Retry original request
          return axios(error.config);
        } catch {
          // Refresh failed, logout
          localStorage.clear();
          window.location.href = '/login';
        }
      }
    }
    return Promise.reject(error);
  }
);
```

**Files to create:**
- `web/src/pages/Login.tsx`
- `web/src/pages/Register.tsx`
- `web/src/pages/ForgotPassword.tsx`
- `web/src/pages/ResetPassword.tsx`
- `web/src/pages/VerifyEmail.tsx`
- `web/src/services/axiosInterceptor.ts`
- `web/src/contexts/AuthContext.tsx`
- `web/src/hooks/useAuth.ts`

**Estimated:** 5 days

---

### Sprint 1 Deliverables
- ✅ User registration with email verification
- ✅ Login/logout functionality
- ✅ JWT-based authentication
- ✅ Session management
- ✅ Password reset flow
- ✅ Auth middleware for API protection
- ✅ Frontend auth pages and routing
- ✅ Protected routes in React

**Sprint 1 Testing:**
- [ ] Manual testing of all auth flows
- [ ] Test password hashing security
- [ ] Test JWT expiration and refresh
- [ ] Test protected routes

---

## 🏃 Sprint 2: Security & Authorization (Week 3-4)

### Goal
Implement role-based access control, 2FA, API keys, and security hardening.

### Backend Tasks (Engineer 1)

#### Task 2.1: RBAC System (Day 1-3)

**Permission System:**
```go
// internal/auth/permissions.go
package auth

type Permission string

const (
    PermViewDashboard      Permission = "view:dashboard"
    PermExecuteTrades      Permission = "execute:trades"
    PermModifyStrategies   Permission = "modify:strategies"
    PermManageRisk         Permission = "manage:risk"
    PermManageUsers        Permission = "manage:users"
    PermViewAuditLogs      Permission = "view:audit_logs"
    PermManageAPIKeys      Permission = "manage:api_keys"
)

var RolePermissions = map[Role][]Permission{
    RoleAdmin: {
        PermViewDashboard,
        PermExecuteTrades,
        PermModifyStrategies,
        PermManageRisk,
        PermManageUsers,
        PermViewAuditLogs,
        PermManageAPIKeys,
    },
    RoleTrader: {
        PermViewDashboard,
        PermExecuteTrades,
        PermModifyStrategies,
        PermManageRisk,
    },
    RoleViewer: {
        PermViewDashboard,
    },
}

func (r Role) HasPermission(perm Permission) bool {
    perms, ok := RolePermissions[r]
    if !ok {
        return false
    }
    for _, p := range perms {
        if p == perm {
            return true
        }
    }
    return false
}
```

**Middleware:**
```go
// internal/api/middleware/rbac.go
func RequirePermission(perm auth.Permission) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            user := c.Get("user").(*auth.User)
            if !user.Role.HasPermission(perm) {
                return echo.NewHTTPError(403, "Insufficient permissions")
            }
            return next(c)
        }
    }
}
```

**Apply to routes:**
```go
// Protect trading endpoints
tradingGroup := api.Group("/trading", authMiddleware, RequirePermission(PermExecuteTrades))
tradingGroup.POST("/start", handlers.StartTrading)
tradingGroup.POST("/stop", handlers.StopTrading)

// Protect admin endpoints
adminGroup := api.Group("/admin", authMiddleware, RequireRole(RoleAdmin))
adminGroup.GET("/users", handlers.ListUsers)
```

**Estimated:** 2 days

---

#### Task 2.2: Two-Factor Authentication (Day 4-7)

**Database Schema:**
```sql
CREATE TABLE two_factor_auth (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    secret VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT false,
    backup_codes TEXT[], -- Array of hashed backup codes
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP
);
```

**Dependencies:**
```bash
go get github.com/pquerna/otp/totp
```

**2FA Service:**
```go
// internal/auth/totp.go
package auth

import (
    "github.com/pquerna/otp/totp"
)

type TOTPService struct {
    repo Repository
}

// Generate new TOTP secret for user
func (s *TOTPService) GenerateSecret(userID uuid.UUID, email string) (*TOTPSetup, error) {
    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      "ETH Trader",
        AccountName: email,
    })
    if err != nil {
        return nil, err
    }

    // Generate backup codes
    backupCodes := generateBackupCodes(10)

    return &TOTPSetup{
        Secret:      key.Secret(),
        QRCode:      key.URL(),
        BackupCodes: backupCodes,
    }, nil
}

// Verify TOTP code
func (s *TOTPService) Verify(userID uuid.UUID, code string) (bool, error) {
    tfa, err := s.repo.GetTwoFactorAuth(userID)
    if err != nil {
        return false, err
    }

    // Try TOTP code
    if totp.Validate(code, tfa.Secret) {
        return true, nil
    }

    // Try backup codes
    for _, hashedBackup := range tfa.BackupCodes {
        if bcrypt.CompareHashAndPassword([]byte(hashedBackup), []byte(code)) == nil {
            // Remove used backup code
            s.repo.RemoveBackupCode(userID, hashedBackup)
            return true, nil
        }
    }

    return false, nil
}
```

**API Endpoints:**
```go
// POST /api/v1/auth/2fa/setup
func (h *AuthHandler) Setup2FA(c echo.Context) error {}

// POST /api/v1/auth/2fa/enable
func (h *AuthHandler) Enable2FA(c echo.Context) error {}

// POST /api/v1/auth/2fa/disable
func (h *AuthHandler) Disable2FA(c echo.Context) error {}

// POST /api/v1/auth/2fa/verify
func (h *AuthHandler) Verify2FA(c echo.Context) error {}
```

**Modified Login Flow:**
```go
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
    // ... existing login logic

    // Check if 2FA enabled
    tfa, err := s.repo.GetTwoFactorAuth(user.ID)
    if err == nil && tfa.Enabled {
        // Return partial token requiring 2FA verification
        return &LoginResponse{
            RequiresTOTP: true,
            TempToken:    generateTempToken(user.ID),
        }, nil
    }

    // Continue normal login
}
```

**Estimated:** 3 days

---

#### Task 2.3: API Keys Management (Day 8-10)

**Database Schema:**
```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(20) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    secret_hash VARCHAR(255) NOT NULL,
    permissions TEXT[], -- Array of permissions
    ip_whitelist TEXT[], -- Array of IP addresses/CIDR
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix);
```

**API Key Service:**
```go
// internal/auth/apikey.go
package auth

type APIKeyService struct {
    repo Repository
}

type CreateAPIKeyRequest struct {
    Name        string       `json:"name"`
    Permissions []Permission `json:"permissions"`
    IPWhitelist []string     `json:"ip_whitelist"`
    ExpiresAt   *time.Time   `json:"expires_at"`
}

type APIKeyResponse struct {
    Key    string `json:"key"`     // Only returned once on creation
    Secret string `json:"secret"`  // Only returned once on creation
}

func (s *APIKeyService) Create(ctx context.Context, userID uuid.UUID, req CreateAPIKeyRequest) (*APIKeyResponse, error) {
    // 1. Generate random key and secret
    key := "prod_" + generateRandomString(32)
    secret := "sk_" + generateRandomString(48)

    // 2. Hash key and secret
    keyHash, _ := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
    secretHash, _ := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)

    // 3. Store in database
    apiKey := &APIKey{
        UserID:      userID,
        Name:        req.Name,
        KeyPrefix:   key[:12], // Store prefix for identification
        KeyHash:     string(keyHash),
        SecretHash:  string(secretHash),
        Permissions: req.Permissions,
        IPWhitelist: req.IPWhitelist,
        ExpiresAt:   req.ExpiresAt,
    }

    if err := s.repo.CreateAPIKey(apiKey); err != nil {
        return nil, err
    }

    return &APIKeyResponse{
        Key:    key,
        Secret: secret,
    }, nil
}

func (s *APIKeyService) Verify(ctx context.Context, key, secret, ipAddress string) (*User, error) {
    // 1. Find API key by prefix
    // 2. Verify key hash
    // 3. Verify secret hash
    // 4. Check expiration
    // 5. Check IP whitelist
    // 6. Update last_used_at
    // 7. Return user
}
```

**API Middleware:**
```go
// internal/api/middleware/apikey.go
func APIKeyAuth(apiKeyService *auth.APIKeyService) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            apiKey := c.Request().Header.Get("X-API-Key")
            apiSecret := c.Request().Header.Get("X-API-Secret")

            if apiKey == "" || apiSecret == "" {
                return echo.NewHTTPError(401, "Missing API credentials")
            }

            ipAddr := c.RealIP()
            user, err := apiKeyService.Verify(c.Request().Context(), apiKey, apiSecret, ipAddr)
            if err != nil {
                return echo.NewHTTPError(401, "Invalid API credentials")
            }

            c.Set("user", user)
            c.Set("auth_method", "api_key")
            return next(c)
        }
    }
}
```

**Endpoints:**
```go
// POST /api/v1/api-keys
func (h *AuthHandler) CreateAPIKey(c echo.Context) error {}

// GET /api/v1/api-keys
func (h *AuthHandler) ListAPIKeys(c echo.Context) error {}

// DELETE /api/v1/api-keys/:id
func (h *AuthHandler) RevokeAPIKey(c echo.Context) error {}

// PUT /api/v1/api-keys/:id
func (h *AuthHandler) UpdateAPIKey(c echo.Context) error {}
```

**Estimated:** 2.5 days

---

#### Task 2.4: Rate Limiting & Security Hardening (Day 11-14)

**Rate Limiter:**
```go
// internal/api/middleware/ratelimit.go
package middleware

import (
    "github.com/ulule/limiter/v3"
    "github.com/ulule/limiter/v3/drivers/store/memory"
    echolimiter "github.com/ulule/limiter/v3/drivers/middleware/echo"
)

func RateLimiter() echo.MiddlewareFunc {
    rate := limiter.Rate{
        Period: 1 * time.Minute,
        Limit:  60, // 60 requests per minute per IP
    }

    store := memory.NewStore()
    instance := limiter.New(store, rate)

    middleware := echolimiter.NewMiddleware(instance)
    return middleware
}

// Stricter rate limit for auth endpoints
func AuthRateLimiter() echo.MiddlewareFunc {
    rate := limiter.Rate{
        Period: 15 * time.Minute,
        Limit:  5, // 5 login attempts per 15 minutes
    }

    store := memory.NewStore()
    instance := limiter.New(store, rate)

    return echolimiter.NewMiddleware(instance)
}
```

**CORS Configuration:**
```go
// cmd/bot/main.go
e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: []string{os.Getenv("FRONTEND_URL")},
    AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
    AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType},
    AllowCredentials: true,
}))
```

**Request Signing (Optional Advanced):**
```go
// internal/api/middleware/signature.go
func VerifyRequestSignature() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Get signature from header
            signature := c.Request().Header.Get("X-Signature")
            timestamp := c.Request().Header.Get("X-Timestamp")

            // Verify timestamp is recent (prevent replay attacks)
            // Verify signature matches HMAC of request body + timestamp

            return next(c)
        }
    }
}
```

**Dependencies:**
```bash
go get github.com/ulule/limiter/v3
```

**Estimated:** 2.5 days

---

### Frontend Tasks (Engineer 2)

#### Task 2.5: 2FA UI (Day 1-5)

**Pages/Modals:**
1. **2FA Setup Modal** (`web/src/components/Setup2FAModal.tsx`)
2. **2FA Verification Page** (`web/src/pages/Verify2FA.tsx`)
3. **API Keys Management Page** (`web/src/pages/APIKeys.tsx`)
4. **User Management Page (Admin)** (`web/src/pages/admin/Users.tsx`)

**Estimated:** 4 days

---

#### Task 2.6: Protected Routes & Permissions (Day 6-10)

**Protected Route Component:**
```typescript
// web/src/components/ProtectedRoute.tsx
import { Navigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredPermission?: string;
  requiredRole?: string;
}

export function ProtectedRoute({
  children,
  requiredPermission,
  requiredRole
}: ProtectedRouteProps) {
  const { user, isAuthenticated, hasPermission, hasRole } = useAuth();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (requiredPermission && !hasPermission(requiredPermission)) {
    return <Navigate to="/unauthorized" replace />;
  }

  if (requiredRole && !hasRole(requiredRole)) {
    return <Navigate to="/unauthorized" replace />;
  }

  return <>{children}</>;
}
```

**Router Configuration:**
```typescript
// web/src/App.tsx
<Routes>
  <Route path="/login" element={<Login />} />
  <Route path="/register" element={<Register />} />

  <Route path="/" element={
    <ProtectedRoute>
      <Layout />
    </ProtectedRoute>
  }>
    <Route index element={<Dashboard />} />
    <Route path="strategies" element={
      <ProtectedRoute requiredPermission="modify:strategies">
        <Strategies />
      </ProtectedRoute>
    } />
    <Route path="admin/users" element={
      <ProtectedRoute requiredRole="admin">
        <Users />
      </ProtectedRoute>
    } />
  </Route>
</Routes>
```

**Estimated:** 3 days

---

### Sprint 2 Deliverables
- ✅ Role-based access control (RBAC)
- ✅ Two-factor authentication (TOTP)
- ✅ API key management
- ✅ Rate limiting on all endpoints
- ✅ CORS configuration
- ✅ Security hardening (HTTPS enforcement, etc.)
- ✅ Protected routes with permissions
- ✅ Admin user management UI

---

## 🏃 Sprint 3: Audit Logging & Monitoring (Week 5-6)

### Goal
Implement comprehensive audit logging and real-time monitoring system.

### Backend Tasks (Engineer 1)

#### Task 3.1: Audit Log System (Day 1-4)

**Database Schema:**
```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    event_type VARCHAR(100) NOT NULL,
    action VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100),
    entity_id VARCHAR(255),
    changes JSONB,
    metadata JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    session_id UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
```

**Audit Service:**
```go
// internal/audit/service.go
package audit

type EventType string

const (
    EventTypeAuth         EventType = "AUTH"
    EventTypeTrade        EventType = "TRADE"
    EventTypeConfig       EventType = "CONFIG"
    EventTypeRisk         EventType = "RISK"
    EventTypeSystem       EventType = "SYSTEM"
)

type AuditLog struct {
    ID         uuid.UUID              `json:"id"`
    UserID     *uuid.UUID             `json:"user_id,omitempty"`
    EventType  EventType              `json:"event_type"`
    Action     string                 `json:"action"`
    EntityType string                 `json:"entity_type,omitempty"`
    EntityID   string                 `json:"entity_id,omitempty"`
    Changes    map[string]interface{} `json:"changes,omitempty"`
    Metadata   map[string]interface{} `json:"metadata,omitempty"`
    IPAddress  string                 `json:"ip_address,omitempty"`
    UserAgent  string                 `json:"user_agent,omitempty"`
    SessionID  *uuid.UUID             `json:"session_id,omitempty"`
    CreatedAt  time.Time              `json:"created_at"`
}

type Service struct {
    repo Repository
}

func (s *Service) Log(ctx context.Context, log *AuditLog) error {
    // Async logging to not block main thread
    go func() {
        if err := s.repo.Create(log); err != nil {
            // Log error but don't fail
            logger.Error().Err(err).Msg("Failed to write audit log")
        }
    }()
    return nil
}

func (s *Service) Query(ctx context.Context, filter AuditFilter) ([]*AuditLog, error) {
    return s.repo.Query(filter)
}
```

**Audit Middleware:**
```go
// internal/api/middleware/audit.go
func AuditMiddleware(auditService *audit.Service) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            start := time.Now()

            // Call next handler
            err := next(c)

            // Log after response
            user, _ := c.Get("user").(*auth.User)

            log := &audit.AuditLog{
                EventType: determineEventType(c.Path()),
                Action:    c.Request().Method + " " + c.Path(),
                IPAddress: c.RealIP(),
                UserAgent: c.Request().UserAgent(),
                Metadata: map[string]interface{}{
                    "duration_ms": time.Since(start).Milliseconds(),
                    "status_code": c.Response().Status,
                },
            }

            if user != nil {
                log.UserID = &user.ID
            }

            auditService.Log(c.Request().Context(), log)

            return err
        }
    }
}
```

**Helper Functions:**
```go
// Use in handlers to log specific events
func logTradeExecution(auditService *audit.Service, user *auth.User, trade *execution.Trade) {
    auditService.Log(context.Background(), &audit.AuditLog{
        UserID:     &user.ID,
        EventType:  audit.EventTypeTrade,
        Action:     "TRADE_EXECUTED",
        EntityType: "trade",
        EntityID:   trade.ID.String(),
        Changes: map[string]interface{}{
            "symbol":      trade.Symbol,
            "side":        trade.Side,
            "quantity":    trade.Quantity,
            "entry_price": trade.EntryPrice,
        },
    })
}
```

**API Endpoints:**
```go
// GET /api/v1/audit-logs
func (h *AuditHandler) QueryLogs(c echo.Context) error {}

// GET /api/v1/audit-logs/:id
func (h *AuditHandler) GetLog(c echo.Context) error {}

// POST /api/v1/audit-logs/export
func (h *AuditHandler) ExportLogs(c echo.Context) error {}
```

**Estimated:** 3 days

---

#### Task 3.2: Notification Service (Day 5-8)

**Database Schema:**
```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    severity VARCHAR(20) DEFAULT 'info', -- info, warning, error, critical
    read BOOLEAN DEFAULT false,
    data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    email_enabled BOOLEAN DEFAULT true,
    sms_enabled BOOLEAN DEFAULT false,
    telegram_enabled BOOLEAN DEFAULT false,
    discord_enabled BOOLEAN DEFAULT false,
    in_app_enabled BOOLEAN DEFAULT true,
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    preferences JSONB, -- Per-notification-type preferences
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    conditions JSONB NOT NULL,
    channels TEXT[], -- email, sms, telegram, etc.
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_read ON notifications(read);
```

**Notification Service:**
```go
// internal/notifications/service.go
package notifications

type NotificationType string

const (
    NotifTypeTradeExecuted   NotificationType = "trade_executed"
    NotifTypeRiskAlert       NotificationType = "risk_alert"
    NotifTypeSystemError     NotificationType = "system_error"
    NotifTypeStrategySignal  NotificationType = "strategy_signal"
)

type Severity string

const (
    SeverityInfo     Severity = "info"
    SeverityWarning  Severity = "warning"
    SeverityError    Severity = "error"
    SeverityCritical Severity = "critical"
)

type Notification struct {
    ID        uuid.UUID        `json:"id"`
    UserID    uuid.UUID        `json:"user_id"`
    Type      NotificationType `json:"type"`
    Title     string           `json:"title"`
    Message   string           `json:"message"`
    Severity  Severity         `json:"severity"`
    Read      bool             `json:"read"`
    Data      interface{}      `json:"data,omitempty"`
    CreatedAt time.Time        `json:"created_at"`
}

type Service struct {
    repo           Repository
    emailSender    EmailSender
    smsSender      SMSSender
    telegramSender TelegramSender
    discordSender  DiscordSender
    wsHub          *websocket.Hub // For real-time in-app notifications
}

func (s *Service) Send(ctx context.Context, notif *Notification) error {
    // 1. Save to database
    if err := s.repo.Create(notif); err != nil {
        return err
    }

    // 2. Get user preferences
    prefs, err := s.repo.GetPreferences(notif.UserID)
    if err != nil {
        return err
    }

    // 3. Check quiet hours
    if s.isQuietHours(prefs) && notif.Severity != SeverityCritical {
        return nil // Don't send during quiet hours unless critical
    }

    // 4. Send through enabled channels
    if prefs.EmailEnabled {
        go s.emailSender.Send(notif)
    }

    if prefs.TelegramEnabled {
        go s.telegramSender.Send(notif)
    }

    if prefs.InAppEnabled {
        // Send through WebSocket
        s.wsHub.BroadcastToUser(notif.UserID, websocket.Message{
            Type: "NOTIFICATION",
            Data: notif,
        })
    }

    return nil
}
```

**Integration Points:**
```go
// Example: Send notification on trade execution
func (e *LiveExecutor) ExecuteTrade(trade *Trade) error {
    // ... execute trade logic

    // Send notification
    e.notificationService.Send(context.Background(), &notifications.Notification{
        UserID:   e.userID,
        Type:     notifications.NotifTypeTradeExecuted,
        Title:    "Trade Executed",
        Message:  fmt.Sprintf("%s position opened at $%.2f", trade.Side, trade.EntryPrice),
        Severity: notifications.SeverityInfo,
        Data: map[string]interface{}{
            "trade_id": trade.ID,
            "symbol":   trade.Symbol,
            "side":     trade.Side,
        },
    })

    return nil
}
```

**Email/Telegram/Discord Senders:**
```go
// internal/notifications/email.go
type EmailSender interface {
    Send(notif *Notification) error
}

type SMTPEmailSender struct {
    host     string
    port     int
    username string
    password string
}

// Use net/smtp or third-party like sendgrid

// internal/notifications/telegram.go
type TelegramSender struct {
    botToken string
    chatID   string
}

// Use telegram bot API

// internal/notifications/discord.go
type DiscordSender struct {
    webhookURL string
}

// Use discord webhook
```

**Dependencies:**
```bash
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
```

**API Endpoints:**
```go
// GET /api/v1/notifications
func (h *NotificationHandler) ListNotifications(c echo.Context) error {}

// PUT /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c echo.Context) error {}

// PUT /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c echo.Context) error {}

// GET /api/v1/notifications/preferences
func (h *NotificationHandler) GetPreferences(c echo.Context) error {}

// PUT /api/v1/notifications/preferences
func (h *NotificationHandler) UpdatePreferences(c echo.Context) error {}

// POST /api/v1/alerts
func (h *NotificationHandler) CreateAlert(c echo.Context) error {}

// GET /api/v1/alerts
func (h *NotificationHandler) ListAlerts(c echo.Context) error {}
```

**Estimated:** 3 days

---

### Frontend Tasks (Engineer 2)

#### Task 3.3: Audit Log Viewer UI (Day 1-4)

**Files:**
- `web/src/pages/AuditLogs.tsx`
- `web/src/components/AuditLogTable.tsx`
- `web/src/components/AuditLogFilters.tsx`

**Estimated:** 3 days

---

#### Task 3.4: Notification Center UI (Day 5-10)

**Components:**
- `web/src/components/NotificationBell.tsx` (header dropdown)
- `web/src/pages/Notifications.tsx` (full page)
- `web/src/components/NotificationSettings.tsx`
- `web/src/components/Toast.tsx` (for real-time toasts)

**WebSocket Integration:**
```typescript
// web/src/hooks/useNotifications.ts
export function useNotifications() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const { wsMessage } = useWebSocket();

  useEffect(() => {
    if (wsMessage?.type === 'NOTIFICATION') {
      const notif = wsMessage.data as Notification;
      setNotifications(prev => [notif, ...prev]);

      // Show toast
      toast.info(notif.title, { description: notif.message });
    }
  }, [wsMessage]);

  return { notifications };
}
```

**Estimated:** 4 days

---

### Sprint 3 Deliverables
- ✅ Comprehensive audit logging system
- ✅ Audit log viewer with search/filter
- ✅ Notification service with multiple channels
- ✅ In-app notification center
- ✅ Email/Telegram/Discord integration
- ✅ Custom alert rules
- ✅ Real-time toast notifications

---

## 🏃 Sprint 4: Testing Infrastructure (Week 7-8)

### Goal
Build comprehensive testing infrastructure with 80%+ code coverage.

### Backend Testing (Engineer 1)

#### Task 4.1: Testing Setup & Strategy Repository Tests (Day 1-3)

**Setup:**
```bash
# Add testing dependencies
go get github.com/stretchr/testify
go get github.com/DATA-DOG/go-sqlmock
```

**Test Structure:**
```
internal/
├── strategy/
│   ├── manager.go
│   ├── manager_test.go
│   ├── trend_following.go
│   ├── trend_following_test.go
│   ├── mean_reversion.go
│   ├── mean_reversion_test.go
│   └── ...
```

**Example Strategy Test:**
```go
// internal/strategy/trend_following_test.go
package strategy

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestTrendFollowingStrategy_Evaluate(t *testing.T) {
    tests := []struct {
        name           string
        indicators     map[string]interface{}
        expectedSignal SignalDirection
        expectedConf   float64
    }{
        {
            name: "Strong bullish trend",
            indicators: map[string]interface{}{
                "adx":      28.0,
                "ma_10":    3200.0,
                "ma_20":    3150.0,
                "ma_50":    3100.0,
                "macd":     5.0,
                "signal":   3.0,
            },
            expectedSignal: SignalLong,
            expectedConf:   0.8,
        },
        {
            name: "Weak trend - no signal",
            indicators: map[string]interface{}{
                "adx":      22.0, // Below threshold
                "ma_10":    3200.0,
                "ma_20":    3150.0,
            },
            expectedSignal: SignalNone,
            expectedConf:   0.0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            strategy := NewTrendFollowingStrategy()
            signal, conf := strategy.Evaluate(tt.indicators)

            assert.Equal(t, tt.expectedSignal, signal)
            assert.InDelta(t, tt.expectedConf, conf, 0.1)
        })
    }
}
```

**Indicator Tests:**
```go
// internal/indicators/rsi_test.go
func TestRSI_Calculate(t *testing.T) {
    prices := []float64{
        44.0, 44.5, 45.0, 45.5, 45.0, 44.5, 44.0,
        43.5, 44.0, 44.5, 45.0, 45.5, 46.0, 46.5,
    }

    rsi := NewRSI(14)

    for _, price := range prices {
        rsi.Update(price)
    }

    value := rsi.Value()

    // RSI should be between 0 and 100
    assert.GreaterOrEqual(t, value, 0.0)
    assert.LessOrEqual(t, value, 100.0)

    // For uptrend, RSI should be > 50
    assert.Greater(t, value, 50.0)
}
```

**Estimated:** 2.5 days

---

#### Task 4.2: API Integration Tests (Day 4-6)

**Test Server Setup:**
```go
// internal/api/server_test.go
package api

import (
    "net/http/httptest"
    "testing"

    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/suite"
)

type APITestSuite struct {
    suite.Suite
    server *Server
    router *echo.Echo
}

func (s *APITestSuite) SetupSuite() {
    // Create test database
    // Initialize services
    // Create server
}

func (s *APITestSuite) TearDownSuite() {
    // Clean up database
}

func (s *APITestSuite) TestLogin_Success() {
    payload := `{"email":"test@example.com","password":"password123"}`
    req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(payload))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    s.router.ServeHTTP(rec, req)

    s.Equal(200, rec.Code)

    var response map[string]interface{}
    json.Unmarshal(rec.Body.Bytes(), &response)

    s.NotEmpty(response["access_token"])
    s.NotEmpty(response["refresh_token"])
}

func (s *APITestSuite) TestProtectedEndpoint_WithoutAuth() {
    req := httptest.NewRequest("GET", "/api/v1/dashboard", nil)
    rec := httptest.NewRecorder()

    s.router.ServeHTTP(rec, req)

    s.Equal(401, rec.Code)
}

func TestAPITestSuite(t *testing.T) {
    suite.Run(t, new(APITestSuite))
}
```

**Test Coverage Goals:**
- Auth endpoints: 90%+
- Trading endpoints: 85%+
- Risk endpoints: 85%+
- Settings endpoints: 80%+

**Estimated:** 2.5 days

---

#### Task 4.3: Mock Binance Client & Risk Manager Tests (Day 7-10)

**Mock Binance Client:**
```go
// internal/binance/mock_client.go
package binance

type MockClient struct {
    mock.Mock
}

func (m *MockClient) PlaceOrder(req *OrderRequest) (*Order, error) {
    args := m.Called(req)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Order), args.Error(1)
}

func (m *MockClient) CancelOrder(orderID string) error {
    args := m.Called(orderID)
    return args.Error(0)
}

// Use in tests:
func TestLiveExecutor_ExecuteTrade(t *testing.T) {
    mockBinance := new(MockClient)
    executor := NewLiveExecutor(mockBinance, riskManager)

    // Setup expectation
    mockBinance.On("PlaceOrder", mock.Anything).Return(&Order{
        ID:     "order_123",
        Status: "FILLED",
    }, nil)

    err := executor.ExecuteTrade(trade)

    assert.NoError(t, err)
    mockBinance.AssertExpectations(t)
}
```

**Risk Manager Tests:**
```go
// internal/risk/manager_test.go
func TestRiskManager_CheckPosition_ExceedsMaxPosition(t *testing.T) {
    rm := NewRiskManager(&Config{
        MaxPositionSize: 0.1, // 10%
    })

    rm.accountBalance = 100000

    trade := &Trade{
        Quantity:   0.5,
        EntryPrice: 3000,
    }

    err := rm.ValidatePosition(trade)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "exceeds max position size")
}

func TestRiskManager_CheckDailyLoss(t *testing.T) {
    rm := NewRiskManager(&Config{
        MaxDailyLoss: 0.03, // 3%
    })

    rm.dailyPnL = -3500 // -3.5%
    rm.accountBalance = 100000

    canTrade := rm.CanOpenNewPosition()

    assert.False(t, canTrade)
}
```

**Estimated:** 3 days

---

#### Task 4.4: CI/CD Pipeline Setup (Day 11-14)

**GitHub Actions Workflow:**
```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  backend:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.22'

    - name: Run tests
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

    - name: Check coverage threshold
      run: |
        coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        if (( $(echo "$coverage < 80" | bc -l) )); then
          echo "Coverage $coverage% is below 80%"
          exit 1
        fi

  frontend:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Set up Node
      uses: actions/setup-node@v3
      with:
        node-version: '20'

    - name: Install dependencies
      run: cd web && npm ci

    - name: Run tests
      run: cd web && npm test -- --coverage

    - name: Build
      run: cd web && npm run build

  lint:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
```

**Makefile Updates:**
```makefile
test:
	go test -v -race ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

ci: lint test-coverage
```

**Estimated:** 2 days

---

### Frontend Testing (Engineer 2)

#### Task 4.5: Component Unit Tests (Day 1-7)

**Setup:**
```bash
cd web
npm install --save-dev @testing-library/react @testing-library/jest-dom @testing-library/user-event vitest
```

**Vitest Config:**
```typescript
// web/vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/tests/setup.ts',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
      exclude: ['node_modules/', 'src/tests/'],
    },
  },
});
```

**Example Component Test:**
```typescript
// web/src/components/__tests__/LoginForm.test.tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { LoginForm } from '../LoginForm';
import * as api from '../../services/api';

jest.mock('../../services/api');

describe('LoginForm', () => {
  it('renders email and password fields', () => {
    render(<LoginForm />);

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
  });

  it('shows error on failed login', async () => {
    (api.login as jest.Mock).mockRejectedValue(new Error('Invalid credentials'));

    render(<LoginForm />);

    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: 'test@example.com' },
    });
    fireEvent.change(screen.getByLabelText(/password/i), {
      target: { value: 'wrong' },
    });

    fireEvent.click(screen.getByRole('button', { name: /login/i }));

    await waitFor(() => {
      expect(screen.getByText(/invalid credentials/i)).toBeInTheDocument();
    });
  });

  it('redirects on successful login', async () => {
    (api.login as jest.Mock).mockResolvedValue({
      data: { access_token: 'token123', refresh_token: 'refresh123' },
    });

    const mockNavigate = jest.fn();
    jest.mock('react-router-dom', () => ({
      ...jest.requireActual('react-router-dom'),
      useNavigate: () => mockNavigate,
    }));

    render(<LoginForm />);

    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: 'test@example.com' },
    });
    fireEvent.change(screen.getByLabelText(/password/i), {
      target: { value: 'password123' },
    });

    fireEvent.click(screen.getByRole('button', { name: /login/i }));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/');
    });
  });
});
```

**Test Coverage Goals:**
- Components: 80%+
- Hooks: 90%+
- Utils: 90%+

**Estimated:** 5 days

---

#### Task 4.6: E2E Tests with Playwright (Day 8-14)

**Setup:**
```bash
npm install --save-dev @playwright/test
```

**E2E Tests:**
```typescript
// web/tests/e2e/auth.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
  test('complete login flow', async ({ page }) => {
    await page.goto('http://localhost:3000/login');

    await page.fill('input[name="email"]', 'test@example.com');
    await page.fill('input[name="password"]', 'password123');

    await page.click('button[type="submit"]');

    await expect(page).toHaveURL('http://localhost:3000/');
    await expect(page.locator('text=Dashboard')).toBeVisible();
  });

  test('2FA flow', async ({ page }) => {
    await page.goto('http://localhost:3000/login');

    await page.fill('input[name="email"]', '2fa@example.com');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');

    // Should redirect to 2FA page
    await expect(page).toHaveURL(/.*verify-2fa.*/);

    await page.fill('input[name="code"]', '123456');
    await page.click('button[type="submit"]');

    await expect(page).toHaveURL('http://localhost:3000/');
  });
});

test.describe('Trading Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Login first
    await page.goto('http://localhost:3000/login');
    await page.fill('input[name="email"]', 'trader@example.com');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('http://localhost:3000/');
  });

  test('start and stop bot', async ({ page }) => {
    const startButton = page.locator('button:has-text("Start")');
    await startButton.click();

    await expect(page.locator('button:has-text("Stop")')).toBeVisible();
    await expect(page.locator('text=Running')).toBeVisible();

    const stopButton = page.locator('button:has-text("Stop")');
    await stopButton.click();

    await expect(page.locator('button:has-text("Start")')).toBeVisible();
  });
});
```

**Estimated:** 4 days

---

### Sprint 4 Deliverables
- ✅ Backend unit tests (80%+ coverage)
- ✅ API integration tests
- ✅ Mock Binance client for testing
- ✅ Frontend component tests
- ✅ E2E tests with Playwright
- ✅ CI/CD pipeline with automated testing
- ✅ Coverage reporting

---

## 🏃 Sprint 5: Analytics & UI Polish (Week 9-10)

### Goal
Complete analytics calculations, polish UI/UX, and improve user experience.

### Backend Tasks (Engineer 1)

#### Task 5.1: Complete Analytics Calculations (Day 1-5)

**Sharpe/Sortino/Calmar Ratios:**
```go
// internal/analytics/metrics.go
package analytics

import (
    "math"
)

type PerformanceMetrics struct {
    SharpeRatio  float64 `json:"sharpe_ratio"`
    SortinoRatio float64 `json:"sortino_ratio"`
    CalmarRatio  float64 `json:"calmar_ratio"`
    ProfitFactor float64 `json:"profit_factor"`
    MaxDrawdown  float64 `json:"max_drawdown"`
    WinRate      float64 `json:"win_rate"`
    // ... other metrics
}

type Service struct {
    repo Repository
}

func (s *Service) CalculateSharpeRatio(returns []float64, riskFreeRate float64) float64 {
    if len(returns) == 0 {
        return 0
    }

    // Calculate excess returns
    excessReturns := make([]float64, len(returns))
    for i, r := range returns {
        excessReturns[i] = r - riskFreeRate
    }

    // Mean excess return
    meanExcess := mean(excessReturns)

    // Standard deviation of excess returns
    stdDev := standardDeviation(excessReturns)

    if stdDev == 0 {
        return 0
    }

    // Sharpe Ratio = (Mean Excess Return) / (Std Dev of Excess Returns)
    // Annualize: multiply by sqrt(252) for daily returns
    return (meanExcess / stdDev) * math.Sqrt(252)
}

func (s *Service) CalculateSortinoRatio(returns []float64, riskFreeRate float64, targetReturn float64) float64 {
    if len(returns) == 0 {
        return 0
    }

    // Calculate excess returns
    excessReturns := make([]float64, len(returns))
    for i, r := range returns {
        excessReturns[i] = r - targetReturn
    }

    meanExcess := mean(excessReturns)

    // Downside deviation (only negative returns)
    downsideReturns := []float64{}
    for _, r := range returns {
        if r < targetReturn {
            downsideReturns = append(downsideReturns, r-targetReturn)
        }
    }

    if len(downsideReturns) == 0 {
        return 0
    }

    downsideDeviation := standardDeviation(downsideReturns)

    if downsideDeviation == 0 {
        return 0
    }

    return (meanExcess / downsideDeviation) * math.Sqrt(252)
}

func (s *Service) CalculateCalmarRatio(returns []float64, maxDrawdown float64) float64 {
    if maxDrawdown == 0 {
        return 0
    }

    // Annualized return
    annualizedReturn := mean(returns) * 252

    // Calmar = Annualized Return / Max Drawdown
    return annualizedReturn / math.Abs(maxDrawdown)
}

func (s *Service) CalculateMaxDrawdown(equityCurve []float64) float64 {
    if len(equityCurve) == 0 {
        return 0
    }

    var maxDrawdown float64
    peak := equityCurve[0]

    for _, value := range equityCurve {
        if value > peak {
            peak = value
        }

        drawdown := (peak - value) / peak
        if drawdown > maxDrawdown {
            maxDrawdown = drawdown
        }
    }

    return maxDrawdown
}

func mean(values []float64) float64 {
    if len(values) == 0 {
        return 0
    }

    sum := 0.0
    for _, v := range values {
        sum += v
    }
    return sum / float64(len(values))
}

func standardDeviation(values []float64) float64 {
    if len(values) == 0 {
        return 0
    }

    m := mean(values)
    variance := 0.0

    for _, v := range values {
        variance += math.Pow(v-m, 2)
    }

    variance /= float64(len(values))
    return math.Sqrt(variance)
}
```

**API Endpoints:**
```go
// GET /api/v1/analytics/performance
func (h *AnalyticsHandler) GetPerformanceMetrics(c echo.Context) error {
    startDate := c.QueryParam("start_date")
    endDate := c.QueryParam("end_date")

    metrics, err := h.analyticsService.CalculateMetrics(startDate, endDate)
    if err != nil {
        return handleError(c, err)
    }

    return c.JSON(200, metrics)
}

// GET /api/v1/analytics/returns
func (h *AnalyticsHandler) GetReturns(c echo.Context) error {}

// GET /api/v1/analytics/drawdown
func (h *AnalyticsHandler) GetDrawdownHistory(c echo.Context) error {}
```

**Estimated:** 3.5 days

---

#### Task 5.2: Data Export Service (Day 6-10)

**Export Service:**
```go
// internal/export/service.go
package export

import (
    "encoding/csv"
    "encoding/json"
)

type Format string

const (
    FormatCSV  Format = "csv"
    FormatJSON Format = "json"
    FormatExcel Format = "excel"
)

type Service struct {
    tradeRepo    *storage.TradeRepository
    positionRepo *storage.PositionRepository
}

func (s *Service) ExportTrades(startDate, endDate time.Time, format Format) ([]byte, error) {
    trades, err := s.tradeRepo.GetByDateRange(startDate, endDate)
    if err != nil {
        return nil, err
    }

    switch format {
    case FormatCSV:
        return s.exportTradesCSV(trades)
    case FormatJSON:
        return json.Marshal(trades)
    default:
        return nil, errors.New("unsupported format")
    }
}

func (s *Service) exportTradesCSV(trades []*Trade) ([]byte, error) {
    var buf bytes.Buffer
    writer := csv.NewWriter(&buf)

    // Header
    writer.Write([]string{
        "ID", "Symbol", "Side", "Entry Price", "Exit Price",
        "Quantity", "PnL", "Commission", "Entry Time", "Exit Time",
        "Strategy", "Duration",
    })

    // Rows
    for _, trade := range trades {
        writer.Write([]string{
            trade.ID.String(),
            trade.Symbol,
            string(trade.Side),
            fmt.Sprintf("%.2f", trade.EntryPrice),
            fmt.Sprintf("%.2f", trade.ExitPrice),
            fmt.Sprintf("%.4f", trade.Quantity),
            fmt.Sprintf("%.2f", trade.PnL),
            fmt.Sprintf("%.2f", trade.Commission),
            trade.EntryTime.Format(time.RFC3339),
            trade.ExitTime.Format(time.RFC3339),
            trade.Strategy,
            trade.ExitTime.Sub(trade.EntryTime).String(),
        })
    }

    writer.Flush()
    return buf.Bytes(), nil
}
```

**API Endpoints:**
```go
// POST /api/v1/export/trades
func (h *ExportHandler) ExportTrades(c echo.Context) error {
    var req ExportRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(400, ErrorResponse{Message: "Invalid request"})
    }

    data, err := h.exportService.ExportTrades(req.StartDate, req.EndDate, req.Format)
    if err != nil {
        return handleError(c, err)
    }

    c.Response().Header().Set("Content-Disposition", "attachment; filename=trades.csv")
    c.Response().Header().Set("Content-Type", "text/csv")

    return c.Blob(200, "text/csv", data)
}
```

**Estimated:** 2 days

---

### Frontend Tasks (Engineer 2)

#### Task 5.3: Enhanced Analytics Page (Day 1-5)

**Components:**
- Advanced metrics display (Sharpe, Sortino, Calmar)
- Trade distribution charts
- Performance by time of day
- Win/loss streak analysis
- Risk metrics visualization (VaR)

**Files:**
- `web/src/pages/Analytics.tsx` (major update)
- `web/src/components/MetricCard.tsx`
- `web/src/components/TradeDistributionChart.tsx`
- `web/src/components/PerformanceHeatmap.tsx`

**Estimated:** 4 days

---

#### Task 5.4: Dashboard Enhancements (Day 6-10)

**Features:**
- Customizable widget layout (react-grid-layout)
- Chart drawing tools
- Position size calculator widget
- Theme toggle (dark/light)
- Keyboard shortcuts
- Loading skeletons
- Empty states

**Files:**
- `web/src/components/widgets/` (new directory)
- `web/src/components/PositionCalculator.tsx`
- `web/src/components/ChartDrawingTools.tsx`
- `web/src/contexts/ThemeContext.tsx`

**Estimated:** 4 days

---

### Sprint 5 Deliverables
- ✅ Complete Sharpe/Sortino/Calmar calculations
- ✅ Data export (CSV, JSON)
- ✅ Enhanced analytics page
- ✅ Customizable dashboard widgets
- ✅ Chart drawing tools
- ✅ Theme support
- ✅ Improved loading states
- ✅ Keyboard shortcuts

---

## 🏃 Sprint 6: Final Polish & Documentation (Week 11-12)

*(Abbreviated - similar structure to above sprints)*

### Goals
- Mobile responsiveness improvements
- Settings page reorganization
- Comprehensive documentation
- Performance optimization
- Accessibility improvements

### Key Tasks
- Mobile bottom navigation
- Settings tabs
- User guide documentation
- API documentation
- Performance audits
- Accessibility audit (WCAG 2.1 AA)

**Estimated:** 2 weeks

---

## 🏃 Sprint 7: Integration & Deployment (Week 13-14)

### Goals
- Final integration testing
- Production deployment setup
- Database migration to PostgreSQL
- Docker containerization
- HTTPS/SSL setup
- Monitoring setup

### Backend (Engineer 1)

#### Task 7.1: PostgreSQL Migration (Day 1-3)
- Migrate from SQLite to PostgreSQL
- Connection pooling
- Database migrations with golang-migrate

#### Task 7.2: Docker Setup (Day 4-7)
```dockerfile
# Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bot ./cmd/bot

FROM alpine:latest
WORKDIR /
COPY --from=builder /bot /bot

EXPOSE 8080

CMD ["/bot"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: eth_trading
      POSTGRES_USER: trader
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  backend:
    build: .
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://trader:${DB_PASSWORD}@postgres:5432/eth_trading
      BINANCE_API_KEY: ${BINANCE_API_KEY}
      BINANCE_API_SECRET: ${BINANCE_API_SECRET}
      JWT_SECRET: ${JWT_SECRET}
    depends_on:
      - postgres

  frontend:
    build: ./web
    ports:
      - "3000:3000"
    environment:
      VITE_API_URL: http://backend:8080

volumes:
  postgres_data:
```

#### Task 7.3: Production Deployment (Day 8-10)
- Set up cloud hosting (AWS/GCP/DigitalOcean)
- HTTPS with Let's Encrypt
- Environment variable management
- Database backups
- Monitoring (Prometheus/Grafana or similar)

### Frontend (Engineer 2)

#### Task 7.4: Production Build Optimization (Day 1-5)
- Code splitting
- Lazy loading
- Bundle size optimization
- PWA configuration
- Error tracking (Sentry)

#### Task 7.5: Final E2E Testing (Day 6-10)
- Complete E2E test suite
- Performance testing
- Security testing
- Load testing

---

## 📦 Deliverables Summary

### By End of Week 14:

**Backend:**
- ✅ Complete authentication system with 2FA
- ✅ Role-based access control
- ✅ API key management
- ✅ Comprehensive audit logging
- ✅ Multi-channel notifications
- ✅ Complete analytics engine
- ✅ Data export functionality
- ✅ 80%+ test coverage
- ✅ PostgreSQL database
- ✅ Docker containerization
- ✅ Production deployment

**Frontend:**
- ✅ Authentication UI (login, register, 2FA)
- ✅ Protected routes with permissions
- ✅ Notification center
- ✅ Audit log viewer
- ✅ Enhanced analytics page
- ✅ Customizable dashboard
- ✅ User/API key management
- ✅ Theme support
- ✅ Mobile responsive
- ✅ Keyboard shortcuts
- ✅ 80%+ test coverage
- ✅ Production build optimized

**Infrastructure:**
- ✅ CI/CD pipeline
- ✅ Automated testing
- ✅ HTTPS/SSL
- ✅ Database backups
- ✅ Monitoring

---

## 🎯 Post-MVP Roadmap (Week 15+)

### Phase 2: Advanced Features (Week 15-20)
- Multi-asset portfolio support
- Advanced order types (TWAP, iceberg, etc.)
- Visual strategy builder
- ML-based regime detection
- Strategy optimization tools

### Phase 3: Scale & Expansion (Week 21+)
- Mobile applications (iOS/Android)
- Additional exchange integrations
- Strategy marketplace
- Social trading features
- Advanced backtesting (walk-forward, Monte Carlo)

---

## 📋 Success Metrics

### Technical Metrics
- Test coverage: ≥80%
- API response time: <100ms (p95)
- WebSocket latency: <50ms
- Uptime: ≥99.5%
- Zero critical security vulnerabilities

### User Experience Metrics
- Login success rate: >95%
- Page load time: <2s
- Dashboard interactive time: <1s
- Mobile usability score: >80

### Business Metrics
- User onboarding completion: >70%
- Daily active users retention: >60%
- Zero data loss incidents
- Zero unauthorized access incidents

---

## 🚨 Risk Mitigation

### Technical Risks
1. **Database Migration Issues**
   - Mitigation: Thorough testing, backup strategy, rollback plan

2. **Security Vulnerabilities**
   - Mitigation: Regular audits, penetration testing, bug bounty program

3. **Performance Bottlenecks**
   - Mitigation: Load testing, profiling, caching strategy

### Schedule Risks
1. **Scope Creep**
   - Mitigation: Strict MVP definition, change control process

2. **Resource Constraints**
   - Mitigation: Prioritization matrix, phased delivery

3. **Third-Party Dependencies**
   - Mitigation: Vendor evaluation, fallback plans

---

## 📞 Support & Maintenance Plan

### Week 15+: Ongoing Maintenance
- Weekly security patches
- Monthly feature releases
- Quarterly major updates
- 24/7 uptime monitoring
- User support system

---

**End of Implementation Roadmap**
