package utils

import (
	"github.com/edward-20/chinese-learning-app/database"
)

func IsUserRegisteredInDatabase(sessionID string) bool {
	var result bool
	database.ReadOnlyDb.QueryRow("SELECT EXISTS (SELECT 1 FROM Users WHERE sessionID = ?)", sessionID).Scan(&result)
	return result
}

func DoesUserHaveTest(sessionID string) bool {
	// determine if they have a test
	var result bool
	database.ReadOnlyDb.QueryRow("SELECT EXISTS (SELECT 1 FROM Tests WHERE userSessionID = ?)", sessionID).Scan(&result)
	return result
}
