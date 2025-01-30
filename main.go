package main

import (
	cryptorand "crypto"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"sync"
	"time"
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

type User struct {
	sessionId string
	score     int
	outOf     int
}

// map sessionIDs to session relevant data
var userSessions = make(map[string]any)

var userMutex sync.RWMutex

var templates = template.Must(template.ParseGlob("templates/*.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	w.Header().Set("Content-Type", "text/html")
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

/* Session Management */
func generateSessionID() (string, error) {
	b := make([]byte, 16) // 16 bytes = 128 bits
	_, err := cryptorand.Read(b)
	if err != nil {
		return "", errors.New("Error generating session ID:")
	}
	return hex.EncodeToString(b), nil
}

func addUserSession(sessionId string) {
	defer userMutex.Unlock()

	userMutex.Lock()
	userSessions[sessionId] = User{
		sessionId: sessionId,
	}
}

func setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,                                             // To prevent access from JavaScript
		Secure:   false,                                            // Should be true if using HTTPS
		Expires:  time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC), // Far future date
		SameSite: http.SameSiteStrictMode,
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("session_id")
	if err != nil {
		sessionID, err := generateSessionID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		addUserSession(sessionID)
		setSessionCookie(w, sessionID)
	}
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
