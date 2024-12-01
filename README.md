# Jim Carrier API Server

## Directory Structure

```
.
├── cmd
|   ├── api
|   ├── migrate
|   └── main.go
├── config
|   └── env.go
├── constants
|   └── constants.go
├── db
|   └── db.go
├── logger
|   ├── logger.go
|   └── WriteLog.go
├── service
|   ├── auth
|   |   ├── jwt
|   |   |   ├── jwt.go
|   |   |   └── JwtMiddleware.go
|   |   ├── oauth
|   |   |   └── VerifyAccessToken.go
|   |   ├── CorsMiddleware.go
|   |   └── password.go
|   ├── bank
|   |   └── store.go
|   ├── currency
|   |   └── store.go
|   ├── fcm
|   |   └── store.go
|   ├── listing
|   |   ├── routes.go
|   |   └── store.go
|   ├── order
|   |   ├── routes.go
|   |   └── store.go
|   ├── review
|   |   ├── routes.go
|   |   └── store.go
|   └── user
|   |   ├── routes.go
|   |   └── store.go
├── types
|   ├── bank.go
|   ├── currency.go
|   ├── fcm.go
|   ├── listing.go
|   ├── order.go
|   ├── review.go
|   ├── types.go
|   └── user.go
├── utils
|   ├── GetImage.go
|   ├── ParamsIntStringConversion.go
|   ├── ParseDate.go
|   ├── SaveImage.go
|   ├── SendEmail.go
|   ├── SendFCM.go
|   ├── utils.go
|   └── WriteJson.go
├── .gitignore
├── go.mod
├── go.sum
├── LICENSE
├── Makefile
└── README.md
```

## Installation

To run the server, make sure you have go and makefile installed

1. Run `go mod download`
2. Run `make run` - to start running the server
