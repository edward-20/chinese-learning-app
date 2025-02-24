package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/edward-20/chinese-learning-app/database"
	handler "github.com/edward-20/chinese-learning-app/handlers"
)

func main() {
	if database.ReadOnlyDbConnectionErr != nil || database.ReadWriteDbConnectionErr != nil {
		log.Fatal("Unable to connect to database")
	}
	defer database.ReadOnlyDb.Close()
	defer database.ReadWriteDb.Close()

	database.ReadOnlyDb.SetMaxOpenConns(8)
	database.ReadWriteDb.SetMaxOpenConns(1)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// pages
	http.HandleFunc("/", handler.HomeHandler)
	http.HandleFunc("/about", handler.AboutHandler)
	http.HandleFunc("/contact", handler.ContactHandler)

	// tests endpoints
	http.HandleFunc("/tests", handler.TestsHandler)
	http.HandleFunc("/tests/", handler.TestsHandler)

	http.HandleFunc("/question", handler.QuestionHandler)

	fmt.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", nil)
}
