package main

import (
	"fmt"
	"html/template"
	"net/http"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	w.Header().Set("Content-Type", "text/html")
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "Home",
		"Name":  "Guest",
	}
	renderTemplate(w, "base.html", data)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "About Us",
	}
	renderTemplate(w, "base.html", data)
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/about", aboutHandler)

	fmt.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", nil)
}
