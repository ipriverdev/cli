package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/ipriverdev/cli/internal/api"
	"github.com/pkg/browser"
)

const (
	deviceCodePath  = "/api/device/code"
	deviceTokenPath = "/api/device/token"
)

type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
	RefreshURI              string `json:"refresh_uri"`
}

type LoginResult struct {
	Credentials Credentials
	User        *User
}

type DeviceFlow struct {
	host   string
	client *api.Client
}

func NewDeviceFlow(host string) *DeviceFlow {
	return &DeviceFlow{
		host:   host,
		client: api.New(host),
	}
}

func (f *DeviceFlow) Login(ctx context.Context) (*LoginResult, error) {
	deviceCode, err := f.requestDeviceCode(ctx)
	if err != nil {
		return nil, err
	}

	if err := f.openBrowser(deviceCode); err != nil {
		return nil, err
	}

	creds, err := f.pollForToken(ctx, deviceCode)
	if err != nil {
		return nil, err
	}

	if err := SaveCredentials(*creds); err != nil {
		return nil, fmt.Errorf("store credentials: %w", err)
	}

	user, _ := CurrentUser(ctx, f.host)
	if user == nil {
		return &LoginResult{Credentials: *creds}, nil
	}

	return &LoginResult{Credentials: *creds, User: user}, nil
}

func (f *DeviceFlow) requestDeviceCode(ctx context.Context) (*DeviceCodeResponse, error) {
	var resp DeviceCodeResponse
	if err := f.client.Post(ctx, deviceCodePath, nil, &resp); err != nil {
		return nil, fmt.Errorf("request device code: %w", err)
	}

	if resp.DeviceCode == "" || resp.UserCode == "" {
		return nil, errors.New("invalid device code response from API")
	}

	if resp.Interval <= 0 {
		resp.Interval = 5
	}
	if resp.ExpiresIn <= 0 {
		resp.ExpiresIn = 900
	}

	return &resp, nil
}

func (f *DeviceFlow) openBrowser(deviceCode *DeviceCodeResponse) error {
	uri := deviceCode.VerificationURIComplete
	if uri == "" {
		uri = deviceCode.VerificationURI
	}
	if uri == "" {
		return errors.New("verification URI missing from API response")
	}

	if !strings.Contains(uri, "user_code=") {
		separator := "?"
		if strings.Contains(uri, "?") {
			separator = "&"
		}
		uri = fmt.Sprintf("%s%suser_code=%s", uri, separator, deviceCode.UserCode)
	}

	fmt.Printf("\n! Your one-time code: %s\n", deviceCode.UserCode)

	if err := browser.OpenURL(uri); err != nil {
		fmt.Printf("- Could not open browser automatically. Visit %s\n", uri)
	} else {
		fmt.Printf("- Opening %s\n", uri)
	}

	fmt.Println()
	return nil
}

func (f *DeviceFlow) pollForToken(ctx context.Context, deviceCode *DeviceCodeResponse) (*Credentials, error) {
	interval := time.Duration(deviceCode.Interval) * time.Second
	deadline := time.Now().Add(time.Duration(deviceCode.ExpiresIn) * time.Second)

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Waiting for authorization..."
	s.Start()
	defer s.Stop()

	for {
		if time.Now().After(deadline) {
			return nil, errors.New("authorization timed out, run `ipriver login` again")
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}

		creds, done, err := f.exchangeDeviceCode(ctx, deviceCode)
		if err != nil {
			return nil, err
		}
		if !done {
			continue
		}
		return creds, nil
	}
}

func (f *DeviceFlow) exchangeDeviceCode(ctx context.Context, dc *DeviceCodeResponse) (*Credentials, bool, error) {
	var resp struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		Message      string `json:"message"`
	}

	status, err := f.client.PostStatus(ctx, deviceTokenPath, map[string]string{
		"device_code": dc.DeviceCode,
	}, &resp)

	if err != nil {
		var apiErr *api.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.Status {
			case http.StatusForbidden:
				return nil, false, errors.New("authorization denied")
			case http.StatusBadRequest:
				return nil, false, fmt.Errorf("login failed: %s", apiErr.Message)
			}
		}
		return nil, false, fmt.Errorf("poll for token: %w", err)
	}

	if status == http.StatusAccepted {
		return nil, false, nil
	}

	if resp.Token == "" {
		return nil, false, errors.New("server returned empty token")
	}

	creds := &Credentials{
		Token:        resp.Token,
		RefreshToken: resp.RefreshToken,
		RefreshURI:   dc.RefreshURI,
	}
	return creds, true, nil
}
