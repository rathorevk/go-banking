package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rathorevk/GoBanking/app/helpers"
	"github.com/rathorevk/GoBanking/app/models"
	"github.com/stretchr/testify/assert"
)

func TestGetBalanceHandler(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Invalid user ID format",
			userID:         "invalid",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			},
		},
		{
			name:           "Zero user ID",
			userID:         "0",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			},
		},
		{
			name:           "Negative user ID",
			userID:         "-1",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := mux.NewRouter()
			router.HandleFunc("/user/{userId}/balance", GetBalanceHandler).Methods("GET")

			url := "/user/" + tt.userID + "/balance"
			req, err := http.NewRequest("GET", url, nil)
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, recorder)
			}
		})
	}
}

// Test helper functions - These test the validation logic separately
func TestValidateIDForAccounts(t *testing.T) {
	tests := []struct {
		name        string
		idStr       string
		expectError bool
		expectedID  int64
	}{
		{
			name:        "Valid ID",
			idStr:       "123",
			expectError: false,
			expectedID:  123,
		},
		{
			name:        "Empty ID",
			idStr:       "",
			expectError: true,
			expectedID:  0,
		},
		{
			name:        "Invalid format",
			idStr:       "abc",
			expectError: true,
			expectedID:  0,
		},
		{
			name:        "Zero ID",
			idStr:       "0",
			expectError: true,
			expectedID:  0,
		},
		{
			name:        "Negative ID",
			idStr:       "-1",
			expectError: true,
			expectedID:  0,
		},
		{
			name:        "Large valid ID",
			idStr:       "999999",
			expectError: false,
			expectedID:  999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := helpers.ValidateID(tt.idStr)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, int64(0), id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestAccountValidationWithDetails(t *testing.T) {
	tests := []struct {
		name        string
		account     models.Account
		expectValid bool
	}{
		{
			name: "Valid account",
			account: models.Account{
				UserID:   "123",
				Balance:  100.0,
				Currency: "USD",
			},
			expectValid: true,
		},
		{
			name: "Missing user ID",
			account: models.Account{
				UserID:   "",
				Balance:  100.0,
				Currency: "USD",
			},
			expectValid: false,
		},
		{
			name: "Missing balance",
			account: models.Account{
				UserID:   "123",
				Currency: "USD",
			},
			expectValid: false,
		},
		{
			name: "Missing currency",
			account: models.Account{
				UserID:  "123",
				Balance: 100.0,
			},
			expectValid: false,
		},
		{
			name: "Invalid currency",
			account: models.Account{
				UserID:   "123",
				Balance:  100.0,
				Currency: "INVALID",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock request with the account data
			body, _ := json.Marshal(tt.account)
			req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			var account models.Account
			valid, errors := helpers.ValidateBodyWithDetails(req, &account)

			if tt.expectValid {
				assert.True(t, valid)
				assert.Empty(t, errors)
			} else {
				assert.False(t, valid)
				assert.NotEmpty(t, errors)
			}
		})
	}
}

// Benchmark tests
func BenchmarkGetBalanceHandler(b *testing.B) {
	router := mux.NewRouter()
	router.HandleFunc("/user/{userId}/balance", GetBalanceHandler).Methods("GET")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/user/invalid/balance", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	}
}
