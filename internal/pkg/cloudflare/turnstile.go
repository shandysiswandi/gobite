package cloudflare

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type Turnstile interface {
	Verify(ctx context.Context, token string) (bool, error)
}

type CFTurnstile struct {
	siteVerifyURL string
	secret        string
	client        *http.Client
}

func NewTurnstile(siteURL, secret string, client *http.Client) *CFTurnstile {
	return &CFTurnstile{
		siteVerifyURL: siteURL,
		secret:        secret,
		client:        client,
	}
}

func (cf *CFTurnstile) Verify(ctx context.Context, token string) (bool, error) {
	form := url.Values{}
	form.Set("secret", cf.secret)
	form.Set("response", token)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cf.siteVerifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := cf.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var respBody map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return false, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, nil
	}

	return true, nil
}
