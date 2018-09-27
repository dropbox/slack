package slack

import (
	"context"
	"errors"
	"net/url"
)

type OAuthResponseIncomingWebhook struct {
	URL              string `json:"url"`
	Channel          string `json:"channel"`
	ChannelID        string `json:"channel_id,omitempty"`
	ConfigurationURL string `json:"configuration_url"`
}

type OAuthResponseBot struct {
	BotUserID      string `json:"bot_user_id"`
	BotAccessToken string `json:"bot_access_token"`
}

type OAuthResponse struct {
	AccessToken     string                       `json:"access_token"`
	RefreshToken	string					     `json:"refresh_token"`
	Scope           string                       `json:"scope"`
	TeamName        string                       `json:"team_name"`
	TeamID          string                       `json:"team_id"`
	IncomingWebhook OAuthResponseIncomingWebhook `json:"incoming_webhook"`
	Bot             OAuthResponseBot             `json:"bot"`
	UserID          string                       `json:"user_id,omitempty"`
	SlackResponse
}

// GetOAuthToken retrieves an AccessToken
func GetOAuthToken(clientID, clientSecret, code, redirectURI string, debug bool) (accessToken string, scope string, err error) {
	return GetOAuthTokenContext(context.Background(), clientID, clientSecret, code, redirectURI, debug)
}

// GetOAuthToken retrieves an AccessToken
func GetOAuthTokenWithRefresh(clientID, clientSecret, code, redirectURI string, debug bool) (accessToken string, refreshToken string, teamId string, scope string, err error) {
	return GetOAuthTokenContextWithRefresh(context.Background(), clientID, clientSecret, code, redirectURI, debug)
}

// GetOAuthTokenContext retrieves an AccessToken with a custom context
func GetOAuthTokenContext(ctx context.Context, clientID, clientSecret, code, redirectURI string, debug bool) (accessToken string, scope string, err error) {
	response, err := GetOAuthResponseContext(ctx, clientID, clientSecret, code, redirectURI, debug)
	if err != nil {
		return "", "", err
	}
	return response.AccessToken, response.Scope, nil
}

// GetOAuthTokenContext retrieves an AccessToken with a custom context
func GetOAuthTokenContextWithRefresh(ctx context.Context, clientID, clientSecret, code, redirectURI string, debug bool) (accessToken string, refreshToken string, teamId string, scope string, err error) {
	response, err := GetOAuthResponseContext(ctx, clientID, clientSecret, code, redirectURI, debug)
	if err != nil {
		return "", "", "", "", err
	}
	return response.AccessToken, response.RefreshToken, response.TeamID, response.Scope, nil
}

func GetOAuthResponse(clientID, clientSecret, code, redirectURI string, debug bool) (resp *OAuthResponse, err error) {
	return GetOAuthResponseContext(context.Background(), clientID, clientSecret, code, redirectURI, debug)
}

func GetOAuthResponseContext(ctx context.Context, clientID, clientSecret, code, redirectURI string, debug bool) (resp *OAuthResponse, err error) {
	values := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}
	response := &OAuthResponse{}
	err = postSlackMethod(ctx, customHTTPClient, "oauth.access", values, response, debug)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, errors.New(response.Error)
	}
	return response, nil
}

func RefreshToken(ctx context.Context, refreshConfig RefreshTokenConfig, debug bool) (string, error) {
	values := url.Values{
		"client_id":     {refreshConfig.ClientId},
		"client_secret": {refreshConfig.ClientSecret},
		"refresh_token": {refreshConfig.RefreshToken},
		"grant_type": {"refresh_token"},
	}
	response := &OAuthResponse{}
	err := postSlackMethod(ctx, customHTTPClient, "oauth.access", values, response, debug)
	if err != nil {
		return "", err
	}
	if !response.Ok {
		return "", errors.New(response.Error)
	}

	if refreshConfig.internalCallback != nil {
		updateArgs := AuthTokenUpdateArgs{
			TeamId: response.TeamID,
			AccessToken: response.AccessToken,
		}
		refreshConfig.internalCallback(updateArgs)
	}

	return response.AccessToken, nil
}
