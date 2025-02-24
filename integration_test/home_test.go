package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	handler "github.com/edward-20/chinese-learning-app/handlers"
)

const (
	TEST_URL = "http://localhost:8080"
)

func TestStartTest(t *testing.T) {
	// make a first request to the home handler
	request := httptest.NewRequest(http.MethodGet, TEST_URL, nil)
	writer := httptest.NewRecorder()

	handler.HomeHandler(writer, request)

	response := writer.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("Didn't return correct status code: %d", writer.Code)
	}

	cookies := response.Cookies()
	if len(cookies) > 1 {
		t.Fatal("There should be only one cookie, instead there are more")
	} else if len(cookies) < 1 {
		t.Fatal("There should be only one cookie, instead there is none")
	}

	sessionIDCookie := cookies[0]
	if sessionIDCookie.Name == "session_id" {
		// needs to have a valid session_id value (16 bytes)
		if len(sessionIDCookie.Value) != 32 {
			t.Fatalf("session_id cookie is not an expected value: %s", sessionIDCookie.Value)
		}
		// needs to virtually never expire
		if sessionIDCookie.Expires.Compare(time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)) != 0 {
			t.Fatalf("session_id cookie doesn't have the correct expiration date: %s", sessionIDCookie.RawExpires)
		}
		// needs to be a samesite cookie
		if sessionIDCookie.SameSite != http.SameSiteStrictMode {
			t.Fatal("session_id cookie isn't set as samesite")
		}
	}

	request = httptest.NewRequest(http.MethodPost, "/tests?number-of-questions=50", nil)
	request.AddCookie(sessionIDCookie)
	writer = httptest.NewRecorder()
	handler.TestsHandler(writer, request)
	response = writer.Result()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("POST /tests response status code is not 201: %d", response.StatusCode)
	}
}
