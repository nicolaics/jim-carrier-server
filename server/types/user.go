package types

import (
	"net/http"
	"time"
)

type UserStore interface {
	GetUserByEmail(string) (*User, error)
	GetUserByName(string) (*User, error)
	GetUserByID(int) (*User, error)

	GetUserBySearchName(string) ([]User, error)
	GetUserBySearchPhoneNumber(string) ([]User, error)

	CreateUser(User) error

	DeleteUser(*User) error

	UpdateLastLoggedIn(int) error
	ModifyUser(int, User) error

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
	Name             string `json:"name" validate:"required"`
	Email            string `json:"email" validate:"required,email"`
	Password         string `json:"password" validate:"required,min=3,max=130"`
	PhoneNumber      string `json:"phoneNumber" validate:"required"`
	VerificationCode string `json:"verificationCode" validate:"required"`
}

// delete user account
type RemoveUserPayload struct {
	ID int `json:"id" validate:"required"`
}

// modify the data of the user
type ModifyUserPayload struct {
	ID          int    `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Password    string `json:"password" validate:"required,min=3,max=130"`
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

// request verification code payload
type UserVerificationCodePayload struct {
	Email string `json:"email" validate:"required,email"`
}

// basic user data info
type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Password     string    `json:"password"`
	PhoneNumber  string    `json:"phoneNumber"`
	Provider     string    `json:"provider"`
	LastLoggedIn time.Time `json:"lastLoggedIn"`
	CreatedAt    time.Time `json:"createdAt"`
}
