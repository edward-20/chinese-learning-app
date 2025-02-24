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

func TestFirstRequestToHome(t *testing.T) {
	t.Setenv("MODE", "testing")
	t.Setenv("DB_FILE_PREFIX", "../db/chinese-learning-database")
	request := httptest.NewRequest(http.MethodGet, TEST_URL, nil)
	writer := httptest.NewRecorder()
	handler.HomeHandler(writer, request)

	response := writer.Result()
	defer response.Body.Close()

	if writer.Code != http.StatusOK {
		t.Fatalf("Didn't return correct status code")
	}

	cookies := response.Cookies()
	if len(cookies) > 1 {
		t.Fatalf("There should be only one cookie, instead there are more")
	} else if len(cookies) < 1 {
		t.Fatalf("There should be only one cookie, instead there is none")
	}

	sessionIDCookie := cookies[0]
	if sessionIDCookie.Name == "session_id" {
		// needs to have a valid session_id value (16 bytes)
		if len(sessionIDCookie.Value) != 16 {
			t.Fatalf("session_id cookie is not an expected value")
		}
		// needs to virtually never expire
		if sessionIDCookie.Expires.Compare(time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)) != 0 {
			t.Fatalf("session_id cookie doesn't have the correct expiration date")
		}
		// needs to be set to this test domain
		if sessionIDCookie.Domain != TEST_URL {
			t.Fatalf("session_id cookie isn't set to the correct domain")
		}
		// needs to be a samesite cookie
		if sessionIDCookie.SameSite != http.SameSiteStrictMode {
			t.Fatalf("session_id cookie isn't set as samesite")
		}
	}
}
