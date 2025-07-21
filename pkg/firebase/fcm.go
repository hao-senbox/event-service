package firebase

import (
	"context"
	"path/filepath"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

func SetUpFireBase() (*firebase.App, context.Context, *messaging.Client) {

	ctx := context.Background()
	
	serviceAccountKeyFilePath, err := filepath.Abs("./credentials/fcm_firebase.json")
    if err != nil {
        panic("Unable to load serviceAccountKeys.json file")
    }
	
	opt := option.WithCredentialsFile(serviceAccountKeyFilePath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		panic(err)
	}

	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		panic(err)
	}

	return app, ctx, messagingClient

}