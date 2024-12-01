package utils

import (
	"context"
	"encoding/json"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/nicolaics/jim-carrier/config"
	"github.com/nicolaics/jim-carrier/types"
	"google.golang.org/api/option"
)

func SendFCMToOne(fcm types.FCMHistory) (string, error) {
	if fcm.ToToken == "" {
		return "", fmt.Errorf("destination token must be specified")
	}

	cfg := option.WithCredentialsFile(config.Envs.GoogleApplicationCredentialsPath)

	app, err := firebase.NewApp(context.Background(), nil, cfg)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	client, err := app.Messaging(ctx)
	if err != nil {
		return "", err
	}

	var fcmData map[string]string

	fcmDataMarshal, _ := json.Marshal(fcm.Data)
	json.Unmarshal(fcmDataMarshal, &fcmData)

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: fcm.Title,
			Body:  fcm.Body,
		},
		Token: fcm.ToToken,
		Data: fcmData,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
	}

	response, err := client.Send(ctx, message)
	if err != nil {
		return "", err
	}

	return response, nil
}