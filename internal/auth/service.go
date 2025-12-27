package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/eth-trading/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	// AccessTokenDuration is the duration for access tokens (15 minutes)
	AccessTokenDuration = 15 * time.Minute
	// RefreshTokenDuration is the duration for refresh tokens (7 days)
	RefreshTokenDuration = 7 * 24 * time.Hour
	// BcryptCost is the cost factor for bcrypt hashing
	BcryptCost = 12
)

// Service provides authentication services
type Service struct {
	jwtSecret          []byte
	userRepo           UserRepository
	sessionRepo        SessionRepository
	tradingAccountRepo TradingAccountRepository
	tokenExpiry        time.Duration
	refreshTokenExpiry time.Duration
}

// UserRepository defines methods for user data access
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uuid.UUID) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	UpdateLastLogin(userID uuid.UUID) error
	EmailExists(email string) (bool, error)
}

// SessionRepository defines methods for session data access
type SessionRepository interface {
	Create(session *models.Session) error
	GetByRefreshToken(token string) (*models.Session, error)
	DeleteByUserID(userID uuid.UUID) error
	DeleteExpired() error
	Delete(id uuid.UUID) error
}

// TradingAccountRepository defines methods for trading account data access
type TradingAccountRepository interface {
	Create(account *models.TradingAccount) error
	GetByID(id uuid.UUID) (*models.TradingAccount, error)
	GetByUserID(userID uuid.UUID) ([]*models.TradingAccount, error)
	Update(account *models.TradingAccount) error
}

// Config holds authentication service configuration
type Config struct {
	JWTSecret          string
	TokenExpiry        time.Duration
	RefreshTokenExpiry time.Duration
}

// NewService creates a new authentication service
func NewService(cfg *Config, userRepo UserRepository, sessionRepo SessionRepository, tradingAccountRepo TradingAccountRepository) *Service {
	tokenExpiry := AccessTokenDuration
	if cfg.TokenExpiry > 0 {
		tokenExpiry = cfg.TokenExpiry
	}

	refreshTokenExpiry := RefreshTokenDuration
	if cfg.RefreshTokenExpiry > 0 {
		refreshTokenExpiry = cfg.RefreshTokenExpiry
	}

	return &Service{
		jwtSecret:          []byte(cfg.JWTSecret),
		userRepo:           userRepo,
		sessionRepo:        sessionRepo,
		tradingAccountRepo: tradingAccountRepo,
		tokenExpiry:        tokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// Register registers a new user with initial trading account
func (s *Service) Register(req *models.RegisterRequest) (*models.User, *models.TradingAccount, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if user already exists
	exists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("check email existence: %w", err)
	}
	if exists {
		return nil, nil, models.ErrUserAlreadyExists
	}

	// Hash password
	passwordHash, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: passwordHash,
		FullName:     req.FullName,
		Role:         models.RoleTrader, // Default role
		IsActive:     true,
		IsEmailVerified: false, // Require email verification in production
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, nil, fmt.Errorf("create user: %w", err)
	}

	// Create initial trading account
	account := &models.TradingAccount{
		ID:          uuid.New(),
		UserID:      user.ID,
		AccountType: req.AccountType,
		AccountName: req.AccountName,
		TradingSymbol: "ETHUSDT",
		TradingMode: models.TradingModePaper, // Start with paper trading
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set demo or live account fields
	if req.AccountType == models.AccountTypeDemo {
		account.DemoInitialCapital = req.DemoInitialCapital
		account.DemoCurrentBalance = req.DemoInitialCapital
	} else if req.AccountType == models.AccountTypeLive {
		account.BinanceAPIKey = req.BinanceAPIKey
		account.BinanceTestnet = req.BinanceTestnet
		// TODO: Encrypt Binance secret key before storing
		// For now, we'll leave it nil and handle encryption separately
	}

	if err := s.tradingAccountRepo.Create(account); err != nil {
		return nil, nil, fmt.Errorf("create trading account: %w", err)
	}

	return user, account, nil
}

// Login authenticates a user and returns tokens
func (s *Service) Login(email, password, ipAddress, userAgent string) (*models.LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, models.ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, models.ErrUserInactive
	}

	// Verify password
	if err := s.VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, models.ErrInvalidCredentials
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(user.ID); err != nil {
		// Log error but don't fail login
		fmt.Printf("failed to update last login: %v\n", err)
	}

	// Generate access token
	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// Create session
	session := &models.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    time.Now().Add(s.refreshTokenExpiry),
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	// Get user's trading accounts
	accounts, err := s.tradingAccountRepo.GetByUserID(user.ID)
	if err != nil {
		// Log error but don't fail login
		accounts = []*models.TradingAccount{}
	}

	// Convert to response
	accountResponses := make([]*models.TradingAccountResponse, len(accounts))
	for i, acc := range accounts {
		accountResponses[i] = acc.ToResponse()
	}

	return &models.LoginResponse{
		User:         user.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.tokenExpiry.Seconds()),
		Accounts:     accountResponses,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *Service) RefreshToken(refreshToken string) (*models.RefreshTokenResponse, error) {
	// Get session by refresh token
	session, err := s.sessionRepo.GetByRefreshToken(refreshToken)
	if err != nil {
		return nil, models.ErrSessionNotFound
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		_ = s.sessionRepo.Delete(session.ID)
		return nil, models.ErrSessionExpired
	}

	// Get user
	user, err := s.userRepo.GetByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, models.ErrUserInactive
	}

	// Generate new access token
	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	// Generate new refresh token
	newRefreshToken, err := s.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// Update session with new refresh token
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = time.Now().Add(s.refreshTokenExpiry)
	if err := s.sessionRepo.Create(session); err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	// Delete old session
	_ = s.sessionRepo.Delete(session.ID)

	return &models.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.tokenExpiry.Seconds()),
	}, nil
}

// Logout logs out a user by deleting their sessions
func (s *Service) Logout(userID uuid.UUID) error {
	return s.sessionRepo.DeleteByUserID(userID)
}

// GenerateAccessToken generates a JWT access token
func (s *Service) GenerateAccessToken(user *models.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
		"exp":     now.Add(s.tokenExpiry).Unix(),
		"iat":     now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateAccessToken validates a JWT access token and returns claims
func (s *Service) ValidateAccessToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	if !token.Valid {
		return nil, models.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, models.ErrInvalidToken
	}

	// Extract claims
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, models.ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, models.ErrInvalidToken
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, models.ErrInvalidToken
	}

	roleStr, ok := claims["role"].(string)
	if !ok {
		return nil, models.ErrInvalidToken
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, models.ErrInvalidToken
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, models.ErrInvalidToken
	}

	return &models.JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   models.UserRole(roleStr),
		Exp:    int64(exp),
		Iat:    int64(iat),
	}, nil
}

// GenerateRefreshToken generates a random refresh token
func (s *Service) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// HashPassword hashes a password using bcrypt
func (s *Service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a hash
func (s *Service) VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ChangePassword changes a user's password
func (s *Service) ChangePassword(userID uuid.UUID, currentPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	// Verify current password
	if err := s.VerifyPassword(user.PasswordHash, currentPassword); err != nil {
		return models.ErrPasswordMismatch
	}

	// Hash new password
	newHash, err := s.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Update password
	user.PasswordHash = newHash
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	// Invalidate all sessions to force re-login
	_ = s.sessionRepo.DeleteByUserID(userID)

	return nil
}
