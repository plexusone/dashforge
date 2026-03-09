package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTService handles JWT token generation and validation.
type JWTService struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string
}

// Claims represents JWT claims.
type Claims struct {
	jwt.RegisteredClaims
	UserID         uuid.UUID `json:"uid"`
	Email          string    `json:"email"`
	Role           string    `json:"role,omitempty"`
	OrganizationID uuid.UUID `json:"oid,omitempty"`
}

// TokenPair contains access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"` // seconds until access token expires
	TokenType    string `json:"tokenType"`
}

// JWTConfig configures the JWT service.
type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
}

// NewJWTService creates a new JWT service.
func NewJWTService(cfg JWTConfig) (*JWTService, error) {
	if cfg.Secret == "" {
		return nil, errors.New("JWT secret is required")
	}

	if cfg.AccessTokenTTL == 0 {
		cfg.AccessTokenTTL = 15 * time.Minute
	}
	if cfg.RefreshTokenTTL == 0 {
		cfg.RefreshTokenTTL = 7 * 24 * time.Hour // 7 days
	}
	if cfg.Issuer == "" {
		cfg.Issuer = "dashforge"
	}

	return &JWTService{
		secret:          []byte(cfg.Secret),
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
		issuer:          cfg.Issuer,
	}, nil
}

// GenerateTokenPair generates access and refresh tokens.
func (s *JWTService) GenerateTokenPair(userID uuid.UUID, email, role string) (*TokenPair, error) {
	now := time.Now()

	// Generate access token
	accessClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
		},
		UserID: userID,
		Email:  email,
		Role:   role,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	// Generate refresh token (opaque token)
	refreshToken, err := generateRandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// GenerateTokenPairWithOrganization generates tokens with organization context.
func (s *JWTService) GenerateTokenPairWithOrganization(userID uuid.UUID, email, role string, orgID uuid.UUID) (*TokenPair, error) {
	now := time.Now()

	accessClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
		},
		UserID:         userID,
		Email:          email,
		Role:           role,
		OrganizationID: orgID,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	refreshToken, err := generateRandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// ValidateAccessToken validates and parses an access token.
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshTokens generates new tokens using a refresh token.
// Note: In production, you'd store and validate refresh tokens in a database.
func (s *JWTService) RefreshTokens(refreshToken string) (*TokenPair, error) {
	// In a real implementation, you would:
	// 1. Look up the refresh token in a database
	// 2. Verify it hasn't been revoked
	// 3. Get the associated user info
	// 4. Generate new tokens
	// 5. Optionally rotate the refresh token

	// For now, this is a placeholder that requires the client to re-authenticate
	return nil, errors.New("refresh token validation not implemented - please re-authenticate")
}

// generateRandomToken generates a cryptographically secure random token.
func generateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Middleware for JWT authentication

type jwtContextKey string

const claimsContextKey jwtContextKey = "jwt_claims"

// JWTMiddleware creates middleware that validates JWT tokens.
func (s *JWTService) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := s.ValidateAccessToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalMiddleware validates JWT if present but doesn't require it.
func (s *JWTService) OptionalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := s.ValidateAccessToken(parts[1])
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ClaimsFromContext retrieves JWT claims from the request context.
func ClaimsFromContext(ctx context.Context) *Claims {
	claims, _ := ctx.Value(claimsContextKey).(*Claims)
	return claims
}

// RequireJWTRole middleware checks if the JWT user has the required role.
func RequireJWTRole(roles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]bool)
	for _, r := range roles {
		roleSet[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromContext(r.Context())
			if claims == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !roleSet[claims.Role] {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
