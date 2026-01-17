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
