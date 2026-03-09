package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/plexusone/dashforge/ent"
	"github.com/plexusone/dashforge/ent/user"
)

// OAuthConfig holds OAuth provider configurations.
type OAuthConfig struct {
	GitHub      *oauth2.Config
	Google      *oauth2.Config
	CoreControl *CoreControlConfig
}

// CoreControlConfig holds CoreControl (CoreAuth) OAuth configuration.
type CoreControlConfig struct {
	OAuth2   *oauth2.Config
	URL      string   // Base URL of CoreControl server
	Scopes   []string // OAuth scopes to request
}

// OAuthHandler handles OAuth login flows.
type OAuthHandler struct {
	config     OAuthConfig
	jwtService *JWTService
	client     *ent.Client
	logger     *slog.Logger
	baseURL    string
	mux        *http.ServeMux
}

// NewOAuthHandler creates a new OAuth handler.
func NewOAuthHandler(cfg OAuthConfig, jwtSvc *JWTService, client *ent.Client, logger *slog.Logger, baseURL string) *OAuthHandler {
	h := &OAuthHandler{
		config:     cfg,
		jwtService: jwtSvc,
		client:     client,
		logger:     logger,
		baseURL:    baseURL,
		mux:        http.NewServeMux(),
	}
	h.setupRoutes()
	return h
}

func (h *OAuthHandler) setupRoutes() {
	// GitHub OAuth
	h.mux.HandleFunc("GET /api/v1/auth/github", h.handleGitHubLogin)
	h.mux.HandleFunc("GET /api/v1/auth/github/callback", h.handleGitHubCallback)

	// Google OAuth
	h.mux.HandleFunc("GET /api/v1/auth/google", h.handleGoogleLogin)
	h.mux.HandleFunc("GET /api/v1/auth/google/callback", h.handleGoogleCallback)

	// CoreControl OAuth
	h.mux.HandleFunc("GET /api/v1/auth/corecontrol", h.handleCoreControlLogin)
	h.mux.HandleFunc("GET /api/v1/auth/corecontrol/callback", h.handleCoreControlCallback)

	// Token endpoints
	h.mux.HandleFunc("POST /api/v1/auth/refresh", h.handleRefreshToken)
	h.mux.HandleFunc("POST /api/v1/auth/logout", h.handleLogout)

	// User info
	h.mux.HandleFunc("GET /api/v1/auth/me", h.handleMe)
}

// ServeHTTP implements http.Handler.
func (h *OAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// generateState creates a random state string for CSRF protection.
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GitHub OAuth handlers

func (h *OAuthHandler) handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	if h.config.GitHub == nil {
		http.Error(w, "GitHub OAuth not configured", http.StatusNotImplemented)
		return
	}

	state, err := generateState()
	if err != nil {
		h.logger.Error("failed to generate state", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store state in cookie for verification
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})

	url := h.config.GitHub.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	if h.config.GitHub == nil {
		http.Error(w, "GitHub OAuth not configured", http.StatusNotImplemented)
		return
	}

	// Verify state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Check for error from provider
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		h.logger.Warn("OAuth error from GitHub",
			"error", errParam,
			"description", r.URL.Query().Get("error_description"))
		http.Error(w, "Authentication failed: "+errParam, http.StatusUnauthorized)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Fetch user info from GitHub
	oauthUser, err := FetchGitHubUser(r.Context(), h.config.GitHub, code)
	if err != nil {
		h.logger.Error("failed to fetch GitHub user", "error", err)
		http.Error(w, "Failed to authenticate with GitHub", http.StatusInternalServerError)
		return
	}

	h.completeOAuthLogin(w, r, oauthUser, "github")
}

// Google OAuth handlers

func (h *OAuthHandler) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.config.Google == nil {
		http.Error(w, "Google OAuth not configured", http.StatusNotImplemented)
		return
	}

	state, err := generateState()
	if err != nil {
		h.logger.Error("failed to generate state", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store state in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})

	url := h.config.Google.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.config.Google == nil {
		http.Error(w, "Google OAuth not configured", http.StatusNotImplemented)
		return
	}

	// Verify state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Check for error
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		h.logger.Warn("OAuth error from Google",
			"error", errParam,
			"description", r.URL.Query().Get("error_description"))
		http.Error(w, "Authentication failed: "+errParam, http.StatusUnauthorized)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Fetch user info from Google
	oauthUser, err := FetchGoogleUser(r.Context(), h.config.Google, code)
	if err != nil {
		h.logger.Error("failed to fetch Google user", "error", err)
		http.Error(w, "Failed to authenticate with Google", http.StatusInternalServerError)
		return
	}

	h.completeOAuthLogin(w, r, oauthUser, "google")
}

// CoreControl OAuth handlers

func (h *OAuthHandler) handleCoreControlLogin(w http.ResponseWriter, r *http.Request) {
	if h.config.CoreControl == nil {
		http.Error(w, "CoreControl OAuth not configured", http.StatusNotImplemented)
		return
	}

	state, err := generateState()
	if err != nil {
		h.logger.Error("failed to generate state", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store state in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})

	url := h.config.CoreControl.OAuth2.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) handleCoreControlCallback(w http.ResponseWriter, r *http.Request) {
	if h.config.CoreControl == nil {
		http.Error(w, "CoreControl OAuth not configured", http.StatusNotImplemented)
		return
	}

	// Verify state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Check for error
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		h.logger.Warn("OAuth error from CoreControl",
			"error", errParam,
			"description", r.URL.Query().Get("error_description"))
		http.Error(w, "Authentication failed: "+errParam, http.StatusUnauthorized)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := h.config.CoreControl.OAuth2.Exchange(r.Context(), code)
	if err != nil {
		h.logger.Error("failed to exchange code", "error", err)
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	// Fetch user info from CoreControl
	userInfo, err := h.fetchCoreControlUserInfo(r.Context(), token.AccessToken)
	if err != nil {
		h.logger.Error("failed to fetch CoreControl user info", "error", err)
		http.Error(w, "Failed to get user info from CoreControl", http.StatusInternalServerError)
		return
	}

	// Handle CoreControl-specific user creation/linking
	h.completeCoreControlLogin(w, r, userInfo, token.AccessToken)
}

// CoreControlUserInfo represents user info from CoreControl's userinfo endpoint.
type CoreControlUserInfo struct {
	Sub               string `json:"sub"`               // Principal ID (UUID)
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Picture           string `json:"picture"`
}

func (h *OAuthHandler) fetchCoreControlUserInfo(ctx context.Context, accessToken string) (*CoreControlUserInfo, error) {
	userInfoURL := h.config.CoreControl.URL + "/oauth/userinfo"

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status %d", resp.StatusCode)
	}

	var userInfo CoreControlUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (h *OAuthHandler) completeCoreControlLogin(w http.ResponseWriter, r *http.Request, userInfo *CoreControlUserInfo, accessToken string) {
	ctx := r.Context()

	if userInfo.Email == "" {
		http.Error(w, "Email is required for authentication", http.StatusBadRequest)
		return
	}

	// Find or create user via CoreControl principal ID
	u, err := h.findOrCreateUserViaCoreControl(ctx, userInfo)
	if err != nil {
		h.logger.Error("failed to find/create user", "error", err, "email", userInfo.Email)
		http.Error(w, "Failed to create user account", http.StatusInternalServerError)
		return
	}

	// Update last login
	_, err = h.client.User.UpdateOneID(u.ID).
		SetLastLoginAt(time.Now()).
		Save(ctx)
	if err != nil {
		h.logger.Warn("failed to update last login", "error", err)
	}

	// Generate JWT tokens
	tokens, err := h.jwtService.GenerateTokenPair(u.ID, u.Email, "")
	if err != nil {
		h.logger.Error("failed to generate tokens", "error", err)
		http.Error(w, "Failed to generate session", http.StatusInternalServerError)
		return
	}

	// Return tokens (or redirect with token in query for web apps)
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL != "" {
		http.Redirect(w, r, fmt.Sprintf("%s?access_token=%s", redirectURL, tokens.AccessToken), http.StatusTemporaryRedirect)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tokens); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *OAuthHandler) findOrCreateUserViaCoreControl(ctx context.Context, userInfo *CoreControlUserInfo) (*ent.User, error) {
	// Parse the principal ID from sub claim
	principalID, err := parseUUID(userInfo.Sub)
	if err != nil {
		return nil, fmt.Errorf("invalid principal ID: %w", err)
	}

	// Try to find existing user by CoreControl principal ID
	u, err := h.client.User.Query().
		Where(user.CoreControlPrincipalIDEQ(*principalID)).
		Only(ctx)

	if err == nil {
		// Found existing user - update info if needed
		updateQuery := h.client.User.UpdateOneID(u.ID)
		needsUpdate := false

		if userInfo.Name != "" && userInfo.Name != u.Name {
			updateQuery.SetName(userInfo.Name)
			needsUpdate = true
		}
		if userInfo.Picture != "" && userInfo.Picture != u.AvatarURL {
			updateQuery.SetAvatarURL(userInfo.Picture)
			needsUpdate = true
		}

		if needsUpdate {
			u, err = updateQuery.Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("updating user: %w", err)
			}
		}
		return u, nil
	}

	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("querying user by principal ID: %w", err)
	}

	// No user found by principal ID - check if user exists by email (for account linking)
	if userInfo.Email != "" {
		u, err = h.client.User.Query().
			Where(user.EmailEQ(userInfo.Email)).
			Only(ctx)

		if err == nil {
			// User exists with this email - link to CoreControl principal
			u, err = h.client.User.UpdateOneID(u.ID).
				SetCoreControlPrincipalID(*principalID).
				Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("linking user to CoreControl: %w", err)
			}
			h.logger.Info("linked existing user to CoreControl", "user_id", u.ID, "principal_id", principalID)
			return u, nil
		}

		if !ent.IsNotFound(err) {
			return nil, fmt.Errorf("querying user by email: %w", err)
		}
	}

	// Create new user with CoreControl principal ID
	name := userInfo.Name
	if name == "" {
		name = userInfo.Email
	}

	u, err = h.client.User.Create().
		SetEmail(userInfo.Email).
		SetName(name).
		SetCoreControlPrincipalID(*principalID).
		SetAvatarURL(userInfo.Picture).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	h.logger.Info("created new user via CoreControl",
		"user_id", u.ID,
		"email", u.Email,
		"principal_id", principalID)

	return u, nil
}

// parseUUID parses a string into a UUID pointer.
func parseUUID(s string) (*uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// completeOAuthLogin handles the common logic after OAuth authentication.
func (h *OAuthHandler) completeOAuthLogin(w http.ResponseWriter, r *http.Request, oauthUser *OAuthUser, provider string) {
	ctx := r.Context()

	if oauthUser.Email == "" {
		http.Error(w, "Email is required for authentication", http.StatusBadRequest)
		return
	}

	// Find or create user
	u, err := h.findOrCreateUser(ctx, oauthUser, provider)
	if err != nil {
		h.logger.Error("failed to find/create user", "error", err, "email", oauthUser.Email)
		http.Error(w, "Failed to create user account", http.StatusInternalServerError)
		return
	}

	// Update last login
	_, err = h.client.User.UpdateOneID(u.ID).
		SetLastLoginAt(time.Now()).
		Save(ctx)
	if err != nil {
		h.logger.Warn("failed to update last login", "error", err)
	}

	// Generate JWT tokens
	// Role is determined by membership, so we pass empty here
	// The actual role will be checked per-organization when needed
	tokens, err := h.jwtService.GenerateTokenPair(u.ID, u.Email, "")
	if err != nil {
		h.logger.Error("failed to generate tokens", "error", err)
		http.Error(w, "Failed to generate session", http.StatusInternalServerError)
		return
	}

	// Return tokens (or redirect with token in query for web apps)
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL != "" {
		// Redirect back to app with token
		http.Redirect(w, r, fmt.Sprintf("%s?access_token=%s", redirectURL, tokens.AccessToken), http.StatusTemporaryRedirect)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tokens); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// findOrCreateUser finds an existing user by email or creates a new one.
func (h *OAuthHandler) findOrCreateUser(ctx context.Context, oauthUser *OAuthUser, provider string) (*ent.User, error) {
	// Try to find existing user
	u, err := h.client.User.Query().
		Where(user.EmailEQ(oauthUser.Email)).
		Only(ctx)

	if err == nil {
		return u, nil
	}

	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("querying user: %w", err)
	}

	// Create new user
	// Note: In multi-org setup, you'd need to determine the organization here
	// For now, we create without organization (would need to be assigned later via membership)
	name := oauthUser.Name
	if name == "" {
		name = oauthUser.Email
	}

	u, err = h.client.User.Create().
		SetEmail(oauthUser.Email).
		SetName(name).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	h.logger.Info("created new user via OAuth",
		"user_id", u.ID,
		"email", u.Email,
		"provider", provider)

	return u, nil
}

// Token handlers

func (h *OAuthHandler) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	tokens, err := h.jwtService.RefreshTokens(req.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tokens); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *OAuthHandler) handleLogout(w http.ResponseWriter, _ *http.Request) {
	// For stateless JWT, logout is handled client-side by discarding the token
	// Here we could add the token to a blacklist if needed
	w.WriteHeader(http.StatusNoContent)
}

func (h *OAuthHandler) handleMe(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch full user info
	u, err := h.client.User.Get(r.Context(), claims.UserID)
	if err != nil {
		if ent.IsNotFound(err) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"id":              u.ID,
		"email":           u.Email,
		"name":            u.Name,
		"isPlatformAdmin": u.IsPlatformAdmin,
		"active":          u.Active,
		"lastLoginAt":     u.LastLoginAt,
		"createdAt":       u.CreatedAt,
	}); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// OAuthProviderConfig holds all OAuth provider settings.
type OAuthProviderConfig struct {
	GitHubClientID          string
	GitHubClientSecret      string
	GoogleClientID          string
	GoogleClientSecret      string
	CoreControlURL          string
	CoreControlClientID     string
	CoreControlClientSecret string
	CoreControlCallbackURL  string
	CoreControlScopes       []string
	BaseURL                 string
}

// NewOAuthConfig creates OAuth configurations from provider config.
func NewOAuthConfig(pc OAuthProviderConfig) OAuthConfig {
	cfg := OAuthConfig{}

	if pc.GitHubClientID != "" && pc.GitHubClientSecret != "" {
		cfg.GitHub = &oauth2.Config{
			ClientID:     pc.GitHubClientID,
			ClientSecret: pc.GitHubClientSecret,
			RedirectURL:  pc.BaseURL + "/api/v1/auth/github/callback",
			Endpoint:     github.Endpoint,
			Scopes:       []string{"user:email"},
		}
	}

	if pc.GoogleClientID != "" && pc.GoogleClientSecret != "" {
		cfg.Google = &oauth2.Config{
			ClientID:     pc.GoogleClientID,
			ClientSecret: pc.GoogleClientSecret,
			RedirectURL:  pc.BaseURL + "/api/v1/auth/google/callback",
			Endpoint:     google.Endpoint,
			Scopes: []string{
				"openid",
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
		}
	}

	if pc.CoreControlURL != "" && pc.CoreControlClientID != "" && pc.CoreControlClientSecret != "" {
		scopes := pc.CoreControlScopes
		if len(scopes) == 0 {
			scopes = []string{"openid", "profile", "email"}
		}
		callbackURL := pc.CoreControlCallbackURL
		if callbackURL == "" {
			callbackURL = pc.BaseURL + "/api/v1/auth/corecontrol/callback"
		}
		cfg.CoreControl = &CoreControlConfig{
			URL:    pc.CoreControlURL,
			Scopes: scopes,
			OAuth2: &oauth2.Config{
				ClientID:     pc.CoreControlClientID,
				ClientSecret: pc.CoreControlClientSecret,
				RedirectURL:  callbackURL,
				Endpoint: oauth2.Endpoint{
					AuthURL:  pc.CoreControlURL + "/oauth/authorize",
					TokenURL: pc.CoreControlURL + "/oauth/token",
				},
				Scopes: scopes,
			},
		}
	}

	return cfg
}
