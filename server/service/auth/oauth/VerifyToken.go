package oauth

import (
	"github.com/nicolaics/jim-carrier/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

// Google OAuth2 Config
var GoogleOauthConfig = &oauth2.Config{
	ClientID:     config.Envs.GoogleClientID,
	ClientSecret: config.Envs.GoogleClientSecret,
	Endpoint:     google.Endpoint,
	RedirectURL:  "ReURL", // 환경 변수에서 리디렉션 URL을 가져옵니다.
}

func VerifyIDToken(idToken string) (*idtoken.Payload, error) {
	// ID 토큰 검증
	payload, err := idtoken.Validate(oauth2.NoContext, idToken, GoogleOauthConfig.ClientID)
	if err != nil {
		return nil, err
	}
	
	return payload, nil
}

//  {
//    "access_token": "ya29.a0AfB_byBH_v_h2meyb_7pPScVR1SdtzQUE0JMcABLUDi-vZLfsVOGFtq_Ka3oz87fwpPjNT6A-IGW7a5woAKBs5yOza85P-eGrnv7pCPOwWfW1CSvV9JN-oRvVwkDFSs-7LtWZH30Aafi8Ata1TrxU0Hl7YF4R-h9jcUaCgYKAR0SARMSFQHGX2Mi8nS1YDw1toKOv3ttdzT0yg0170",
//    "token_type": "Bearer",
//    "expires_in": 3599,
//    "scope": "email profile openid https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile",
//    "authuser": "2",
//    "prompt": "none"
//}
