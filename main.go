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
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var aboutTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/about.html"))
var contactTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/contact.html"))
var startTestTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/test-start.html"))
var resumeTestTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/test-resume.html"))
var testQuestionTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/single-character-question.html"))
var testSolutionTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/check-answer.html"))
var testReviewTemplate = template.Must(template.ParseFiles("templates/base.html", "templates/review.html"))

var readWriteDB, readWriteDBConnectionErr = sql.Open("sqlite3", "./db/chinese-learning-database.db?_journal=WAL&busy_timeout=5000&_foreign_keys=on")
var readOnlyDB, readOnlyDBConnectionErr = sql.Open("sqlite3", "./db/chinese-learning-database.db?_journal=WAL&busy_timeout=5000&mode=ro&_foreign_keys=on")

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
	_, err := readWriteDB.Exec("INSERT INTO Users (sessionID) VALUES (?)", sessionId)
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
	var result bool
	readOnlyDB.QueryRow("SELECT EXISTS (SELECT 1 FROM Users WHERE sessionID = ?)", sessionID).Scan(&result)
	return result
}

func doesUserHaveTest(sessionID string) bool {
	// determine if they have a test
	var result bool
	readOnlyDB.QueryRow("SELECT EXISTS (SELECT 1 FROM Tests WHERE userSessionID = ?)", sessionID).Scan(&result)
	return result
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
		_, err := readWriteDB.Exec("INSERT INTO Users (sessionID) VALUES (?)", sessionID)
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
			http.Error(w, "Invalid Request to POST /tests, provide number of questions as integer query", http.StatusBadRequest)
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
		_, err = tx.Exec("INSERT INTO Tests (userSessionId, totalNumberOfQuestions) VALUES (?, ?)", sessionID, numQuestions)
		if err != nil {
			http.Error(w, "Could not execute INSERT to Tests in transaction", http.StatusInternalServerError)
			tx.Rollback()
			return
		}

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

		var chineseCharacter sql.NullString
		err = readOnlyDB.QueryRow("SELECT chineseCharacters FROM Words WHERE id = (SELECT wordID FROM Questions WHERE testID = ? AND questionNumber = ?)", sessionID, 1).Scan(&chineseCharacter)
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
			testID           string
		}{chineseCharacter: chineseCharacter.String, questionNumber: 1, testID: sessionID}
		renderTemplate(w, testQuestionTemplate, context)

	case http.MethodGet:
		path := strings.TrimPrefix(r.URL.Path, "/tests")
		if path == "" {
			// get the testID from the user
			http.Error(w, "GET /tests has not been implemented", http.StatusNotFound)
			return
		}

		// check that this test belongs to this user
		if path != sessionID {
			http.Error(w, "The test doesn't belong to this user", http.StatusForbidden)
			return
		}

		// check that the user actually has a test
		var currentQuestion int
		var totalNumberOfQuestions int
		err := readOnlyDB.QueryRow("SELECT currentQuestion, totalNumberOfQuestions FROM Tests WHERE userSessionID = ?)", path).Scan(&currentQuestion)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "The user doesn't have a test", http.StatusNotFound)
				return
			}
			http.Error(w, "The user doesn't have a test in an unforseen way", http.StatusNotFound)
			return
		}

		// get their question
		var wordID int
		var usersAnswer sql.NullString
		readOnlyDB.QueryRow("SELECT wordID, usersAnswer FROM Questions WHERE testID = ? AND questionNumber = ?", path, currentQuestion).Scan(&wordID, &usersAnswer)

		// if they've already answered this question
		if usersAnswer.Valid {
			// end the test
			if currentQuestion == totalNumberOfQuestions {
				// compute the number of correct answers
				var score int
				err := readOnlyDB.QueryRow("SELECT COUNT(*) FROM Questions q JOIN Words w ON q.wordID = w.id WHERE q.testID = ? AND q.usersAnswer = w.pinyin", path).Scan(&score)
				if err != nil {
					http.Error(w, "Score could not be obtained", http.StatusInternalServerError)
				}
				renderTemplate(w, testReviewTemplate, struct {
					score                  int
					totalNumberOfQuestions int
					testID                 string
				}{score: score, totalNumberOfQuestions: totalNumberOfQuestions, testID: sessionID})
				return
			}
			currentQuestion += 1
			// or move onto the next question (not sure why this happens)
			_, err = readWriteDB.Exec("UPDATE Tests SET currentQuestion = ? WHERE userSessionID = ?", currentQuestion)
			if err != nil {
				http.Error(w, "Unable to update the current question", http.StatusInternalServerError)
			}
			var chineseCharacter string
			err = readOnlyDB.QueryRow("SELECT chineseCharacters FROM Words WHERE id = (SELECT wordID FROM Questions WHERE testID = ? AND questionNumber = ?)", sessionID, currentQuestion).Scan(&chineseCharacter)
			renderTemplate(w, testQuestionTemplate, struct {
				chineseCharacter string
				questionNumber   int
				testID           string
			}{chineseCharacter: chineseCharacter, questionNumber: currentQuestion, testID: sessionID})
			return
		}
		// else they haven't
		var chineseCharacter string
		readOnlyDB.QueryRow("SELECT chineseCharacters FROM Word WHERE id = ?", wordID).Scan(&chineseCharacter)

		context := struct {
			chineseCharacter string
			questionNumber   int
			testID           string
		}{chineseCharacter: chineseCharacter, questionNumber: currentQuestion, testID: sessionID}
		renderTemplate(w, testQuestionTemplate, context)
	case http.MethodDelete:
		path := strings.TrimPrefix(r.URL.Path, "/tests")
		if path == "" {
			// get the testID from the user
			http.Error(w, "DELETE /tests has not been implemented", http.StatusNotFound)
			return
		}

		// check that this test belongs to this user
		if path != sessionID {
			http.Error(w, "The test doesn't belong to this user", http.StatusForbidden)
			return
		}

		// delete their test and their questions and then give them the test-start page
		_, err := readWriteDB.Exec("DELETE FROM Tests WHERE userSessionID = ?", sessionID)
		if err != nil {
			http.Error(w, "Unable to delete test", http.StatusInternalServerError)
		}
		renderTemplate(w, startTestTemplate, nil)
	default:
		http.Error(w, "/tests does not have implementation for methods outside of GET DELETE and POST", http.StatusNotFound)
	}
	return
}

func questionHandler(w http.ResponseWriter, r *http.Request) {
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

	testID := r.URL.Query().Get("testID")
	questionNumber := r.URL.Query().Get("questionNumber")
	if sessionID != testID {
		http.Error(w, "Test doesn't belong to user", http.StatusForbidden)
		return
	}

	if questionNumber == "" {
		http.Error(w, "Malformed Request to /question. Needs questionNumber query to endpoint", http.StatusBadRequest)
		return
	}

	currentQuestion, err := strconv.Atoi(questionNumber)
	if err != nil {
		http.Error(w, "Malformed Request to /question. questionNumber query needs to be an integer", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// get the chinese character
		var chineseCharacter string
		readOnlyDB.QueryRow("SELECT chineseCharacters FROM Words WHERE id = (SELECT wordID FROM Questions WHERE testID = ? AND questionNumber = ?)", testID, currentQuestion).Scan(&chineseCharacter)
		context := struct {
			chineseCharacter string
			questionNumber   int
			testID           string
		}{chineseCharacter: chineseCharacter, questionNumber: currentQuestion}
		renderTemplate(w, testQuestionTemplate, context)
	case http.MethodPatch:
		userAnswer := r.URL.Query().Get("userAnswer")
		if userAnswer == "" {
			http.Error(w, "Malformed Request to PATCH /question. userAnswer needs to be supplied.", http.StatusBadRequest)
			return
		}
		// update the question with the users input
		readWriteDB.Exec("UPDATE Questions SET usersAnswer = ? WHERE testID = ? AND questionNumber = ?", testID, currentQuestion)
		// update the current question in tests
		readWriteDB.Exec("UPDATE Tests SET currentQuestion = currentQuestion + 1 WHERE id = ?", testID)
		// render a template telling them if they're correct or not
		var chineseCharacter, correctPinyinAnswer string
		readOnlyDB.QueryRow("SELECT chineseCharacters, pinyin FROM Words WHERE id = (SELECT wordID from Questions WHERE testID = ? AND questionNumber = ?)", testID, currentQuestion+1).Scan(&chineseCharacter, &correctPinyinAnswer)
		context := struct {
			chineseCharacter    string
			correctPinyinAnswer string
			userPinyinAnswer    string
			testID              string
			nextQuestionNumber  int
		}{chineseCharacter: chineseCharacter, correctPinyinAnswer: correctPinyinAnswer, userPinyinAnswer: userAnswer, testID: testID, nextQuestionNumber: currentQuestion + 1}
		renderTemplate(w, testSolutionTemplate, context)
		return
	}
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

	http.HandleFunc("/question", questionHandler)

	fmt.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", nil)
}
