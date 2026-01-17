# OAuth Package Refactor with UI Management Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Refactor the OAuth package to be testable, configurable via application properties, and add UI-based OAuth re-authorization flow with RBAC permission control.

**Architecture:**
- Refactor `oauth` package to use dependency injection for file reading and token sources
- Add OAuth-related configuration to `application_properties` package
- Create new `OAuthService` in `service/` with controllable lifecycle (start/stop token refresh)
- Add API endpoints for OAuth flow (initiate consent, handle callback)
- Add database migration for `manage_oauth` permission
- Create admin UI page for OAuth management with Google consent flow integration

**Tech Stack:** Go (Gin framework), OAuth2 (golang.org/x/oauth2), React + MUI, TypeScript

---

## Phase 1: Application Properties for OAuth

### Task 1.1: Add OAuth Configuration Fields to ApplicationConfiguration

**Files:**
- Modify: `sensor_hub/application_properties/application_configuration.go`
- Modify: `sensor_hub/application_properties/application_properties_defaults.go`

**Step 1: Add configuration struct fields**

Add these fields to `ApplicationConfiguration` struct in `application_configuration.go` after the existing fields (around line 39):

```go
// OAuth configuration
OAuthCredentialsFilePath    string
OAuthTokenFilePath          string
OAuthTokenRefreshIntervalMinutes int
```

**Step 2: Add default values**

Add to `ApplicationPropertiesDefaults` in `application_properties_defaults.go`:

```go
// OAuth defaults
"oauth.credentials.file.path":       "configuration/credentials.json",
"oauth.token.file.path":             "configuration/token.json",
"oauth.token.refresh.interval.minutes": "30",
```

**Step 3: Commit**

```bash
git add sensor_hub/application_properties/application_configuration.go sensor_hub/application_properties/application_properties_defaults.go
git commit -m "feat: add OAuth configuration fields to application properties"
```

---

### Task 1.2: Add OAuth Property Parsing and Setters

**Files:**
- Modify: `sensor_hub/application_properties/application_configuration.go`

**Step 1: Add setter functions**

Add after existing setters (around line 130):

```go
func SetOAuthCredentialsFilePath(path string) {
	AppConfig.OAuthCredentialsFilePath = path
}

func SetOAuthTokenFilePath(path string) {
	AppConfig.OAuthTokenFilePath = path
}

func SetOAuthTokenRefreshIntervalMinutes(minutes int) {
	AppConfig.OAuthTokenRefreshIntervalMinutes = minutes
}
```

**Step 2: Add to ConvertConfigurationToMaps**

Add in `ConvertConfigurationToMaps` function (around line 163):

```go
// OAuth
appProps["oauth.credentials.file.path"] = cfg.OAuthCredentialsFilePath
appProps["oauth.token.file.path"] = cfg.OAuthTokenFilePath
appProps["oauth.token.refresh.interval.minutes"] = strconv.Itoa(cfg.OAuthTokenRefreshIntervalMinutes)
```

**Step 3: Add to LoadConfigurationFromMaps**

Add in `LoadConfigurationFromMaps` function (around line 310):

```go
// OAuth
if v, ok := appProps["oauth.credentials.file.path"]; ok {
	cfg.OAuthCredentialsFilePath = v
}
if v, ok := appProps["oauth.token.file.path"]; ok {
	cfg.OAuthTokenFilePath = v
}
if v, ok := appProps["oauth.token.refresh.interval.minutes"]; ok {
	if i, err := strconv.Atoi(v); err == nil {
		cfg.OAuthTokenRefreshIntervalMinutes = i
	} else {
		log.Printf("invalid oauth.token.refresh.interval.minutes '%s': %v", v, err)
		return nil, err
	}
}
```

**Step 4: Add to ReloadConfig logging struct**

Add to the logging struct in `ReloadConfig` (around line 395):

```go
OAuthCredentialsFilePath         string
OAuthTokenFilePath               string
OAuthTokenRefreshIntervalMinutes int
```

And the values:

```go
AppConfig.OAuthCredentialsFilePath,
AppConfig.OAuthTokenFilePath,
AppConfig.OAuthTokenRefreshIntervalMinutes,
```

**Step 5: Run tests to verify no regressions**

Run: `cd sensor_hub && go test ./application_properties/... -v`
Expected: All existing tests pass

**Step 6: Commit**

```bash
git add sensor_hub/application_properties/application_configuration.go
git commit -m "feat: add OAuth property parsing and setters"
```

---

### Task 1.3: Update Application Properties Tests

**Files:**
- Modify: `sensor_hub/application_properties/application_properties_test.go`

**Step 1: Add OAuth properties to validAppPropsMap helper**

Update `validAppPropsMap()` function to include:

```go
"oauth.credentials.file.path":          "configuration/credentials.json",
"oauth.token.file.path":                "configuration/token.json",
"oauth.token.refresh.interval.minutes": "30",
```

**Step 2: Add test for OAuth config loading**

Add test function:

```go
func TestLoadConfigurationFromMaps_OAuthConfig(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["oauth.credentials.file.path"] = "/custom/creds.json"
	appProps["oauth.token.file.path"] = "/custom/token.json"
	appProps["oauth.token.refresh.interval.minutes"] = "45"
	smtpProps := validSmtpPropsMap()
	dbProps := validDbPropsMap()

	cfg, err := LoadConfigurationFromMaps(appProps, smtpProps, dbProps)

	assert.NoError(t, err)
	assert.Equal(t, "/custom/creds.json", cfg.OAuthCredentialsFilePath)
	assert.Equal(t, "/custom/token.json", cfg.OAuthTokenFilePath)
	assert.Equal(t, 45, cfg.OAuthTokenRefreshIntervalMinutes)
}
```

**Step 3: Run tests**

Run: `cd sensor_hub && go test ./application_properties/... -v`
Expected: All tests pass including new OAuth test

**Step 4: Commit**

```bash
git add sensor_hub/application_properties/application_properties_test.go
git commit -m "test: add OAuth configuration property tests"
```

---

## Phase 2: OAuth Package Refactor with TDD

### Task 2.1: Create OAuth Service Interface and Types

**Files:**
- Create: `sensor_hub/oauth/types.go`

**Step 1: Create types file**

```go
package oauth

import (
	"net/smtp"
	"time"

	"golang.org/x/oauth2"
)

// OAuthStatus represents the current state of OAuth authentication
type OAuthStatus struct {
	Configured   bool      `json:"configured"`
	TokenValid   bool      `json:"token_valid"`
	TokenExpiry  time.Time `json:"token_expiry,omitempty"`
	RefresherActive bool   `json:"refresher_active"`
	LastRefreshAt time.Time `json:"last_refresh_at,omitempty"`
	LastError    string    `json:"last_error,omitempty"`
}

// OAuthServiceInterface defines the contract for OAuth operations
type OAuthServiceInterface interface {
	// Initialise loads credentials and token from configured paths
	Initialise() error
	
	// GetStatus returns the current OAuth status
	GetStatus() OAuthStatus
	
	// IsReady returns true if OAuth is configured and token is valid
	IsReady() bool
	
	// GetToken returns the current OAuth token (may be nil)
	GetToken() *oauth2.Token
	
	// GetSMTPAuth returns an smtp.Auth for use with Gmail SMTP
	GetSMTPAuth(username string) smtp.Auth
	
	// StartTokenRefresher starts the background token refresh goroutine
	StartTokenRefresher()
	
	// StopTokenRefresher stops the background token refresh goroutine
	StopTokenRefresher()
	
	// GetAuthURL returns the Google OAuth consent URL for re-authorization
	GetAuthURL(state string) (string, error)
	
	// ExchangeCode exchanges an authorization code for a token and persists it
	ExchangeCode(code string) error
}

// FileReader abstracts file reading for testability
type FileReader interface {
	ReadFile(path string) ([]byte, error)
}

// FileWriter abstracts file writing for testability
type FileWriter interface {
	WriteFile(path string, data []byte, perm uint32) error
}
```

**Step 2: Commit**

```bash
git add sensor_hub/oauth/types.go
git commit -m "feat: add OAuth service interface and types"
```

---

### Task 2.2: Create OAuth XOAuth2Auth Implementation Tests

**Files:**
- Create: `sensor_hub/oauth/oauth_test.go`

**Step 1: Write failing test for XOauth2Auth.Start**

```go
package oauth

import (
	"net/smtp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXOauth2Auth_Start_ReturnsCorrectMechanism(t *testing.T) {
	auth := &XOauth2Auth{
		Username:    "user@example.com",
		AccessToken: "test-access-token",
	}

	mechanism, response, err := auth.Start(&smtp.ServerInfo{})

	assert.NoError(t, err)
	assert.Equal(t, "XOAUTH2", mechanism)
	assert.NotEmpty(t, response)
}

func TestXOauth2Auth_Start_FormatsAuthStringCorrectly(t *testing.T) {
	auth := &XOauth2Auth{
		Username:    "user@example.com",
		AccessToken: "my-access-token",
	}

	_, response, err := auth.Start(&smtp.ServerInfo{})

	assert.NoError(t, err)
	// XOAUTH2 format: "user=<user>\x01auth=Bearer <token>\x01\x01"
	expected := "user=user@example.com\x01auth=Bearer my-access-token\x01\x01"
	assert.Equal(t, expected, string(response))
}

func TestXOauth2Auth_Next_ReturnsNil(t *testing.T) {
	auth := &XOauth2Auth{
		Username:    "user@example.com",
		AccessToken: "test-token",
	}

	response, err := auth.Next([]byte("challenge"), true)

	assert.NoError(t, err)
	assert.Nil(t, response)
}
```

**Step 2: Run test to verify it fails**

Run: `cd sensor_hub && go test ./oauth/... -v`
Expected: FAIL - XOauth2Auth not properly implemented yet

**Step 3: Commit failing test**

```bash
git add sensor_hub/oauth/oauth_test.go
git commit -m "test: add XOauth2Auth tests (failing)"
```

---

### Task 2.3: Implement XOauth2Auth

**Files:**
- Modify: `sensor_hub/oauth/oauth.go`

**Step 1: Update XOauth2Auth implementation**

Replace the existing oauth.go with a cleaned up version:

```go
package oauth

import (
	"fmt"
	"net/smtp"
)

// XOauth2Auth implements smtp.Auth for XOAUTH2 authentication
type XOauth2Auth struct {
	Username    string
	AccessToken string
}

// Start initiates the XOAUTH2 authentication
func (a *XOauth2Auth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
	authString := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.Username, a.AccessToken)
	return "XOAUTH2", []byte(authString), nil
}

// Next handles additional authentication challenges (not used in XOAUTH2)
func (a *XOauth2Auth) Next(_ []byte, _ bool) ([]byte, error) {
	return nil, nil
}
```

**Step 2: Run tests to verify they pass**

Run: `cd sensor_hub && go test ./oauth/... -v`
Expected: PASS

**Step 3: Commit**

```bash
git add sensor_hub/oauth/oauth.go
git commit -m "feat: implement XOauth2Auth SMTP authentication"
```

---

### Task 2.4: Create Mock File System for Testing

**Files:**
- Create: `sensor_hub/oauth/test_helpers_test.go`

**Step 1: Create test helpers**

```go
package oauth

import (
	"encoding/json"
	"errors"

	"golang.org/x/oauth2"
)

// MockFileReader implements FileReader for testing
type MockFileReader struct {
	Files map[string][]byte
	Err   error
}

func (m *MockFileReader) ReadFile(path string) ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	data, ok := m.Files[path]
	if !ok {
		return nil, errors.New("file not found: " + path)
	}
	return data, nil
}

// MockFileWriter implements FileWriter for testing
type MockFileWriter struct {
	Written map[string][]byte
	Err     error
}

func NewMockFileWriter() *MockFileWriter {
	return &MockFileWriter{Written: make(map[string][]byte)}
}

func (m *MockFileWriter) WriteFile(path string, data []byte, perm uint32) error {
	if m.Err != nil {
		return m.Err
	}
	m.Written[path] = data
	return nil
}

// Test data factories

func testCredentialsJSON() []byte {
	// Minimal valid Google OAuth credentials structure
	creds := map[string]interface{}{
		"installed": map[string]interface{}{
			"client_id":     "test-client-id.apps.googleusercontent.com",
			"client_secret": "test-client-secret",
			"auth_uri":      "https://accounts.google.com/o/oauth2/auth",
			"token_uri":     "https://oauth2.googleapis.com/token",
			"redirect_uris": []string{"urn:ietf:wg:oauth:2.0:oob", "http://localhost"},
		},
	}
	data, _ := json.Marshal(creds)
	return data
}

func testTokenJSON() []byte {
	token := oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
	}
	data, _ := json.Marshal(token)
	return data
}

func testExpiredTokenJSON() []byte {
	token := map[string]interface{}{
		"access_token":  "expired-access-token",
		"token_type":    "Bearer",
		"refresh_token": "test-refresh-token",
		"expiry":        "2020-01-01T00:00:00Z",
	}
	data, _ := json.Marshal(token)
	return data
}
```

**Step 2: Commit**

```bash
git add sensor_hub/oauth/test_helpers_test.go
git commit -m "test: add OAuth mock file system helpers"
```

---

### Task 2.5: Create OAuthService Tests

**Files:**
- Modify: `sensor_hub/oauth/oauth_test.go`

**Step 1: Add service tests**

Append to `oauth_test.go`:

```go
func TestNewOAuthService_CreatesService(t *testing.T) {
	reader := &MockFileReader{Files: make(map[string][]byte)}
	writer := NewMockFileWriter()

	service := NewOAuthService(reader, writer, "/creds.json", "/token.json", 30)

	assert.NotNil(t, service)
}

func TestOAuthService_GetStatus_NotConfigured(t *testing.T) {
	reader := &MockFileReader{Files: make(map[string][]byte)}
	writer := NewMockFileWriter()

	service := NewOAuthService(reader, writer, "/creds.json", "/token.json", 30)
	status := service.GetStatus()

	assert.False(t, status.Configured)
	assert.False(t, status.TokenValid)
	assert.False(t, status.RefresherActive)
}

func TestOAuthService_Initialise_MissingCredentials(t *testing.T) {
	reader := &MockFileReader{Files: make(map[string][]byte)}
	writer := NewMockFileWriter()

	service := NewOAuthService(reader, writer, "/creds.json", "/token.json", 30)
	err := service.Initialise()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "credentials")
}

func TestOAuthService_Initialise_MissingToken(t *testing.T) {
	reader := &MockFileReader{
		Files: map[string][]byte{
			"/creds.json": testCredentialsJSON(),
		},
	}
	writer := NewMockFileWriter()

	service := NewOAuthService(reader, writer, "/creds.json", "/token.json", 30)
	err := service.Initialise()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token")
}

func TestOAuthService_Initialise_Success(t *testing.T) {
	reader := &MockFileReader{
		Files: map[string][]byte{
			"/creds.json": testCredentialsJSON(),
			"/token.json": testTokenJSON(),
		},
	}
	writer := NewMockFileWriter()

	service := NewOAuthService(reader, writer, "/creds.json", "/token.json", 30)
	err := service.Initialise()

	assert.NoError(t, err)
	status := service.GetStatus()
	assert.True(t, status.Configured)
}

func TestOAuthService_IsReady_AfterSuccessfulInit(t *testing.T) {
	reader := &MockFileReader{
		Files: map[string][]byte{
			"/creds.json": testCredentialsJSON(),
			"/token.json": testTokenJSON(),
		},
	}
	writer := NewMockFileWriter()

	service := NewOAuthService(reader, writer, "/creds.json", "/token.json", 30)
	_ = service.Initialise()

	assert.True(t, service.IsReady())
}

func TestOAuthService_GetToken_ReturnsToken(t *testing.T) {
	reader := &MockFileReader{
		Files: map[string][]byte{
			"/creds.json": testCredentialsJSON(),
			"/token.json": testTokenJSON(),
		},
	}
	writer := NewMockFileWriter()

	service := NewOAuthService(reader, writer, "/creds.json", "/token.json", 30)
	_ = service.Initialise()
	token := service.GetToken()

	assert.NotNil(t, token)
	assert.Equal(t, "test-access-token", token.AccessToken)
}

func TestOAuthService_GetSMTPAuth_ReturnsAuth(t *testing.T) {
	reader := &MockFileReader{
		Files: map[string][]byte{
			"/creds.json": testCredentialsJSON(),
			"/token.json": testTokenJSON(),
		},
	}
	writer := NewMockFileWriter()

	service := NewOAuthService(reader, writer, "/creds.json", "/token.json", 30)
	_ = service.Initialise()
	auth := service.GetSMTPAuth("user@example.com")

	assert.NotNil(t, auth)
}

func TestOAuthService_GetAuthURL_ReturnsURL(t *testing.T) {
	reader := &MockFileReader{
		Files: map[string][]byte{
			"/creds.json": testCredentialsJSON(),
			"/token.json": testTokenJSON(),
		},
	}
	writer := NewMockFileWriter()

	service := NewOAuthService(reader, writer, "/creds.json", "/token.json", 30)
	_ = service.Initialise()
	url, err := service.GetAuthURL("test-state")

	assert.NoError(t, err)
	assert.Contains(t, url, "accounts.google.com")
	assert.Contains(t, url, "test-state")
}

func TestOAuthService_StartStopTokenRefresher(t *testing.T) {
	reader := &MockFileReader{
		Files: map[string][]byte{
			"/creds.json": testCredentialsJSON(),
			"/token.json": testTokenJSON(),
		},
	}
	writer := NewMockFileWriter()

	service := NewOAuthService(reader, writer, "/creds.json", "/token.json", 30)
	_ = service.Initialise()
	
	service.StartTokenRefresher()
	assert.True(t, service.GetStatus().RefresherActive)
	
	service.StopTokenRefresher()
	assert.False(t, service.GetStatus().RefresherActive)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd sensor_hub && go test ./oauth/... -v`
Expected: FAIL - OAuthService not implemented

**Step 3: Commit failing tests**

```bash
git add sensor_hub/oauth/oauth_test.go
git commit -m "test: add OAuthService tests (failing)"
```

---

### Task 2.6: Implement OAuthService

**Files:**
- Create: `sensor_hub/oauth/oauth_service.go`

**Step 1: Implement the service**

```go
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/smtp"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthService manages OAuth tokens and provides SMTP authentication
type OAuthService struct {
	reader              FileReader
	writer              FileWriter
	credentialsPath     string
	tokenPath           string
	refreshIntervalMins int

	mu              sync.RWMutex
	config          *oauth2.Config
	token           *oauth2.Token
	configured      bool
	lastRefreshAt   time.Time
	lastError       string
	refresherActive bool
	stopChan        chan struct{}
}

// NewOAuthService creates a new OAuthService
func NewOAuthService(reader FileReader, writer FileWriter, credentialsPath, tokenPath string, refreshIntervalMins int) *OAuthService {
	return &OAuthService{
		reader:              reader,
		writer:              writer,
		credentialsPath:     credentialsPath,
		tokenPath:           tokenPath,
		refreshIntervalMins: refreshIntervalMins,
	}
}

// Initialise loads credentials and token from configured paths
func (s *OAuthService) Initialise() error {
	credBytes, err := s.reader.ReadFile(s.credentialsPath)
	if err != nil {
		s.mu.Lock()
		s.lastError = fmt.Sprintf("unable to read credentials: %v", err)
		s.mu.Unlock()
		return fmt.Errorf("unable to read credentials.json: %w", err)
	}

	config, err := google.ConfigFromJSON(credBytes, "https://mail.google.com/")
	if err != nil {
		s.mu.Lock()
		s.lastError = fmt.Sprintf("unable to parse credentials: %v", err)
		s.mu.Unlock()
		return fmt.Errorf("unable to parse credentials.json: %w", err)
	}

	tokenBytes, err := s.reader.ReadFile(s.tokenPath)
	if err != nil {
		s.mu.Lock()
		s.lastError = fmt.Sprintf("unable to read token: %v", err)
		s.mu.Unlock()
		return fmt.Errorf("unable to read token.json: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		s.mu.Lock()
		s.lastError = fmt.Sprintf("unable to parse token: %v", err)
		s.mu.Unlock()
		return fmt.Errorf("unable to unmarshal token.json: %w", err)
	}

	s.mu.Lock()
	s.config = config
	s.token = &token
	s.configured = true
	s.lastError = ""
	s.mu.Unlock()

	return nil
}

// GetStatus returns the current OAuth status
func (s *OAuthService) GetStatus() OAuthStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := OAuthStatus{
		Configured:      s.configured,
		RefresherActive: s.refresherActive,
		LastRefreshAt:   s.lastRefreshAt,
		LastError:       s.lastError,
	}

	if s.token != nil {
		status.TokenValid = s.token.Valid()
		status.TokenExpiry = s.token.Expiry
	}

	return status
}

// IsReady returns true if OAuth is configured and token is valid
func (s *OAuthService) IsReady() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.configured && s.token != nil
}

// GetToken returns the current OAuth token
func (s *OAuthService) GetToken() *oauth2.Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.token
}

// GetSMTPAuth returns an smtp.Auth for use with Gmail SMTP
func (s *OAuthService) GetSMTPAuth(username string) smtp.Auth {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.token == nil {
		return nil
	}

	return &XOauth2Auth{
		Username:    username,
		AccessToken: s.token.AccessToken,
	}
}

// StartTokenRefresher starts the background token refresh goroutine
func (s *OAuthService) StartTokenRefresher() {
	s.mu.Lock()
	if s.refresherActive {
		s.mu.Unlock()
		return
	}
	s.stopChan = make(chan struct{})
	s.refresherActive = true
	s.mu.Unlock()

	go s.refreshLoop()
}

// StopTokenRefresher stops the background token refresh goroutine
func (s *OAuthService) StopTokenRefresher() {
	s.mu.Lock()
	if !s.refresherActive {
		s.mu.Unlock()
		return
	}
	close(s.stopChan)
	s.refresherActive = false
	s.mu.Unlock()
}

func (s *OAuthService) refreshLoop() {
	ticker := time.NewTicker(time.Duration(s.refreshIntervalMins) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.refreshToken()
		}
	}
}

func (s *OAuthService) refreshToken() {
	s.mu.RLock()
	config := s.config
	token := s.token
	s.mu.RUnlock()

	if config == nil || token == nil {
		return
	}

	tokenSource := config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		s.mu.Lock()
		s.lastError = fmt.Sprintf("token refresh failed: %v", err)
		s.mu.Unlock()
		return
	}

	s.mu.Lock()
	s.token = newToken
	s.lastRefreshAt = time.Now()
	s.lastError = ""
	s.mu.Unlock()

	// Persist token
	tokenBytes, err := json.Marshal(newToken)
	if err != nil {
		s.mu.Lock()
		s.lastError = fmt.Sprintf("unable to marshal token: %v", err)
		s.mu.Unlock()
		return
	}

	if err := s.writer.WriteFile(s.tokenPath, tokenBytes, 0600); err != nil {
		s.mu.Lock()
		s.lastError = fmt.Sprintf("unable to write token: %v", err)
		s.mu.Unlock()
	}
}

// GetAuthURL returns the Google OAuth consent URL for re-authorization
func (s *OAuthService) GetAuthURL(state string) (string, error) {
	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	if config == nil {
		return "", fmt.Errorf("OAuth not configured")
	}

	return config.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

// ExchangeCode exchanges an authorization code for a token and persists it
func (s *OAuthService) ExchangeCode(code string) error {
	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	if config == nil {
		return fmt.Errorf("OAuth not configured")
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		s.mu.Lock()
		s.lastError = fmt.Sprintf("code exchange failed: %v", err)
		s.mu.Unlock()
		return fmt.Errorf("unable to exchange code: %w", err)
	}

	s.mu.Lock()
	s.token = token
	s.lastError = ""
	s.mu.Unlock()

	// Persist token
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("unable to marshal token: %w", err)
	}

	if err := s.writer.WriteFile(s.tokenPath, tokenBytes, 0600); err != nil {
		return fmt.Errorf("unable to write token: %w", err)
	}

	return nil
}
```

**Step 2: Run tests to verify they pass**

Run: `cd sensor_hub && go test ./oauth/... -v`
Expected: PASS

**Step 3: Commit**

```bash
git add sensor_hub/oauth/oauth_service.go
git commit -m "feat: implement OAuthService with testable dependencies"
```

---

### Task 2.7: Create Real File System Adapter

**Files:**
- Create: `sensor_hub/oauth/file_adapter.go`

**Step 1: Create file adapter**

```go
package oauth

import (
	"os"
)

// OSFileReader implements FileReader using the OS filesystem
type OSFileReader struct{}

func (r *OSFileReader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// OSFileWriter implements FileWriter using the OS filesystem
type OSFileWriter struct{}

func (w *OSFileWriter) WriteFile(path string, data []byte, perm uint32) error {
	return os.WriteFile(path, data, os.FileMode(perm))
}
```

**Step 2: Commit**

```bash
git add sensor_hub/oauth/file_adapter.go
git commit -m "feat: add OS file adapter for OAuth service"
```

---

### Task 2.8: Create Global OAuth Service Instance and Legacy Compatibility

**Files:**
- Modify: `sensor_hub/oauth/oauth.go`

**Step 1: Update oauth.go with global instance and backward compatibility**

Replace oauth.go entirely:

```go
package oauth

import (
	"fmt"
	"net/smtp"

	appProps "example/sensorHub/application_properties"

	"golang.org/x/oauth2"
)

// Global OAuth service instance
var oauthService *OAuthService

// Legacy compatibility - these are kept for backward compatibility with smtp package
var OauthToken *oauth2.Token
var OauthSet = false

// XOauth2Auth implements smtp.Auth for XOAUTH2 authentication
type XOauth2Auth struct {
	Username    string
	AccessToken string
}

// Start initiates the XOAUTH2 authentication
func (a *XOauth2Auth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
	authString := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.Username, a.AccessToken)
	return "XOAUTH2", []byte(authString), nil
}

// Next handles additional authentication challenges (not used in XOAUTH2)
func (a *XOauth2Auth) Next(_ []byte, _ bool) ([]byte, error) {
	return nil, nil
}

// GetService returns the global OAuth service instance
func GetService() *OAuthService {
	return oauthService
}

// InitialiseOauth initializes the global OAuth service using application config
func InitialiseOauth() error {
	cfg := appProps.AppConfig
	if cfg == nil {
		return fmt.Errorf("application config not initialized")
	}

	credPath := cfg.OAuthCredentialsFilePath
	if credPath == "" {
		credPath = "configuration/credentials.json"
	}
	tokenPath := cfg.OAuthTokenFilePath
	if tokenPath == "" {
		tokenPath = "configuration/token.json"
	}
	refreshInterval := cfg.OAuthTokenRefreshIntervalMinutes
	if refreshInterval <= 0 {
		refreshInterval = 30
	}

	oauthService = NewOAuthService(
		&OSFileReader{},
		&OSFileWriter{},
		credPath,
		tokenPath,
		refreshInterval,
	)

	if err := oauthService.Initialise(); err != nil {
		return err
	}

	// Update legacy globals for backward compatibility
	OauthToken = oauthService.GetToken()
	OauthSet = oauthService.IsReady()

	// Start background refresher
	oauthService.StartTokenRefresher()

	return nil
}

// StopOauth stops the token refresher
func StopOauth() {
	if oauthService != nil {
		oauthService.StopTokenRefresher()
	}
}
```

**Step 2: Run tests**

Run: `cd sensor_hub && go test ./oauth/... -v`
Expected: PASS

**Step 3: Commit**

```bash
git add sensor_hub/oauth/oauth.go
git commit -m "feat: add global OAuth service with legacy compatibility"
```

---

## Phase 3: Database Migration for manage_oauth Permission

### Task 3.1: Create Database Migration

**Files:**
- Create: `sensor_hub/db/changesets/V16__add_oauth_permission.sql`

**Step 1: Create migration**

```sql
-- V16: Add manage_oauth permission for OAuth configuration management

INSERT IGNORE INTO permissions (name, description) VALUES ('manage_oauth', 'Manage OAuth credentials and re-authorize');

-- Grant manage_oauth permission to admin role
INSERT IGNORE INTO role_permissions (role_id, permission_id)
  SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin' AND p.name = 'manage_oauth';
```

**Step 2: Commit**

```bash
git add sensor_hub/db/changesets/V16__add_oauth_permission.sql
git commit -m "feat: add manage_oauth permission migration"
```

---

## Phase 4: OAuth API Endpoints

### Task 4.1: Create OAuth API Tests

**Files:**
- Create: `sensor_hub/api/oauth_api_test.go`

**Step 1: Write failing tests**

```go
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"example/sensorHub/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOAuthService struct {
	mock.Mock
}

func (m *MockOAuthService) GetStatus() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockOAuthService) GetAuthURL(state string) (string, error) {
	args := m.Called(state)
	return args.String(0), args.Error(1)
}

func (m *MockOAuthService) ExchangeCode(code string) error {
	args := m.Called(code)
	return args.Error(0)
}

func (m *MockOAuthService) IsReady() bool {
	args := m.Called()
	return args.Bool(0)
}

func setupOAuthRouter() (*gin.Engine, *MockOAuthService) {
	mockService := new(MockOAuthService)
	InitOAuthAPI(mockService)
	router := gin.New()
	return router, mockService
}

func TestGetOAuthStatus_Success(t *testing.T) {
	router, mockService := setupOAuthRouter()
	router.GET("/oauth/status", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "admin", Roles: []string{"admin"}})
		oauthStatusHandler(c)
	})

	mockService.On("GetStatus").Return(map[string]interface{}{
		"configured":   true,
		"token_valid":  true,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/oauth/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "configured")
}

func TestGetOAuthAuthURL_Success(t *testing.T) {
	router, mockService := setupOAuthRouter()
	router.GET("/oauth/authorize", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "admin", Roles: []string{"admin"}})
		oauthAuthorizeHandler(c)
	})

	mockService.On("GetAuthURL", mock.Anything).Return("https://accounts.google.com/oauth?...", nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/oauth/authorize", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "auth_url")
}

func TestOAuthCallback_Success(t *testing.T) {
	router, mockService := setupOAuthRouter()
	router.GET("/oauth/callback", oauthCallbackHandler)

	mockService.On("ExchangeCode", "test-code").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/oauth/callback?code=test-code&state=test-state", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd sensor_hub && go test ./api/... -v -run OAuth`
Expected: FAIL

**Step 3: Commit failing tests**

```bash
git add sensor_hub/api/oauth_api_test.go
git commit -m "test: add OAuth API tests (failing)"
```

---

### Task 4.2: Implement OAuth API Handlers

**Files:**
- Create: `sensor_hub/api/oauth_api.go`

**Step 1: Create API handlers**

```go
package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// OAuthAPIServiceInterface defines what the API needs from OAuth service
type OAuthAPIServiceInterface interface {
	GetStatus() map[string]interface{}
	GetAuthURL(state string) (string, error)
	ExchangeCode(code string) error
	IsReady() bool
}

var oauthAPIService OAuthAPIServiceInterface

// pendingStates stores CSRF states for OAuth flow
var pendingStates = struct {
	sync.RWMutex
	states map[string]bool
}{states: make(map[string]bool)}

func InitOAuthAPI(s OAuthAPIServiceInterface) {
	oauthAPIService = s
}

func oauthStatusHandler(ctx *gin.Context) {
	if oauthAPIService == nil {
		ctx.IndentedJSON(http.StatusServiceUnavailable, gin.H{"message": "OAuth not configured"})
		return
	}
	status := oauthAPIService.GetStatus()
	ctx.IndentedJSON(http.StatusOK, status)
}

func oauthAuthorizeHandler(ctx *gin.Context) {
	if oauthAPIService == nil {
		ctx.IndentedJSON(http.StatusServiceUnavailable, gin.H{"message": "OAuth not configured"})
		return
	}

	// Generate CSRF state
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to generate state"})
		return
	}
	state := hex.EncodeToString(stateBytes)

	// Store state for validation
	pendingStates.Lock()
	pendingStates.states[state] = true
	pendingStates.Unlock()

	authURL, err := oauthAPIService.GetAuthURL(state)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to get auth URL", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"auth_url": authURL, "state": state})
}

func oauthCallbackHandler(ctx *gin.Context) {
	if oauthAPIService == nil {
		ctx.IndentedJSON(http.StatusServiceUnavailable, gin.H{"message": "OAuth not configured"})
		return
	}

	code := ctx.Query("code")
	state := ctx.Query("state")

	if code == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "missing authorization code"})
		return
	}

	// Validate state if provided (callback might be from external redirect)
	if state != "" {
		pendingStates.Lock()
		valid := pendingStates.states[state]
		delete(pendingStates.states, state)
		pendingStates.Unlock()

		if !valid {
			ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid state"})
			return
		}
	}

	if err := oauthAPIService.ExchangeCode(code); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to exchange code", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "OAuth authorization successful"})
}
```

**Step 2: Run tests**

Run: `cd sensor_hub && go test ./api/... -v -run OAuth`
Expected: PASS

**Step 3: Commit**

```bash
git add sensor_hub/api/oauth_api.go
git commit -m "feat: implement OAuth API handlers"
```

---

### Task 4.3: Create OAuth Routes

**Files:**
- Create: `sensor_hub/api/oauth_routes.go`

**Step 1: Create routes file**

```go
package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterOAuthRoutes(router *gin.Engine) {
	oauthGroup := router.Group("/oauth")
	{
		// Status endpoint requires manage_oauth permission
		oauthGroup.GET("/status", middleware.AuthRequired(), middleware.RequirePermission("manage_oauth"), oauthStatusHandler)
		
		// Authorize endpoint requires manage_oauth permission
		oauthGroup.GET("/authorize", middleware.AuthRequired(), middleware.RequirePermission("manage_oauth"), oauthAuthorizeHandler)
		
		// Callback is public (Google redirects here) but validates state
		oauthGroup.GET("/callback", oauthCallbackHandler)
	}
}
```

**Step 2: Commit**

```bash
git add sensor_hub/api/oauth_routes.go
git commit -m "feat: add OAuth routes with permission protection"
```

---

### Task 4.4: Create OAuth Service Adapter for API

**Files:**
- Create: `sensor_hub/service/oauth_adapter.go`

**Step 1: Create adapter**

```go
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
```

**Step 2: Commit**

```bash
git add sensor_hub/service/oauth_adapter.go
git commit -m "feat: add OAuth service adapter for API"
```

---

### Task 4.5: Register OAuth Routes in main.go

**Files:**
- Modify: `sensor_hub/main.go`
- Modify: `sensor_hub/api/api.go`

**Step 1: Update api.go to register OAuth routes**

Add after line 47 (after RegisterAlertRoutes):

```go
RegisterOAuthRoutes(router)
```

**Step 2: Update main.go to initialize OAuth API**

Add after line 63 (after api.InitAlertAPI):

```go
// Initialize OAuth API (may be nil if OAuth not configured)
oauthAdapter := service.NewOAuthServiceAdapter(oauth.GetService())
api.InitOAuthAPI(oauthAdapter)
```

Also update the oauth initialization section (around line 96) to be:

```go
err = oauth.InitialiseOauth()
if err != nil {
	log.Printf("Failed to initialise OAuth: %v", err)
}
// Re-initialize OAuth API adapter after OAuth service is ready
oauthAdapter = service.NewOAuthServiceAdapter(oauth.GetService())
api.InitOAuthAPI(oauthAdapter)
```

**Step 3: Run the application to verify it compiles**

Run: `cd sensor_hub && go build ./...`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add sensor_hub/api/api.go sensor_hub/main.go
git commit -m "feat: wire up OAuth API in main application"
```

---

## Phase 5: UI Implementation

### Task 5.1: Create OAuth API Client

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/api/OAuth.ts`

**Step 1: Create OAuth API client**

```typescript
import { get, post } from './Client';

export interface OAuthStatus {
  configured: boolean;
  token_valid: boolean;
  token_expiry?: string;
  refresher_active: boolean;
  last_refresh_at?: string;
  last_error?: string;
}

export interface OAuthAuthorizeResponse {
  auth_url: string;
  state: string;
}

export const getOAuthStatus = async (): Promise<OAuthStatus> => {
  return get<OAuthStatus>('/oauth/status');
};

export const getOAuthAuthorizeURL = async (): Promise<OAuthAuthorizeResponse> => {
  return get<OAuthAuthorizeResponse>('/oauth/authorize');
};

export const exchangeOAuthCode = async (code: string, state: string): Promise<void> => {
  return get<void>(`/oauth/callback?code=${encodeURIComponent(code)}&state=${encodeURIComponent(state)}`);
};
```

**Step 2: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/api/OAuth.ts
git commit -m "feat: add OAuth API client"
```

---

### Task 5.2: Create OAuth Management Page

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/pages/admin/OAuthPage.tsx`

**Step 1: Create OAuth management page**

```tsx
import { useEffect, useState, useCallback } from 'react';
import {
  Button,
  Box,
  Grid,
  Typography,
  Alert,
  CircularProgress,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from '@mui/material';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import RefreshIcon from '@mui/icons-material/Refresh';
import { getOAuthStatus, getOAuthAuthorizeURL, type OAuthStatus } from '../../api/OAuth';
import PageContainer from '../../tools/PageContainer';
import LayoutCard from '../../tools/LayoutCard';
import { useAuth } from '../../providers/AuthContext';
import { hasPerm } from '../../tools/Utils';

export default function OAuthPage() {
  const [status, setStatus] = useState<OAuthStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [authorizing, setAuthorizing] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const { user } = useAuth();

  const loadStatus = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const s = await getOAuthStatus();
      setStatus(s);
    } catch (err: unknown) {
      const e = err as { message?: string };
      setError(e.message || 'Failed to load OAuth status');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadStatus();
  }, [loadStatus]);

  // Check for OAuth callback parameters in URL
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const code = params.get('code');
    const state = params.get('state');
    if (code && state) {
      // Clear URL parameters
      window.history.replaceState({}, '', window.location.pathname);
      // Reload status after callback
      loadStatus();
    }
  }, [loadStatus]);

  const handleAuthorize = async () => {
    setConfirmOpen(false);
    try {
      setAuthorizing(true);
      const { auth_url } = await getOAuthAuthorizeURL();
      // Open Google OAuth in new window/tab
      window.open(auth_url, '_blank', 'width=600,height=700');
    } catch (err: unknown) {
      const e = err as { message?: string };
      setError(e.message || 'Failed to start authorization');
    } finally {
      setAuthorizing(false);
    }
  };

  if (user === undefined) {
    return (
      <PageContainer titleText="OAuth Management">
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      </PageContainer>
    );
  }

  const canManage = hasPerm(user, 'manage_oauth');

  return (
    <PageContainer titleText="OAuth Management">
      <Box sx={{ flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: '100%', width: '100%' }}>
          <Box sx={{ width: '100%' }}>
            <LayoutCard variant="secondary" changes={{ alignItems: 'stretch', height: '100%', width: '100%' }}>
              <Box display="flex" alignItems="center" justifyContent="space-between" gap={2} mb={2}>
                <Typography variant="h4">OAuth Configuration</Typography>
                <Button
                  variant="outlined"
                  startIcon={<RefreshIcon />}
                  onClick={loadStatus}
                  disabled={loading}
                >
                  Refresh
                </Button>
              </Box>

              {error && (
                <Alert severity="error" sx={{ mb: 2 }}>
                  {error}
                </Alert>
              )}

              {loading && !status ? (
                <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
                  <CircularProgress />
                </Box>
              ) : status ? (
                <Box>
                  <Typography variant="h6" gutterBottom>
                    Status
                  </Typography>

                  <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2, mb: 3 }}>
                    <Chip
                      icon={status.configured ? <CheckCircleIcon /> : <ErrorIcon />}
                      label={status.configured ? 'Configured' : 'Not Configured'}
                      color={status.configured ? 'success' : 'error'}
                      variant="outlined"
                    />
                    <Chip
                      icon={status.token_valid ? <CheckCircleIcon /> : <ErrorIcon />}
                      label={status.token_valid ? 'Token Valid' : 'Token Invalid/Expired'}
                      color={status.token_valid ? 'success' : 'warning'}
                      variant="outlined"
                    />
                    <Chip
                      label={status.refresher_active ? 'Auto-refresh Active' : 'Auto-refresh Inactive'}
                      color={status.refresher_active ? 'info' : 'default'}
                      variant="outlined"
                    />
                  </Box>

                  {status.token_expiry && (
                    <Typography variant="body2" color="text.secondary" gutterBottom>
                      Token Expiry: {new Date(status.token_expiry).toLocaleString()}
                    </Typography>
                  )}

                  {status.last_refresh_at && (
                    <Typography variant="body2" color="text.secondary" gutterBottom>
                      Last Refresh: {new Date(status.last_refresh_at).toLocaleString()}
                    </Typography>
                  )}

                  {status.last_error && (
                    <Alert severity="warning" sx={{ mt: 2, mb: 2 }}>
                      Last Error: {status.last_error}
                    </Alert>
                  )}

                  <Box sx={{ mt: 4 }}>
                    <Typography variant="h6" gutterBottom>
                      Re-authorize Gmail Access
                    </Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                      If your OAuth token has expired or you need to re-authorize, click the button below to start
                      the Google authorization flow. This will open a new window where you can sign in with your
                      Google account and grant access to send emails.
                    </Typography>
                    <Button
                      variant="contained"
                      onClick={() => setConfirmOpen(true)}
                      disabled={!canManage || authorizing || !status.configured}
                    >
                      {authorizing ? 'Opening...' : 'Authorize with Google'}
                    </Button>
                    {!status.configured && (
                      <Typography variant="body2" color="error" sx={{ mt: 1 }}>
                        OAuth credentials file not found. Please configure credentials.json first.
                      </Typography>
                    )}
                  </Box>
                </Box>
              ) : null}
            </LayoutCard>
          </Box>
        </Grid>
      </Box>

      <Dialog open={confirmOpen} onClose={() => setConfirmOpen(false)}>
        <DialogTitle>Confirm Re-authorization</DialogTitle>
        <DialogContent>
          <Typography>
            This will open a new window to authorize with Google. After completing the authorization,
            you may need to refresh this page to see the updated status.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setConfirmOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleAuthorize}>
            Continue
          </Button>
        </DialogActions>
      </Dialog>
    </PageContainer>
  );
}
```

**Step 2: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/pages/admin/OAuthPage.tsx
git commit -m "feat: add OAuth management UI page"
```

---

### Task 5.3: Add OAuth Route and Navigation

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/navigation/AppRoutes.tsx`
- Modify: `sensor_hub/ui/sensor_hub_ui/src/navigation/NavigationSidebar.tsx`

**Step 1: Add route in AppRoutes.tsx**

Add import at top:

```tsx
import OAuthPage from "../pages/admin/OAuthPage.tsx";
```

Add route after line 27 (after RolesPage route):

```tsx
<Route path="/admin/oauth" element={<RequireAuth><OAuthPage /></RequireAuth>} />
```

**Step 2: Add navigation item in NavigationSidebar.tsx**

Add import at top (after existing icons):

```tsx
import VpnKeyIcon from '@mui/icons-material/VpnKey';
```

Add navigation item after line 137 (after Manage roles item), inside the admin section:

```tsx
{ hasPerm(user, 'manage_oauth') && (
  <ListItem disablePadding>
    <ListItemButton onClick={() => handleNavigate('/admin/oauth')}>
      <ListItemIcon><VpnKeyIcon /></ListItemIcon>
      <ListItemText primary="OAuth" />
    </ListItemButton>
  </ListItem>
) }
```

**Step 3: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/navigation/AppRoutes.tsx sensor_hub/ui/sensor_hub_ui/src/navigation/NavigationSidebar.tsx
git commit -m "feat: add OAuth page to navigation and routes"
```

---

## Phase 6: Final Integration and Testing

### Task 6.1: Run All Tests

**Step 1: Run Go tests**

Run: `cd sensor_hub && go test ./... -v`
Expected: All tests pass

**Step 2: Build the application**

Run: `cd sensor_hub && go build ./...`
Expected: Build succeeds

**Step 3: Run UI build**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

**Step 4: Final commit**

```bash
git add -A
git commit -m "feat: complete OAuth refactor with UI management"
```

---

## Summary

This plan implements:

1. **Application Properties** - Configurable OAuth file paths and refresh interval
2. **OAuth Package Refactor** - Testable service with dependency injection, controllable lifecycle
3. **Database Migration** - New `manage_oauth` permission
4. **API Endpoints** - Status, authorize, and callback endpoints with RBAC protection
5. **UI Page** - OAuth management page with Google consent flow integration

All code follows existing project patterns:
- Uses testify/mock for mocking
- Uses gin for HTTP handlers
- Uses MUI components for UI
- Uses existing PageContainer/LayoutCard patterns
- Follows existing permission-based access control
