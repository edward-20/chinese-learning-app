package main

import (
	cryptorand "crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var aboutTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/about.html"))
var contactTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/contact.html"))
var startTestTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/test-start.html"))
var resumeTestTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/test-resume.html"))
var testQuestionTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/single-character-question.html"))
var testSolutionRemplate = template.Must(template.ParseFiles("templates/base.html", "templates/check-answer.html"))

var readWriteDB, readWriteDBConnectionErr = sql.Open("sqlite3", "./db/chinese-learning-database.db?_journal=WAL&busy_timeout=5000")
var readOnlyDB, readOnlyDBConnectionErr = sql.Open("sqlite3", "./db/chinese-learning-database.db?_journal=WAL&busy_timeout=5000&mode=ro")

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
	_, err := readWriteDB.Exec("INSERT INTO Users (sessionID) VALUES ('?')", sessionId)
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
	readOnlyDB.QueryRow("SELECT EXISTS (SELECT 1 FROM Users WHERE sessionID = \"?\")", sessionID).Scan(&isUserRegistered)
	return isUserRegistered.Valid
}

func doesUserHaveTest(sessionID string) bool {
	// determine if they have a test
	var currentQuestion sql.NullInt64
	noTestError := readOnlyDB.QueryRow("SELECT currentQuestion FROM Tests WHERE userSessionID = \"?\"", sessionID).Scan(&currentQuestion)
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
		_, err := readWriteDB.Exec("INSERT INTO Users (sessionID) VALUES ('?')", sessionID)
		if err != nil {
			http.Error(w, "Could not create user in database", 500)
			return
		}
	}

	// determine if they have a test
	if !doesUserHaveTest(sessionID) {
		renderTemplate(w, startTestTemplate, nil)
		return
	}
	renderTemplate(w, resumeTestTemplate, nil)
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

	if !isUserRegisteredInDatabase(sessionID) {
		http.Error(w, "Internal Server Error. User is not registered in the database.", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodPost:
		if doesUserHaveTest(sessionID) {
			http.Error(w, "Method Not Allowed. User already has a test.", http.StatusMethodNotAllowed)
			return
		}

		numQuestionsWanted := r.URL.Query().Get("number-of-questions")
		if numQuestionsWanted == "" {
			http.Error(w, "Invalid Request to POST /tests, provide number of questions as query", http.StatusBadRequest)
			return
		}

		numQuestions, err := strconv.Atoi(numQuestionsWanted)
		if err != nil {
			http.Error(w, "Invalid Request to POST /tests, provide number of questions as query", http.StatusBadRequest)
			return
		}

		if numQuestions > 0 || numQuestions > 500 {
			http.Error(w, "Invalid Request to POST /tests, provide number of questions in the query within the range of 1-500.", http.StatusBadRequest)
			return
		}

		// create a test
		tx, err := readWriteDB.Begin()
		if err != nil {
			http.Error(w, "Could not begin transaction", http.StatusInternalServerError)
			return
		}
		newTest, err := tx.Exec("INSERT INTO Tests (userSessionId, totalNumberOfQuestions) VALUES ('?', ?)", sessionID, numQuestions)
		if err != nil {
			http.Error(w, "Could not execute INSERT to Tests in transaction", http.StatusInternalServerError)
			tx.Rollback()
			return
		}

		newTestID, err := newTest.LastInsertId()

		// create the questions
		permutation := rand.Perm(500)
		for questionNumber, randomNumber := range permutation {
			_, err = tx.Exec("INSERT INTO Questions (wordID, testID, questionNumber) VALUES (?, ?, ?)", randomNumber+1, sessionID, questionNumber+1)
			if err != nil {
				http.Error(w, "Could not execute INSERT to Questions in transaction", http.StatusInternalServerError)
				tx.Rollback()
				return
			}
		}
		err = tx.Commit()
		if err != nil {
			http.Error(w, "Could not commit transaction.", http.StatusInternalServerError)
		}

		var wordID, testID int
		err = readOnlyDB.QueryRow("SELECT wordID, testID FROM Questions WHERE testID = ? AND questionNumber = ?", newTestID, 1).Scan(&wordID, &testID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Could not find the first question", 500)
				return
			}
			http.Error(w, "Could not find the first question due to unforseen error", 500)
			return
		}

		var chineseCharacter sql.NullString
		err = readOnlyDB.QueryRow("SELECT chineseCharacters FROM Words WHERE id = ", wordID).Scan(&chineseCharacter)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Could not find the word corresponding to the question", 500)
				return
			}
			http.Error(w, "Could not find the word corresponding to the question due to unforseen error", 500)
			return
		}
		if !chineseCharacter.Valid {
			http.Error(w, "Could not find details of the chinese character of the first question due to unforseen error", 500)
			return
		}

		context := struct {
			chineseCharacter string
			questionNumber   int
			testID           int
		}{chineseCharacter: chineseCharacter.String, questionNumber: 1, testID: testID}
		renderTemplate(w, testQuestionTemplate, context)
	case http.MethodGet:
		http.Error(w, "Endpoint has not been implemented", http.StatusNotFound)
	case http.MethodDelete:
		http.Error(w, "Endpoint has not been implemented", http.StatusNotFound)
	}
	return
}

func main() {
	if readOnlyDBConnectionErr != nil || readWriteDBConnectionErr != nil {
		log.Fatal("Unable to connect to database")
	}
	defer readOnlyDB.Close()
	defer readWriteDB.Close()

	readOnlyDB.SetMaxOpenConns(8)
	readWriteDB.SetMaxOpenConns(1)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// pages
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/contact", contactHandler)

	// tests endpoints
	http.HandleFunc("/tests", testsHandler)

	fmt.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", nil)
}
