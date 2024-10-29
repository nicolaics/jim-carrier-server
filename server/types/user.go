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

	SaveToken(int, *TokenDetails) error
	DeleteToken(int) error
	ValidateUserToken(http.ResponseWriter, *http.Request) (*User, error)

	DelayCodeWithinTime(email string, minutes int) (bool, error)
	SaveVerificationCode(email, code string, requestType int) error
	ValidateLoginCodeWithinTime(email, code string, minutes int) (bool, error)
	UpdateVerificationCodeStatus(email string, status int) error

	IsUserExist(email string) (bool, error)
}

// register new user
type RegisterUserPayload struct {
	Name                   string `json:"name" validate:"required"`
	Email                  string `json:"email" validate:"required,email"`
	Password               string `json:"password" validate:"required,min=3,max=130"`
	PhoneNumber            string `json:"phoneNumber" validate:"required"`
	ProfilePicture         []byte `json:"profilePicture"`
	VerificationCode       string `json:"verificationCode" validate:"required"`
}

// delete user account
type RemoveUserPayload struct {
	ID int `json:"id" validate:"required"`
}

// modify the data of the user
type ModifyUserPayload struct {
	ID          int    `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	PhoneNumber string `json:"phoneNumber" validate:"required"`
}

// get one user data
type GetOneUserPayload struct {
	ID int `json:"id" validate:"required"`
}

// normal log-in
type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginUsingGooglePayload struct {
	IDToken        string `json:"idToken" validate:"required"`
	ServerAuthCode string `json:"serverAuthCode" validate:"required"`
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
	ProfilePictureFileName string `json:"profilePictureFileName" validate:"required"`
	ProfilePicture         []byte `json:"profilePicture" validate:"required"`
}

type ResetPasswordPayload LoginUserPayload

// basic user data info
type User struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	Password          string    `json:"password"`
	PhoneNumber       string    `json:"phoneNumber"`
	Provider          string    `json:"provider"`
	ProfilePictureURL string    `json:"profilePictureURL"`
	LastLoggedIn      time.Time `json:"lastLoggedIn"`
	CreatedAt         time.Time `json:"createdAt"`
}
