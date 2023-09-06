package main

import (
	"log"
	"net/http"

	"github.com/agustfricke/crud-auth0-htmx/auth"
	"github.com/agustfricke/crud-auth0-htmx/database"
	"github.com/agustfricke/crud-auth0-htmx/router"
	"github.com/joho/godotenv"
)

func main() {

    database.ConnectDB()

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	auth, err := auth.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	rtr := router.New(auth)

	log.Print("Server listening on http://192.168.1.51:3000/")
	if err := http.ListenAndServe("0.0.0.0:3000", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
