package handler

import (
	"fmt"
	"net/http"

	"github.com/edward-20/chinese-learning-app/database"
	"github.com/edward-20/chinese-learning-app/templates"
	"github.com/edward-20/chinese-learning-app/utils"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("/home")
	sessionCookie, getCookieError := r.Cookie("session_id")

	// the user has not visited the site before
	if getCookieError != nil {
		fmt.Println("here1")
		sessionID, randomGenerationError := utils.GenerateSessionID()
		if randomGenerationError != nil {
			http.Error(w, randomGenerationError.Error(), http.StatusInternalServerError)
			return
		}
		dbError := utils.AddUserSession(sessionID)
		if dbError != nil {
			fmt.Println("here1.5")
			http.Error(w, dbError.Error(), http.StatusInternalServerError)
			return
		}
		utils.SetSessionCookie(w, sessionID)
		// new user session
		utils.RenderTemplate(w, templates.StartTestTemplate, nil)
		return
	}

	fmt.Println("here2")
	// the user has visited the site before
	sessionID := sessionCookie.Value
	if !utils.IsUserRegisteredInDatabase(sessionID) {
		_, err := database.ReadWriteDb.Exec("INSERT INTO Users (sessionID) VALUES (?)", sessionID)
		if err != nil {
			http.Error(w, "Could not create user in database"+err.Error(), 500)
			return
		}
	}

	// determine if they have a test
	if !utils.DoesUserHaveTest(sessionID) {
		utils.RenderTemplate(w, templates.StartTestTemplate, nil)
		return
	}
	// find their testID
	var testID string
	err := database.ReadOnlyDb.QueryRow("SELECT userSessionID FROM Tests WHERE userSessionID = ?", sessionID).Scan(&testID)
	if err != nil {
		http.Error(w, "Could not find testID of user", http.StatusInternalServerError)
	}
	utils.RenderTemplate(w, templates.ResumeTestTemplate, struct{ TestID string }{TestID: testID})
}
