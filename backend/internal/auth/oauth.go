package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

var logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

type GitHubUser struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type OAuthProvider struct {
	config     *oauth2.Config
	provider   string
	getUserURL string
}

var GitHubOAuth = &OAuthProvider{
	config: &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     githuboauth.Endpoint,
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
	},
	provider:   "github",
	getUserURL: "https://api.github.com/user",
}

func (p *OAuthProvider) GenerateAuthURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *OAuthProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

func (p *OAuthProvider) GetUser(ctx context.Context, token *oauth2.Token) (*GitHubUser, error) {
	client := p.config.Client(ctx, token)
	resp, err := client.Get(p.getUserURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from %s: %w", p.provider, err)
	}
	defer resp.Body.Close()

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	if user.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			return nil, fmt.Errorf("failed to get emails from %s: %w", p.provider, err)
		}
		defer emailResp.Body.Close()

		var emails []struct {
			Email   string `json:"email"`
			Primary bool   `json:"primary"`
		}
		if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil {
			return nil, fmt.Errorf("failed to decode emails response: %w", err)
		}
		for _, e := range emails {
			if e.Primary {
				user.Email = e.Email
				break
			}
		}
	}

	return &user, nil
}

func (p *OAuthProvider) GetProvider() string {
	return p.provider
}

func GenerateOAuthState() string {
	return uuid.New().String()
}

func LogOAuthAttempt(provider, username string, success bool, err error) {
	event := logger.Info()
	if !success {
		event = logger.Warn()
	}
	if err != nil {
		event = event.Err(err)
	}
	event.Str("operation", "oauth").
		Str("provider", provider).
		Str("username", username).
		Bool("success", success).
		Time("timestamp", time.Now()).
		Msg("oauth_attempt")
}
