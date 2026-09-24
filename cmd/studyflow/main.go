package main

import (
	"log"
	"net/http"
	"os"

	"github.com/nerfmener/studyflow-go/internal/store"
)

func main() {
	dataFile := os.Getenv("STUDYFLOW_DATA")
	if dataFile == "" {
		dataFile = "data/studyflow.json"
	}

	studyStore, err := store.New(dataFile)
	if err != nil {
		log.Fatal(err)
	}

	server := NewServer(studyStore)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("StudyFlow is running at http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, server.Routes()))
}
