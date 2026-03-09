package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

// OAuthUser represents normalized user info from any OAuth provider.
type OAuthUser struct {
	ProviderID   string `json:"providerId"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	AvatarURL    string `json:"avatarUrl"`
	Provider     string `json:"provider"`
	AccessToken  string `json:"-"`
	RefreshToken string `json:"-"`
}

// GitHub OAuth

const (
	gitHubUserURL   = "https://api.github.com/user"
	gitHubEmailsURL = "https://api.github.com/user/emails"
)

type gitHubUserInfo struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type gitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// FetchGitHubUser exchanges the authorization code and fetches user info from GitHub.
func FetchGitHubUser(ctx context.Context, config *oauth2.Config, code string) (*OAuthUser, error) {
	// Exchange code for token
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	client := config.Client(ctx, token)

	// Fetch user info
	resp, err := client.Get(gitHubUserURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch user info: status %d, body: %s", resp.StatusCode, string(body))
	}

	var userInfo gitHubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// GitHub might not return email in the user object if it's private
	email := userInfo.Email
	if email == "" {
		email, err = fetchGitHubPrimaryEmail(client)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GitHub email: %w", err)
		}
	}

	name := userInfo.Name
	if name == "" {
		name = userInfo.Login // Fall back to username
	}

	return &OAuthUser{
		ProviderID:   fmt.Sprintf("%d", userInfo.ID),
		Email:        email,
		Name:         name,
		AvatarURL:    userInfo.AvatarURL,
		Provider:     "github",
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}

// fetchGitHubPrimaryEmail fetches the primary verified email from GitHub.
func fetchGitHubPrimaryEmail(client *http.Client) (string, error) {
	resp, err := client.Get(gitHubEmailsURL)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch emails: status %d, body: %s", resp.StatusCode, string(body))
	}

	var emails []gitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	// Find primary verified email
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	// Fall back to any verified email
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}

	return "", fmt.Errorf("no verified email found")
}

// Google OAuth

const googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// FetchGoogleUser exchanges the authorization code and fetches user info from Google.
func FetchGoogleUser(ctx context.Context, config *oauth2.Config, code string) (*OAuthUser, error) {
	// Exchange code for token
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	client := config.Client(ctx, token)

	// Fetch user info
	resp, err := client.Get(googleUserInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch user info: status %d, body: %s", resp.StatusCode, string(body))
	}

	var userInfo googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	if !userInfo.VerifiedEmail {
		return nil, fmt.Errorf("email not verified")
	}

	return &OAuthUser{
		ProviderID:   userInfo.ID,
		Email:        userInfo.Email,
		Name:         userInfo.Name,
		AvatarURL:    userInfo.Picture,
		Provider:     "google",
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}
