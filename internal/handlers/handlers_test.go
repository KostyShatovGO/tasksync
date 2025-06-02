package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KostyShatovGO/tasksync/pkg/db"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer dbMock.Close()

	db.DB = dbMock
	defer func() { db.DB = nil }()

	t.Run("Successful Registration", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, password, created_at FROM users WHERE username = \\$1").
			WithArgs("testuser").
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "created_at"})) // Пользователь не найден

		mock.ExpectQuery("INSERT INTO users \\(username, password, created_at\\) VALUES \\(\\$1, \\$2, \\$3\\) RETURNING id, username, created_at").
			WithArgs("testuser", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).
				AddRow(1, "testuser", time.Now()))

		payload := map[string]string{"username": "testuser", "password": "testpassword"}
		jsonPayload, _ := json.Marshal(payload)
		req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonPayload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(RegisterHandler)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if id, ok := response["id"]; !ok || id != float64(1) {
			t.Errorf("response missing or wrong id: got %v want 1", id)
		}
		if username, ok := response["username"].(string); !ok || username != "testuser" {
			t.Errorf("response wrong username: got %v want testuser", username)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Duplicate User", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "password", "created_at"}).
			AddRow(1, "testuser", "hashedpassword", time.Now())
		mock.ExpectQuery("SELECT id, username, password, created_at FROM users WHERE username = \\$1").
			WithArgs("testuser").
			WillReturnRows(rows)

		payload := map[string]string{"username": "testuser", "password": "testpassword"}
		jsonPayload, _ := json.Marshal(payload)
		req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonPayload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(RegisterHandler)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusConflict {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusConflict)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})
}

func TestLogin(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer dbMock.Close()

	db.DB = dbMock
	defer func() { db.DB = nil }()

	t.Setenv("JWT_SECRET", "testsecret")

	t.Run("Successful Login", func(t *testing.T) {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
		rows := sqlmock.NewRows([]string{"id", "username", "password", "created_at"}).
			AddRow(1, "testuser", string(hashedPassword), time.Now())
		mock.ExpectQuery("SELECT id, username, password, created_at FROM users WHERE username = \\$1").
			WithArgs("testuser").
			WillReturnRows(rows)

		payload := map[string]string{"username": "testuser", "password": "testpassword"}
		jsonPayload, _ := json.Marshal(payload)
		req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonPayload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(LoginHandler)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		var response map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if _, ok := response["token"]; !ok {
			t.Errorf("response missing token: got %v", response)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Invalid Password", func(t *testing.T) {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
		rows := sqlmock.NewRows([]string{"id", "username", "password", "created_at"}).
			AddRow(1, "testuser", string(hashedPassword), time.Now())
		mock.ExpectQuery("SELECT id, username, password, created_at FROM users WHERE username = \\$1").
			WithArgs("testuser").
			WillReturnRows(rows)

		payload := map[string]string{"username": "testuser", "password": "wrongpass"}
		jsonPayload, _ := json.Marshal(payload)
		req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonPayload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(LoginHandler)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})
}
