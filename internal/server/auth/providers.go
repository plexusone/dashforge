// Package auth provides authentication for DashForge.
// OAuth provider functionality is provided by SystemForge's identity/oauth package.
package auth

import (
	"context"

	"golang.org/x/oauth2"

	cfoauth "github.com/grokify/systemforge/identity/oauthclient"
)

// OAuthUser represents normalized user info from any OAuth provider.
// This is an alias for SystemForge's oauth.User type.
type OAuthUser = cfoauth.User

// FetchGitHubUser exchanges the authorization code and fetches user info from GitHub.
func FetchGitHubUser(ctx context.Context, config *oauth2.Config, code string) (*OAuthUser, error) {
	return cfoauth.FetchGitHubUser(ctx, config, code)
}

// FetchGoogleUser exchanges the authorization code and fetches user info from Google.
func FetchGoogleUser(ctx context.Context, config *oauth2.Config, code string) (*OAuthUser, error) {
	return cfoauth.FetchGoogleUser(ctx, config, code)
}

// Re-export SystemForge OAuth helpers.
var (
	// GenerateState generates a cryptographically secure random state string.
	GenerateState = cfoauth.GenerateState

	// GoogleConfig creates an OAuth2 config for Google.
	GoogleConfig = cfoauth.GoogleConfig

	// GitHubConfig creates an OAuth2 config for GitHub.
	GitHubConfig = cfoauth.GitHubConfig
)

// Re-export SystemForge OAuth types.
type (
	// ProviderConfig holds OAuth configuration for a provider.
	ProviderConfig = cfoauth.ProviderConfig

	// StateManager handles OAuth state cookie management.
	StateManager = cfoauth.StateManager
)

// NewStateManager creates a state manager with sensible defaults.
var NewStateManager = cfoauth.NewStateManager
