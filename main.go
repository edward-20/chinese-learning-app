package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type Word struct {
	ChineseCharacter string
	Pinyin           string
}

var dictionary = [500]Word{
	{ChineseCharacter: "爱 ", Pinyin: "ai4"},
	{ChineseCharacter: "爱好", Pinyin: "ai4 hao4"},
	{ChineseCharacter: "吧", Pinyin: "ba"},
	{ChineseCharacter: "爸爸", Pinyin: "bai2"},
	{ChineseCharacter: "白", Pinyin: "bai3"},
	{ChineseCharacter: "白天", Pinyin: "bai2 tian1"},
	{ChineseCharacter: "班", Pinyin: "ban1"},
	{ChineseCharacter: "半", Pinyin: "ban4"},
	{ChineseCharacter: "半年", Pinyin: "ban4 nian2"},
	{ChineseCharacter: "半天", Pinyin: "ban4 tian1"},
	{ChineseCharacter: "帮", Pinyin: "bang1"},
}

var templates = template.Must(template.ParseGlob("templates/*.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	w.Header().Set("Content-Type", "text/html")
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "base.html", nil)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "base.html", "about")
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "base.html", "contact")
}

func main() {
	for _, tmpl := range templates.Templates() {
		fmt.Println(tmpl.Name())
	}
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/contact", contactHandler)

	fmt.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", nil)
}
