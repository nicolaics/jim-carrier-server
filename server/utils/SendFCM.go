package utils

import (
	"context"
	"fmt"

	// "log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/nicolaics/jim-carrier/config"
	"google.golang.org/api/option"
)

func SendFCMToOne(toToken, title, body string) (string, error) {
	if toToken == "" {
		return "", fmt.Errorf("destination token must be specified")
	}

	cfg := option.WithCredentialsFile(config.Envs.GoogleApplicationCredentialsPath)

	app, err := firebase.NewApp(context.Background(), nil, cfg)
	if err != nil {
		// log.Printf("error initializing app: %v\n", err)
		return "", err
	}

	ctx := context.Background()
	client, err := app.Messaging(ctx)
	if err != nil {
		// log.Printf("error getting Messaging client: %v\n", err)
		return "", err
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Token: toToken,
	}

	response, err := client.Send(ctx, message)
	if err != nil {
		return "", err
	}

	return response, nil
}