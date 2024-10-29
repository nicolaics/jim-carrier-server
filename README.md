# Jim Carrier

## Directory Structure
```
.
├── .gitignore
├── LICENSE
├── README.md
├── frontend
|   ├── android
|   ├── assets
|   ├── ios
|   ├── lib
|   ├── linux
|   ├── macos
|   ├── test
|   ├── web
|   ├── windows
|   ├── .metadata
|   ├── analysis_options.yaml
|   ├── pubspec.lock
|   └── pubspec.yaml
└── server
    ├── cmd
    |   ├── api
    |   ├── migrate
    |   └── main.go
    ├── config
    ├── constants
    ├── db
    ├── logger
    ├── service
    |   ├── auth
    |   ├── listing
    |   ├── order
    |   └── user
    ├── types
    ├── utils
    └── Makefile
```

## Server

To run the server, make sure you have go and makefile installed

1. Run `go mod install`
2. Run `make run` - to start running the server
