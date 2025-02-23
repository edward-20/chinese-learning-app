package utils

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/edward-20/chinese-learning-app/database"
)

func GenerateSessionID() (string, error) {
	b := make([]byte, 16) // 16 bytes = 128 bits
	_, err := cryptorand.Read(b)
	if err != nil {
		return "", errors.New("Error generating session ID:")
	}
	return hex.EncodeToString(b), nil
}

func AddUserSession(sessionId string) error {
	_, err := database.ReadWriteDb.Exec("INSERT INTO Users (sessionID) VALUES (?)", sessionId)
	return err
}

func SetSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,                                             // To prevent access from JavaScript
		Secure:   false,                                            // Should be true if using HTTPS
		Expires:  time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC), // Far future date
		SameSite: http.SameSiteStrictMode,
	})
}
