package service

import (
	"example/sensorHub/oauth"
)

// OAuthServiceAdapter adapts the oauth.OAuthService to the API interface
type OAuthServiceAdapter struct {
	service *oauth.OAuthService
}

func NewOAuthServiceAdapter(service *oauth.OAuthService) *OAuthServiceAdapter {
	return &OAuthServiceAdapter{service: service}
}

func (a *OAuthServiceAdapter) GetStatus() map[string]interface{} {
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

func (a *OAuthServiceAdapter) GetAuthURL(state string) (string, error) {
	if a.service == nil {
		return "", nil
	}
	return a.service.GetAuthURL(state)
}

func (a *OAuthServiceAdapter) ExchangeCode(code string) error {
	if a.service == nil {
		return nil
	}
	return a.service.ExchangeCode(code)
}

func (a *OAuthServiceAdapter) IsReady() bool {
	if a.service == nil {
		return false
	}
	return a.service.IsReady()
}

func (a *OAuthServiceAdapter) Reload() error {
	if a.service == nil {
		return nil
	}
	return a.service.Reload()
}
