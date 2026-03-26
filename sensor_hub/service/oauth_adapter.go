package service

import (
	"context"
	"example/sensorHub/oauth"
)

// OAuthServiceAdapter adapts the oauth.OAuthService to the API interface
type OAuthServiceAdapter struct {
	service *oauth.OAuthService
}

func NewOAuthServiceAdapter(service *oauth.OAuthService) *OAuthServiceAdapter {
	return &OAuthServiceAdapter{service: service}
}

func (a *OAuthServiceAdapter) GetStatus(ctx context.Context) map[string]interface{} {
	if a.service == nil {
		return map[string]interface{}{
			"configured": false,
			"error":      "OAuth service not available",
		}
	}
	status := a.service.GetStatus()
	return map[string]interface{}{
		"configured":       status.Configured,
		"needs_auth":       status.NeedsAuth,
		"token_valid":      status.TokenValid,
		"token_expiry":     status.TokenExpiry,
		"refresher_active": status.RefresherActive,
		"last_refresh_at":  status.LastRefreshAt,
		"last_error":       status.LastError,
	}
}

func (a *OAuthServiceAdapter) GetAuthURL(ctx context.Context, state string) (string, error) {
	if a.service == nil {
		return "", nil
	}
	return a.service.GetAuthURL(state)
}

func (a *OAuthServiceAdapter) ExchangeCode(ctx context.Context, code string) error {
	if a.service == nil {
		return nil
	}
	return a.service.ExchangeCode(code)
}

func (a *OAuthServiceAdapter) IsReady(ctx context.Context) bool {
	if a.service == nil {
		return false
	}
	return a.service.IsReady()
}

func (a *OAuthServiceAdapter) Reload(ctx context.Context) error {
	if a.service == nil {
		return nil
	}
	return a.service.Reload()
}
