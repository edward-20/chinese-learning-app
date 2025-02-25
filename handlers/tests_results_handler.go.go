package handler

import (
	"fmt"
	"net/http"

	"github.com/edward-20/chinese-learning-app/database"
	"github.com/edward-20/chinese-learning-app/templates"
	"github.com/edward-20/chinese-learning-app/utils"
)

func TestsResultsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("/test-results")
	sessionCookie, getCookieError := r.Cookie("session_id")
	if getCookieError != nil {
		http.Error(w, "Invalid Request to /test-results, provide sessionID cookie", http.StatusBadRequest)
		return
	}
	sessionID := sessionCookie.Value
	if !utils.IsUserRegisteredInDatabase(sessionID) {
		http.Error(w, "Internal Server Error. User is not registered in the database.", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// get the users responses, their total number of question they wanted in their test and their score

		// total number of questions
		var totalNumberOfQuestions int
		err := database.ReadOnlyDb.QueryRow("SELECT totalNumberOfQuestions FROM Tests WHERE userSessionID = ?", sessionID).Scan(&totalNumberOfQuestions)
		if err != nil {
			http.Error(w, "Could not read Tests table", http.StatusInternalServerError)
			return
		}
		// user's responses and score
		type response struct {
			QuestionNumber    int
			ChineseCharacters string
			UsersAnswer       string
			CorrectAnswer     string
		}
		responses := make([]response, totalNumberOfQuestions)
		score := 0
		rows, err := database.ReadOnlyDb.Query("SELECT q.questionNumber, w.chineseCharacters, q.usersAnswer, w.pinyin FROM Questions as q JOIN Words as w ON q.wordId = w.id WHERE q.testID = ?", sessionID)
		for i := 0; rows.Next(); i++ {
			var questionNumber int
			var chineseCharacters, usersAnswer, pinyin string
			err = rows.Scan(&questionNumber, &chineseCharacters, &usersAnswer, &pinyin)
			if err != nil {
				http.Error(w, "Could not read Questions and Words table", http.StatusInternalServerError)
			}

			if usersAnswer == pinyin {
				score++
			}
			responses[i] = response{
				QuestionNumber:    questionNumber,
				ChineseCharacters: chineseCharacters,
				UsersAnswer:       usersAnswer,
				CorrectAnswer:     pinyin,
			}
		}
		rows.Close()
		context := struct {
			TestID                 string
			Score                  int
			TotalNumberOfQuestions int
			Responses              []response
		}{
			TestID:                 sessionID,
			Score:                  score,
			TotalNumberOfQuestions: totalNumberOfQuestions,
			Responses:              responses,
		}
		fmt.Println("here")
		w.WriteHeader(http.StatusOK)
		utils.RenderTemplate(w, templates.TestReviewTemplate, context)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
}
