package types

import (
	"net/http"
	"time"
)

type UserStore interface {
	GetUserByEmail(string) (*User, error)
	GetUserByName(string) (*User, error)
	GetUserByID(int) (*User, error)
	GetUserPasswordByEmail(string) (string, error)

	GetUserBySearchName(string) ([]User, error)
	GetUserBySearchPhoneNumber(string) ([]User, error)

	CreateUser(User) error

	DeleteUser(*User) error

	UpdateLastLoggedIn(int) error
	ModifyUser(int, User) error
	UpdatePassword(int, string) error
	UpdateProfilePicture(id int, profPicUrl string) error

	SaveToken(userId int, accessTokenDetails *TokenDetails, refreshTokenDetails *TokenDetails) error
	DeleteToken(int) error
	ValidateUserAccessToken(http.ResponseWriter, *http.Request) (*User, error)
	ValidateUserRefreshToken(refreshToken string) (*User, error)
	UpdateAccessToken(userId int, accessTokenDetails *TokenDetails) error
	IsAccessTokenExist(userId int) (bool, error)

	DelayCodeWithinTime(email string, minutes int) (bool, error)
	SaveVerificationCode(email, code string, requestType int) error
	ValidateLoginCodeWithinTime(email, code string, minutes, reqType int) (bool, error)
	UpdateVerificationCodeStatus(email string, status, reqType int) error

	IsUserExist(email string) (bool, error)

	CheckProvider(email string) (bool, string, error)

	UpdateFCMToken(id int, fcmToken string) error

	IsDeleteUserAllowed(id int) (bool, error)
}

// register new user
type RegisterUserPayload struct {
	Name             string `json:"name" validate:"required"`
	Email            string `json:"email" validate:"required,email"`
	Password         string `json:"password" validate:"required,min=3,max=130"`
	PhoneNumber      string `json:"phoneNumber" validate:"required"`
	ProfilePicture   []byte `json:"profilePicture"`
	FCMToken         string `json:"fcmToken"`
	VerificationCode string `json:"verificationCode" validate:"required"`
}

type RegisterGooglePayload struct {
	IDToken           string `json:"idToken" validate:"required"`
	ServerAuthCode    string `json:"serverAuthCode" validate:"required"`
	FCMToken          string `json:"fcmToken"`
	Name              string `json:"name" validate:"required"`
	PhoneNumber       string `json:"phoneNumber" validate:"required"`
	ProfilePictureURL string `json:"profilePictureUrl"`
}

// modify the data of the user
type ModifyUserPayload struct {
	Name        string `json:"name" validate:"required"`
	PhoneNumber string `json:"phoneNumber" validate:"required"`
}

// get one user data
type GetOneUserPayload struct {
	ID int `json:"id" validate:"required"`
}

// normal log-in
type LoginUserPayload struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	FCMToken   string `json:"fcmToken"`
}

type LoginGooglePayload struct {
	IDToken        string `json:"idToken" validate:"required"`
	ServerAuthCode string `json:"serverAuthCode" validate:"required"`
	FCMToken       string `json:"fcmToken"`
}	

// request verification code payload
type UserVerificationCodePayload struct {
	Email string `json:"email" validate:"required,email"`
}

type UpdatePasswordPayload struct {
	OldPassword string `json:"oldPassword" validate:"required,min=3,max=130"`
	NewPassword string `json:"newPassword" validate:"required,min=3,max=130"`
}

type UpdateProfilePicturePayload struct {
	ProfilePicture []byte `json:"profilePicture"`
}

type ResetPasswordPayload struct {
	Email       string `json:"email" validate:"required,email"`
	NewPassword string `json:"newPassword" validate:"required,min=3,max=130"`
}

type VerifyVerificationCodePayload struct {
	Email            string `json:"email" validate:"required,email"`
	VerificationCode string `json:"verificationCode" validate:"required"`
}

type RefreshTokenPayload struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type AutoLoginPayload struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
	FCMToken     string `json:"fcmToken"`
}

type ReturnUserPayload struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	PhoneNumber    string    `json:"phoneNumber"`
	Provider       string    `json:"provider"`
	ProfilePicture []byte    `json:"profilePicture"`
	FCMToken       string    `json:"fcmToken"`
	LastLoggedIn   time.Time `json:"lastLoggedIn"`
	CreatedAt      time.Time `json:"createdAt"`
}

// basic user data info
type User struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	Password          string    `json:"password"`
	PhoneNumber       string    `json:"phoneNumber"`
	Provider          string    `json:"provider"`
	ProfilePictureURL string    `json:"profilePictureURL"`
	FCMToken          string    `json:"fcmToken"` // Firebase Cloud Messaging for notification
	LastLoggedIn      time.Time `json:"lastLoggedIn"`
	CreatedAt         time.Time `json:"createdAt"`
}
