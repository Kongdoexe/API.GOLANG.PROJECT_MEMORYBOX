package firebase

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var MessagingClient *messaging.Client

func InitFirebase() {
	opt := option.WithCredentialsFile("serviceAccountKey.json")

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v\n", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}

	MessagingClient = client
	log.Println("Firebase initialized successfully")
}
