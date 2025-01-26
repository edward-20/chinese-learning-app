package main

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
)

type Word struct {
	ChineseCharacter string `json:"chineseCharacter"`
	Pinyin           string `json:"pinyin"`
}

type QuestionAndAnswer struct {
	ChineseCharacter    string `json:"chineseCharacter`
	CorrectPinyinAnswer string `json:"correctPinyinAnswer`
	UserPinyinAnswer    string `json:"userPinyinAnswer`
}

var dictionary = []Word{
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

type Method int

const (
	GET   Method = 0
	PUT   Method = 1
	POST  Method = 2
	PATCH Method = 3
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
	renderTemplate(w, "base.html", nil)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "base.html", "about")
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "base.html", "contact")
}

func chineseCharactersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// get a random word from the dictionary
		v := rand.Intn(len(dictionary))
		randomWord := dictionary[v]
		renderTemplate(w, "single-character-question.html", randomWord)
	}
}

func checkAnswerHandler(w http.ResponseWriter, r *http.Request) {
	// check the body of the request to see if the pinyin matches the character
	valuesSent := r.URL.Query()
	userAnswer := valuesSent.Get("user-answer")
	correctAnswer := valuesSent.Get("correct-answer")
	chineseCharacter := valuesSent.Get("chinese-character")

	renderTemplate(w, "check-answer.html", QuestionAndAnswer{ChineseCharacter: chineseCharacter, CorrectPinyinAnswer: correctAnswer, UserPinyinAnswer: userAnswer})
}

func main() {
	for _, tmpl := range templates.Templates() {
		fmt.Println(tmpl.Name())
	}
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// pages
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/contact", contactHandler)

	// partial pages
	http.HandleFunc("/api/chinese-character", chineseCharactersHandler)
	http.HandleFunc("/api/check-answer", checkAnswerHandler)

	fmt.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", nil)
}
