package main

import (
	cryptorand "crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))
var db, dbConnectionErr = sql.Open("sqlite3", "./db/chinese-learning-database.db")
var dbMutex sync.RWMutex

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

func addUserSession(sessionId string) error {
	dbMutex.Lock()
	_, err := db.Exec("INSERT INTO Users (sessionID) VALUES (?)", sessionId)
	dbMutex.Unlock()
	return err
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
	_, getCookieError := r.Cookie("session_id")
	if getCookieError != nil {
		sessionID, randomGenerationError := generateSessionID()
		if randomGenerationError != nil {
			http.Error(w, randomGenerationError.Error(), http.StatusInternalServerError)
		}
		err := addUserSession(sessionID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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

/*
	func chineseCharactersHandler(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			numberOfQuestions, err := strconv.Atoi(r.URL.Query().Get("number-of-questions"))

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else if numberOfQuestions <= 0 {
				http.Error(w, "The number of questions must be greater than 0.", http.StatusBadRequest)
			} else {
				currentUserSession := r.Header.Get("session_id")
				userMutex.Lock()
				currentUser := userSessions[currentUserSession]
				currentUser.outOf = numberOfQuestions
				userMutex.Unlock()

				v := rand.Intn(len(dictionary))
				randomWord := dictionary[v]

				renderTemplate(w, "single-character-question.html", randomWord)
			}
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
*/
func main() {
	if dbConnectionErr != nil {
		log.Fatal(dbConnectionErr)
	}
	defer db.Close()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// pages
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/contact", contactHandler)

	fmt.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", nil)
}
