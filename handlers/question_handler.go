package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/edward-20/chinese-learning-app/database"
	"github.com/edward-20/chinese-learning-app/templates"
	"github.com/edward-20/chinese-learning-app/utils"
)

func QuestionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("/questions")
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

	testID := r.FormValue("test-id")
	questionNumber := r.FormValue("question-number")
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
		database.ReadOnlyDb.QueryRow("SELECT chineseCharacters FROM Words WHERE id = (SELECT wordID FROM Questions WHERE testID = ? AND questionNumber = ?)", testID, currentQuestion).Scan(&chineseCharacter)
		context := struct {
			ChineseCharacter string
			QuestionNumber   int
			TestID           string
		}{ChineseCharacter: chineseCharacter, QuestionNumber: currentQuestion, TestID: testID}
		utils.RenderTemplate(w, templates.TestQuestionTemplate, context)
	case http.MethodPatch:
		userAnswer := r.FormValue("user-answer")
		if userAnswer == "" {
			http.Error(w, "Malformed Request to PATCH /question. userAnswer needs to be supplied.", http.StatusBadRequest)
			return
		}
		// update the question with the users input
		_, err := database.ReadWriteDb.Exec("UPDATE Questions SET usersAnswer = ? WHERE testID = ? AND questionNumber = ?", userAnswer, testID, currentQuestion)
		if err != nil {
			// error handling function
			http.Error(w, "Could not update Questions table", http.StatusInternalServerError)
		}
		// update the current question in tests
		_, err = database.ReadWriteDb.Exec("UPDATE Tests SET currentQuestion = ? WHERE userSessionID = ?", currentQuestion+1, testID)
		if err != nil {
			// error handling function
			http.Error(w, "Could not update Tests table", http.StatusInternalServerError)
		}
		// render a template telling them if they're correct or not
		var chineseCharacter, correctPinyinAnswer string
		database.ReadOnlyDb.QueryRow("SELECT chineseCharacters, pinyin FROM Words WHERE id = (SELECT wordID from Questions WHERE testID = ? AND questionNumber = ?)", testID, currentQuestion).Scan(&chineseCharacter, &correctPinyinAnswer)
		context := struct {
			ChineseCharacter    string
			CorrectPinyinAnswer string
			UserPinyinAnswer    string
			TestID              string
			NextQuestionNumber  int
		}{ChineseCharacter: chineseCharacter, CorrectPinyinAnswer: correctPinyinAnswer, UserPinyinAnswer: userAnswer, TestID: testID, NextQuestionNumber: currentQuestion + 1}
		utils.RenderTemplate(w, templates.TestSolutionTemplate, context)
		return
	}
}
