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

var aboutTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/about.html"))
var contactTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/contact.html"))
var startTestTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/test-start.html"))
var resumeTestTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/test-resume.html"))
var db, dbConnectionErr = sql.Open("sqlite3", "./db/chinese-learning-database.db")
var dbMutex sync.RWMutex

func renderTemplate(w http.ResponseWriter, temp *template.Template, data any) {
	w.Header().Set("Content-Type", "text/html")
	err := temp.ExecuteTemplate(w, "base.html", data)
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

func isUserRegisteredInDatabase(sessionID string) bool {
	var isUserRegistered sql.NullBool
	db.QueryRow("SELECT EXISTS (SELECT 1 FROM Users WHERE sessionID = \"?\")", sessionID).Scan(&isUserRegistered)
	return isUserRegistered.Valid
}

func doesUserHaveTest(sessionID string) bool{
	// determine if they have a test
	var currentQuestion sql.NullInt16
	noTestError := db.QueryRow("SELECT currentQuestion FROM Tests WHERE userSessionID = \"?\"", sessionID).Scan(&currentQuestion)
	if noTestError != nil {
		return false
	}
	return true
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	sessionCookie, getCookieError := r.Cookie("session_id")

	// the user has not visited the site before
	if getCookieError != nil {
		sessionID, randomGenerationError := generateSessionID()
		if randomGenerationError != nil {
			http.Error(w, randomGenerationError.Error(), http.StatusInternalServerError)
			return
		}
		dbError := addUserSession(sessionID)
		if dbError != nil {
			http.Error(w, dbError.Error(), http.StatusInternalServerError)
			return
		}
		setSessionCookie(w, sessionID)
		// new user session
		renderTemplate(w, startTestTemplate, nil)
		return
	}

	// the user has visited the site before
	sessionID := sessionCookie.Value
	if !isUserRegisteredInDatabase(sessionID) {
		_, err := db.Exec("INSERT INTO Users (sessionID) VALUES (?)", sessionID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

	// determine if they have a test
	if !doesUserHaveTest(sessionID) {
		renderTemplate(w, startTestTemplate, nil)
		return
	}
	renderTemplate(w, resumeTestTemplate, currentQuestion.Value)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, aboutTemplate, nil)
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, contactTemplate, nil)
}

func testsHandler(w http.ResponseWriter, r *http.Request) {
	// is there a session cookie
	sessionCookie, getCookieError := r.Cookie("session_id")
	if getCookieError != nil {
		http.Error(w, "Invalid Request to /tests, provide sessionID cookie", http.StatusBadRequest)
		return
	}
	sessionID := sessionCookie.Value

	if sessionID
	switch r.Method {
	case http.MethodPost:
		numQuestionsWanted := r.URL.Query().Get("number-of-questions")
		if numQuestionsWanted == "" {
			http.Error(w, "Invalid Request to POST /tests, provide number of questions as query", http.StatusBadRequest)
			return
		}

		/*
			Summary: create a test and give back html
			Preconditions:
				* test must not exist (405)
		*/
		db.QueryRow("SELECT * FROM ")
	case http.MethodGet:
	case http.MethodDelete:
	}
}
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

	// tests endpoints
	http.HandleFunc("/tests")

	fmt.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", nil)
}
