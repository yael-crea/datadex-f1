package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"datadex-f1/api"
	"datadex-f1/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	client := api.NewClient()
	// Préchargement en arrière-plan : le serveur démarre tout de suite.
	go client.Preload()

	app, err := handlers.New(client, "templates")
	if err != nil {
		log.Fatalf("impossible de charger les templates : %v", err)
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           app.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      60 * time.Second,
	}
	log.Printf("🏁 DataDex F1 démarré sur http://localhost:%s", port)
	log.Fatal(srv.ListenAndServe())
}
