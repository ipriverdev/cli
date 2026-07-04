package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ipriverdev/cli/internal/api"
	"github.com/ipriverdev/cli/internal/config"
	keyring "github.com/zalando/go-keyring"
)

type CredentialStore interface {
	Load() (Credentials, error)
	Save(creds Credentials) error
	Delete() error
}

var DefaultStore CredentialStore = &keyringFileStore{}

const (
	keyringService  = "ipriver-cli"
	keyringUser     = "default"
	credentialsFile = "credentials.json"
)

type Credentials struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	RefreshURI   string `json:"refresh_uri,omitempty"`
}

type keyringFileStore struct{}

func (s *keyringFileStore) Save(creds Credentials) error {
	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	if err := keyring.Set(keyringService, keyringUser, string(data)); err == nil {
		_ = deleteCredentialsFile()
		return nil
	}

	return saveCredentialsFile(creds)
}

func (s *keyringFileStore) Load() (Credentials, error) {
	data, err := keyring.Get(keyringService, keyringUser)
	if err == nil {
		var creds Credentials
		if err := json.Unmarshal([]byte(data), &creds); err != nil {
			return Credentials{}, fmt.Errorf("parse keyring credentials: %w", err)
		}
		return creds, nil
	}

	return loadCredentialsFile()
}

func (s *keyringFileStore) Delete() error {
	_ = keyring.Delete(keyringService, keyringUser)
	return deleteCredentialsFile()
}

func SaveCredentials(creds Credentials) error {
	return DefaultStore.Save(creds)
}

func LoadCredentials() (Credentials, error) {
	return DefaultStore.Load()
}

func DeleteCredentials() error {
	return DefaultStore.Delete()
}

func GetToken() (string, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return "", err
	}
	if creds.Token == "" {
		return "", errors.New("no token stored")
	}
	return creds.Token, nil
}

func NewAuthenticatedClient(host string) (*api.Client, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, fmt.Errorf("not logged in: %w", err)
	}
	if creds.Token == "" {
		return nil, errors.New("no token stored")
	}

	token := creds.Token

	if tokenExpiresWithin(token, 2*time.Minute) && creds.RefreshToken != "" {
		if newToken, err := RefreshAccessToken(context.Background(), host); err == nil {
			token = newToken
		}
	}

	client := api.New(host).
		WithToken(token).
		WithAutoRefresh(func(ctx context.Context) (string, error) {
			return RefreshAccessToken(ctx, host)
		})

	return client, nil
}

func tokenExpiresWithin(token string, margin time.Duration) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return true
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return true
	}

	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp == 0 {
		return true
	}

	return time.Now().Add(margin).Unix() > claims.Exp
}

func RefreshAccessToken(ctx context.Context, host string) (string, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return "", fmt.Errorf("load credentials for refresh: %w", err)
	}
	if creds.RefreshToken == "" {
		return "", errors.New("no refresh token available, run `ipriver login`")
	}

	refreshPath := "/api/token/refresh"
	if creds.RefreshURI != "" {
		if strings.HasPrefix(creds.RefreshURI, "http") {
			host = creds.RefreshURI
			refreshPath = ""
		} else {
			refreshPath = creds.RefreshURI
		}
	}

	client := api.New(host)

	var resp struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := client.Post(ctx, refreshPath, map[string]string{
		"refresh_token": creds.RefreshToken,
	}, &resp); err != nil {
		return "", fmt.Errorf("refresh token: %w", err)
	}

	if resp.Token == "" {
		return "", errors.New("refresh returned empty token")
	}

	creds.Token = resp.Token
	if resp.RefreshToken != "" {
		creds.RefreshToken = resp.RefreshToken
	}

	if err := SaveCredentials(creds); err != nil {
		return "", fmt.Errorf("save refreshed credentials: %w", err)
	}

	return creds.Token, nil
}

func credentialsPath() (string, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, credentialsFile), nil
}

func saveCredentialsFile(creds Credentials) error {
	dir, err := config.ConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	path, err := credentialsPath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}

	return nil
}

func loadCredentialsFile() (Credentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return Credentials{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Credentials{}, errors.New("not logged in")
		}
		return Credentials{}, fmt.Errorf("read credentials: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return Credentials{}, fmt.Errorf("parse credentials: %w", err)
	}

	return creds, nil
}

func deleteCredentialsFile() error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete credentials: %w", err)
	}

	return nil
}

type User struct {
	UUID     string `json:"uuid"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

func CurrentUser(ctx context.Context, host string) (*User, error) {
	client, err := NewAuthenticatedClient(host)
	if err != nil {
		return nil, err
	}

	var user User
	if err := client.Get(ctx, "/api/user", &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func DisplayName(user *User) string {
	if user == nil {
		return ""
	}
	if user.FullName != "" {
		return user.FullName
	}
	return user.Email
}
