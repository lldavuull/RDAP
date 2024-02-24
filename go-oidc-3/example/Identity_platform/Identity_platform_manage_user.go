package main

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4" //引入firebase
)

func main() {
	ctx := context.Background() //設定context

	app, err := firebase.NewApp(context.Background(), nil) //初始化firebase
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	// Get an auth client from the firebase.App
	client, err := app.Auth(ctx) //初始化client
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}
	// Set admin privilege on the user corresponding to uid.
	claims := map[string]interface{}{"admin": true}
	err = client.SetCustomUserClaims(ctx, "PeT55jzck8fe8HgMyhUxPW1h9Pm2", claims)
	if err != nil {
		log.Fatalf("error setting custom claims %v\n", err)
	}

	// The new custom claims will propagate to the user's ID token the
	// next time a new one is issued.

	user, err := client.GetUserByEmail(ctx, "lldavuull1@gmail.com")
	if err != nil {
		log.Fatal(err)
	}
	// Add incremental custom claim without overwriting existing claims.
	currentCustomClaims := user.CustomClaims
	if currentCustomClaims == nil {
		currentCustomClaims = map[string]interface{}{}
	}

	if _, found := currentCustomClaims["admin"]; found {
		// Add level.
		currentCustomClaims["accessLevel"] = 10
		// Add custom claims for additional privileges.
		err := client.SetCustomUserClaims(ctx, user.UID, currentCustomClaims)
		if err != nil {
			log.Fatalf("error setting custom claims %v\n", err)
		}
	}
}
