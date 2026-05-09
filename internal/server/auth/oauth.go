package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/plexusone/dashforge/ent"
	"github.com/plexusone/dashforge/ent/human"
	"github.com/plexusone/dashforge/ent/principal"
)

// OAuthConfig holds OAuth provider configurations.
type OAuthConfig struct {
	GitHub      *oauth2.Config
	Google      *oauth2.Config
	CoreControl *CoreControlConfig
}

// CoreControlConfig holds CoreControl (CoreAuth) OAuth configuration.
type CoreControlConfig struct {
	OAuth2 *oauth2.Config
	URL    string   // Base URL of CoreControl server
	Scopes []string // OAuth scopes to request
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

// generateState is provided by SystemForge oauth package via providers.go

// GitHub OAuth handlers

func (h *OAuthHandler) handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	if h.config.GitHub == nil {
		http.Error(w, "GitHub OAuth not configured", http.StatusNotImplemented)
		return
	}

	state, err := GenerateState()
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
	h.handleOAuthCallback(w, r, "GitHub", h.config.GitHub, func(ctx context.Context, cfg *oauth2.Config, code string) (*OAuthUser, error) {
		return FetchGitHubUser(ctx, cfg, code)
	})
}

// Google OAuth handlers

func (h *OAuthHandler) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.config.Google == nil {
		http.Error(w, "Google OAuth not configured", http.StatusNotImplemented)
		return
	}

	state, err := GenerateState()
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
	h.handleOAuthCallback(w, r, "Google", h.config.Google, func(ctx context.Context, cfg *oauth2.Config, code string) (*OAuthUser, error) {
		return FetchGoogleUser(ctx, cfg, code)
	})
}

// oauthUserFetcher fetches user info from an OAuth provider.
type oauthUserFetcher func(ctx context.Context, cfg *oauth2.Config, code string) (*OAuthUser, error)

// handleOAuthCallback is a generic OAuth callback handler.
func (h *OAuthHandler) handleOAuthCallback(w http.ResponseWriter, r *http.Request, providerName string, config *oauth2.Config, fetchUser oauthUserFetcher) {
	if config == nil {
		http.Error(w, providerName+" OAuth not configured", http.StatusNotImplemented)
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
		h.logger.Warn("OAuth error from "+providerName,
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

	// Fetch user info from provider
	oauthUser, err := fetchUser(r.Context(), config, code)
	if err != nil {
		h.logger.Error("failed to fetch "+providerName+" user", "error", err)
		http.Error(w, "Failed to authenticate with "+providerName, http.StatusInternalServerError)
		return
	}

	h.completeOAuthLogin(w, r, oauthUser, strings.ToLower(providerName))
}

// CoreControl OAuth handlers

func (h *OAuthHandler) handleCoreControlLogin(w http.ResponseWriter, r *http.Request) {
	if h.config.CoreControl == nil {
		http.Error(w, "CoreControl OAuth not configured", http.StatusNotImplemented)
		return
	}

	state, err := GenerateState()
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
	Sub               string `json:"sub"` // Principal ID (UUID)
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

func (h *OAuthHandler) completeCoreControlLogin(w http.ResponseWriter, r *http.Request, userInfo *CoreControlUserInfo, _ string) {
	ctx := r.Context()

	if userInfo.Email == "" {
		http.Error(w, "Email is required for authentication", http.StatusBadRequest)
		return
	}

	// Find or create principal via CoreControl principal ID
	p, hum, err := h.findOrCreatePrincipalViaCoreControl(ctx, userInfo)
	if err != nil {
		h.logger.Error("failed to find/create principal", "error", err, "email", userInfo.Email)
		http.Error(w, "Failed to create user account", http.StatusInternalServerError)
		return
	}

	// Update last login on Human
	_, err = h.client.Human.UpdateOneID(hum.ID).
		SetLastLoginAt(time.Now()).
		Save(ctx)
	if err != nil {
		h.logger.Warn("failed to update last login", "error", err)
	}

	// Generate JWT tokens using principal ID
	tokens, err := h.jwtService.GenerateTokenPair(p.ID, hum.Email, p.DisplayName)
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
	if err := json.NewEncoder(w).Encode(tokens); err != nil { //nolint:gosec // G117: OAuth token response per RFC 6749
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *OAuthHandler) findOrCreatePrincipalViaCoreControl(ctx context.Context, userInfo *CoreControlUserInfo) (*ent.Principal, *ent.Human, error) {
	// Parse the CoreControl principal ID from sub claim
	coreControlPrincipalID, err := parseUUID(userInfo.Sub)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid principal ID: %w", err)
	}

	// Try to find existing principal by CoreControl principal ID
	p, err := h.client.Principal.Query().
		Where(principal.CoreControlPrincipalIDEQ(*coreControlPrincipalID)).
		WithHuman().
		Only(ctx)

	if err == nil {
		// Found existing principal - update Human info if needed
		hum := p.Edges.Human
		if hum == nil {
			return nil, nil, fmt.Errorf("principal has no human extension")
		}

		updateQuery := h.client.Human.UpdateOneID(hum.ID)
		needsUpdate := false

		if userInfo.Name != "" && userInfo.Name != hum.Name {
			updateQuery.SetName(userInfo.Name)
			needsUpdate = true
		}
		if userInfo.Picture != "" && (hum.AvatarURL == nil || userInfo.Picture != *hum.AvatarURL) {
			updateQuery.SetAvatarURL(userInfo.Picture)
			needsUpdate = true
		}

		if needsUpdate {
			hum, err = updateQuery.Save(ctx)
			if err != nil {
				return nil, nil, fmt.Errorf("updating human: %w", err)
			}
		}
		return p, hum, nil
	}

	if !ent.IsNotFound(err) {
		return nil, nil, fmt.Errorf("querying principal by CoreControl ID: %w", err)
	}

	// No principal found by CoreControl ID - check if human exists by email (for account linking)
	if userInfo.Email != "" {
		hum, err := h.client.Human.Query().
			Where(human.EmailEQ(userInfo.Email)).
			WithPrincipal().
			Only(ctx)

		if err == nil {
			// Human exists with this email - link principal to CoreControl
			p := hum.Edges.Principal
			p, err = h.client.Principal.UpdateOneID(p.ID).
				SetCoreControlPrincipalID(*coreControlPrincipalID).
				Save(ctx)
			if err != nil {
				return nil, nil, fmt.Errorf("linking principal to CoreControl: %w", err)
			}
			h.logger.Info("linked existing principal to CoreControl", "principal_id", p.ID, "core_control_id", coreControlPrincipalID)
			return p, hum, nil
		}

		if !ent.IsNotFound(err) {
			return nil, nil, fmt.Errorf("querying human by email: %w", err)
		}
	}

	// Create new Principal + Human with CoreControl principal ID
	name := userInfo.Name
	if name == "" {
		name = userInfo.Email
	}

	// Create principal first
	p, err = h.client.Principal.Create().
		SetType(principal.TypeHuman).
		SetIdentifier(userInfo.Email).
		SetDisplayName(name).
		SetCoreControlPrincipalID(*coreControlPrincipalID).
		SetActive(true).
		Save(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("creating principal: %w", err)
	}

	// Create human extension
	hum, err := h.client.Human.Create().
		SetPrincipalID(p.ID).
		SetEmail(userInfo.Email).
		SetName(name).
		SetAvatarURL(userInfo.Picture).
		Save(ctx)
	if err != nil {
		// Rollback principal creation
		_ = h.client.Principal.DeleteOneID(p.ID).Exec(ctx)
		return nil, nil, fmt.Errorf("creating human: %w", err)
	}

	h.logger.Info("created new principal via CoreControl",
		"principal_id", p.ID,
		"email", userInfo.Email,
		"core_control_id", coreControlPrincipalID)

	return p, hum, nil
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

	// Find or create principal
	p, hum, err := h.findOrCreatePrincipal(ctx, oauthUser, provider)
	if err != nil {
		h.logger.Error("failed to find/create principal", "error", err, "email", oauthUser.Email)
		http.Error(w, "Failed to create user account", http.StatusInternalServerError)
		return
	}

	// Update last login on Human
	_, err = h.client.Human.UpdateOneID(hum.ID).
		SetLastLoginAt(time.Now()).
		Save(ctx)
	if err != nil {
		h.logger.Warn("failed to update last login", "error", err)
	}

	// Generate JWT tokens using principal ID
	// Role is determined by membership, so we pass empty here
	// The actual role will be checked per-organization when needed
	tokens, err := h.jwtService.GenerateTokenPair(p.ID, hum.Email, p.DisplayName)
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
	if err := json.NewEncoder(w).Encode(tokens); err != nil { //nolint:gosec // G117: OAuth token response per RFC 6749
		h.logger.Error("failed to encode response", "error", err)
	}
}

// findOrCreatePrincipal finds an existing principal by email or creates a new one.
func (h *OAuthHandler) findOrCreatePrincipal(ctx context.Context, oauthUser *OAuthUser, provider string) (*ent.Principal, *ent.Human, error) {
	// Try to find existing human by email
	hum, err := h.client.Human.Query().
		Where(human.EmailEQ(oauthUser.Email)).
		WithPrincipal().
		Only(ctx)

	if err == nil {
		return hum.Edges.Principal, hum, nil
	}

	if !ent.IsNotFound(err) {
		return nil, nil, fmt.Errorf("querying human: %w", err)
	}

	// Create new Principal + Human
	// Note: In multi-org setup, you'd need to determine the organization here
	// For now, we create without organization (would need to be assigned later via membership)
	name := oauthUser.Name
	if name == "" {
		name = oauthUser.Email
	}

	// Create principal first
	p, err := h.client.Principal.Create().
		SetType(principal.TypeHuman).
		SetIdentifier(oauthUser.Email).
		SetDisplayName(name).
		SetActive(true).
		Save(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("creating principal: %w", err)
	}

	// Create human extension
	hum, err = h.client.Human.Create().
		SetPrincipalID(p.ID).
		SetEmail(oauthUser.Email).
		SetName(name).
		Save(ctx)
	if err != nil {
		// Rollback principal creation
		_ = h.client.Principal.DeleteOneID(p.ID).Exec(ctx)
		return nil, nil, fmt.Errorf("creating human: %w", err)
	}

	h.logger.Info("created new principal via OAuth",
		"principal_id", p.ID,
		"email", oauthUser.Email,
		"provider", provider)

	return p, hum, nil
}

// Token handlers

func (h *OAuthHandler) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement refresh token validation using RefreshToken entity in DB.
	// For now, return error - clients should re-authenticate via OAuth.
	// Implementation would:
	// 1. Look up refresh token in DB
	// 2. Verify it hasn't been revoked and hasn't expired
	// 3. Get the associated principal
	// 4. Generate new token pair
	// 5. Optionally rotate the refresh token
	http.Error(w, "Refresh token validation not implemented - please re-authenticate", http.StatusUnauthorized)
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

	// Fetch principal with human extension
	p, err := h.client.Principal.Query().
		Where(principal.ID(claims.PrincipalID)).
		WithHuman().
		Only(r.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get principal", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	hum := p.Edges.Human
	var isPlatformAdmin bool
	var lastLoginAt *time.Time
	var email, name string

	if hum != nil {
		isPlatformAdmin = hum.IsPlatformAdmin
		lastLoginAt = hum.LastLoginAt
		email = hum.Email
		name = hum.Name
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"id":              p.ID,
		"email":           email,
		"name":            name,
		"displayName":     p.DisplayName,
		"isPlatformAdmin": isPlatformAdmin,
		"active":          p.Active,
		"lastLoginAt":     lastLoginAt,
		"createdAt":       p.CreatedAt,
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
