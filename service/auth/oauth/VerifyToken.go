package oauth

import (
	"context"

	"github.com/nicolaics/jim-carrier-server/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

// Google OAuth2 Config
var GoogleOauthConfig = &oauth2.Config{
	ClientID:     config.Envs.GoogleClientID,
	ClientSecret: config.Envs.GoogleClientSecret,
	Endpoint:     google.Endpoint,
	RedirectURL:  "ReURL", 
}

func VerifyIDToken(idToken string) (*idtoken.Payload, error) {
	payload, err := idtoken.Validate(context.Background(), idToken, GoogleOauthConfig.ClientID)
	if err != nil {
		return nil, err
	}
	
	return payload, nil
}
