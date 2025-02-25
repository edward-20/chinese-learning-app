package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/edward-20/chinese-learning-app/database"
	"github.com/edward-20/chinese-learning-app/templates"
	"github.com/edward-20/chinese-learning-app/utils"
)

func TestsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("/tests")
	// is there a session cookie
	sessionCookie, getCookieError := r.Cookie("session_id")
	if getCookieError != nil {
		http.Error(w, "Invalid Request to /tests, provide sessionID cookie", http.StatusBadRequest)
		return
	}
	sessionID := sessionCookie.Value

	if !utils.IsUserRegisteredInDatabase(sessionID) {
		http.Error(w, "Internal Server Error. User is not registered in the database.", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodPost:
		fmt.Println("POST /tests")
		if utils.DoesUserHaveTest(sessionID) {
			http.Error(w, "Method Not Allowed. User already has a test.", http.StatusMethodNotAllowed)
			return
		}

		numQuestionsWanted := r.FormValue("number-of-questions")
		if numQuestionsWanted == "" {
			http.Error(w, "Invalid Request to POST /tests, provide number of questions as query", http.StatusBadRequest)
			return
		}

		numQuestions, err := strconv.Atoi(numQuestionsWanted)
		if err != nil {
			http.Error(w, "Invalid Request to POST /tests, provide number of questions as integer query", http.StatusBadRequest)
			return
		}

		if numQuestions < 0 || numQuestions > 500 {
			http.Error(w, "Invalid Request to POST /tests, provide number of questions in the query within the range of 1-500.", http.StatusBadRequest)
			return
		}

		// create a test
		_, err = database.ReadWriteDb.Exec("INSERT INTO Tests (userSessionId, totalNumberOfQuestions) VALUES (?, ?)", sessionID, numQuestions)
		if err != nil {
			http.Error(w, "Could not execute INSERT to Tests in transaction", http.StatusInternalServerError)
			return
		}

		// create the questions
		tx, err := database.ReadWriteDb.Begin()
		permutation := rand.Perm(500)
		for questionNumber, randomNumber := range permutation[:numQuestions] {
			_, err = tx.Exec("INSERT INTO Questions (wordID, testID, questionNumber) VALUES (?, ?, ?)", randomNumber+1, sessionID, questionNumber+1)
			if err != nil {
				fmt.Println(questionNumber)
				fmt.Println(err.Error())
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
		err = database.ReadOnlyDb.QueryRow("SELECT chineseCharacters FROM Words WHERE id = (SELECT wordID FROM Questions WHERE testID = ? AND questionNumber = ?)", sessionID, 1).Scan(&chineseCharacter)
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
			ChineseCharacter string
			QuestionNumber   int
			TestID           string
		}{ChineseCharacter: chineseCharacter.String, QuestionNumber: 1, TestID: sessionID}
		w.WriteHeader(http.StatusCreated)
		utils.RenderTemplate(w, templates.TestQuestionTemplate, context)
	case http.MethodGet:
		path := strings.TrimPrefix(r.URL.Path, "/tests/")
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
		err := database.ReadOnlyDb.QueryRow("SELECT currentQuestion, totalNumberOfQuestions FROM Tests WHERE userSessionID = ?", path).Scan(&currentQuestion, &totalNumberOfQuestions)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "The user doesn't have a test", http.StatusNotFound)
				return
			}
			http.Error(w, "The user doesn't have a test in an unforseen way. "+err.Error(), http.StatusNotFound)
			return
		}

		// get their question
		var wordID int
		var usersAnswer sql.NullString
		database.ReadOnlyDb.QueryRow("SELECT wordID, usersAnswer FROM Questions WHERE testID = ? AND questionNumber = ?", path, currentQuestion).Scan(&wordID, &usersAnswer)

		// if they've already answered this question
		if usersAnswer.Valid {
			// end the test
			if currentQuestion == totalNumberOfQuestions {
				// compute the number of correct answers
				var score int
				err := database.ReadOnlyDb.QueryRow("SELECT COUNT(*) FROM Questions q JOIN Words w ON q.wordID = w.id WHERE q.testID = ? AND q.usersAnswer = w.pinyin", path).Scan(&score)
				if err != nil {
					http.Error(w, "Score could not be obtained", http.StatusInternalServerError)
				}
				fmt.Println("here1")
				w.WriteHeader(http.StatusOK)
				utils.RenderTemplate(w, templates.TestReviewTemplate, struct {
					Score                  int
					TotalNumberOfQuestions int
					TestID                 string
				}{Score: score, TotalNumberOfQuestions: totalNumberOfQuestions, TestID: sessionID})
				return
			}
			currentQuestion += 1
			// or move onto the next question (not sure why this happens)
			_, err = database.ReadWriteDb.Exec("UPDATE Tests SET currentQuestion = ? WHERE userSessionID = ?", currentQuestion)
			if err != nil {
				http.Error(w, "Unable to update the current question", http.StatusInternalServerError)
			}
			var chineseCharacter string
			err = database.ReadOnlyDb.QueryRow("SELECT chineseCharacters FROM Words WHERE id = (SELECT wordID FROM Questions WHERE testID = ? AND questionNumber = ?)", sessionID, currentQuestion).Scan(&chineseCharacter)
			fmt.Println("here2")
			w.WriteHeader(http.StatusOK)
			utils.RenderTemplate(w, templates.TestQuestionTemplate, struct {
				ChineseCharacter string
				QuestionNumber   int
				TestID           string
			}{ChineseCharacter: chineseCharacter, QuestionNumber: currentQuestion, TestID: sessionID})
			return
		}
		// else they haven't
		var chineseCharacter string
		database.ReadOnlyDb.QueryRow("SELECT chineseCharacters FROM Words WHERE id = ?", wordID).Scan(&chineseCharacter)

		fmt.Println(chineseCharacter, currentQuestion, sessionID)
		context := struct {
			ChineseCharacter string
			QuestionNumber   int
			TestID           string
		}{ChineseCharacter: chineseCharacter, QuestionNumber: currentQuestion, TestID: sessionID}
		fmt.Println("here3")
		w.WriteHeader(http.StatusOK)
		utils.RenderTemplate(w, templates.TestQuestionTemplate, context)
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
		_, err := database.ReadWriteDb.Exec("DELETE FROM Tests WHERE userSessionID = ?", sessionID)
		if err != nil {
			http.Error(w, "Unable to delete test", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
		utils.RenderTemplate(w, templates.StartTestTemplate, nil)
	default:
		http.Error(w, "/tests does not have implementation for methods outside of GET DELETE and POST", http.StatusNotFound)
	}
	return
}
