package oauth

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// https://developers.facebook.com/docs/facebook-login/guides/advanced/manual-flow
type facebook struct {
	opt *Options
}

func (fb *facebook) AuthorizeURL(redirectURI string) *url.URL {
	query := url.Values{}
	query.Set("client_id", fb.opt.ClientID)
	query.Set("response_type", "code")
	query.Set("redirect_uri", redirectURI)

	authURL, _ := url.Parse("https://www.facebook.com/v14.0/dialog/oauth")
	authURL.RawQuery = query.Encode()

	return authURL
}

func (fb *facebook) Authorize(code, redirectURI string) (*User, error) {
	accessToken, err := fb.requestAccessToken(code, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("get access token, %w", err)
	}

	userID, err := fb.requestUserID(accessToken)
	if err != nil {
		return nil, fmt.Errorf("get user id, %w", err)
	}

	return &User{
		AccessToken: accessToken,
		ID:          userID,
	}, nil
}

func (fb *facebook) Vendor() string {
	return "facebook"
}

func (fb *facebook) requestAccessToken(code, redirectURI string) (string, error) {
	query := url.Values{}
	query.Set("client_id", fb.opt.ClientID)
	query.Set("client_secret", fb.opt.ClientSecret)
	query.Set("redirect_uri", redirectURI)
	query.Set("code", code)

	requestURL, _ := url.Parse("https://graph.facebook.com/v14.0/oauth/access_token")
	requestURL.RawQuery = query.Encode()

	response, err := httpClient.Get(requestURL.String())
	if err != nil {
		return "", fmt.Errorf("send request, %w", err)
	}
	defer response.Body.Close()
	if code := response.StatusCode; code > 299 {
		return "", fmt.Errorf("response status %d", code)
	}

	body := struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}{}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode response, %w", err)
	}
	return body.AccessToken, nil
}

func (fb *facebook) requestUserID(accessToken string) (string, error) {
	query := url.Values{}
	query.Set("input_token", accessToken)
	query.Set("access_token", fmt.Sprintf("%s|%s", fb.opt.ClientID, fb.opt.ClientSecret))

	requestURL, _ := url.Parse("https://graph.facebook.com/debug_token")
	requestURL.RawQuery = query.Encode()

	response, err := httpClient.Get(requestURL.String())
	if err != nil {
		return "", fmt.Errorf("send request, %w", err)
	}
	defer response.Body.Close()
	if code := response.StatusCode; code > 299 {
		return "", fmt.Errorf("response status %d", code)
	}

	body := struct {
		Data struct {
			AppID   string `json:"app_id"`
			Type    string `json:"type"`
			IsValid bool   `json:"is_valid"`
			UserID  string `json:"user_id"`
		} `json:"data"`
	}{}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode response, %w", err)
	} else if !body.Data.IsValid {
		return "", fmt.Errorf("invalid token, %w", err)
	}
	return body.Data.UserID, nil
}
