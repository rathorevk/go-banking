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

func TestCreateUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Empty request body",
			requestBody:    nil,
			expectedStatus: http.StatusUnprocessableEntity,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "errors")
			},
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid-json",
			expectedStatus: http.StatusUnprocessableEntity,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				//assert.Equal(t, false, response["success"])
			},
		},
		{
			name: "Missing required fields",
			requestBody: models.User{
				Username: "", // Missing required field
				FullName: "Test User",
				Email:    "test@example.com",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				//assert.Equal(t, false, response["success"])
				assert.Contains(t, response, "errors")
			},
		},
		{
			name: "Invalid email format",
			requestBody: models.User{
				Username: "testuser",
				FullName: "Test User",
				Email:    "invalid-email",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				//assert.Equal(t, false, response["success"])
				assert.Contains(t, response, "errors")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := mux.NewRouter()
			router.HandleFunc("/user", CreateUserHandler).Methods("POST")

			var body []byte
			if tt.requestBody != nil {
				if str, ok := tt.requestBody.(string); ok {
					body = []byte(str)
				} else {
					var err error
					body, err = json.Marshal(tt.requestBody)
					assert.NoError(t, err)
				}
			}

			req, err := http.NewRequest("POST", "/user", bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, recorder)
			}
		})
	}
}

func TestGetUserHandler(t *testing.T) {
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
			router.HandleFunc("/user/{userId}", GetUserHandler).Methods("GET")

			req, err := http.NewRequest("GET", "/user/"+tt.userID, nil)
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
func TestValidateID(t *testing.T) {
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

func TestValidateBodyWithDetails(t *testing.T) {
	tests := []struct {
		name        string
		user        models.User
		expectValid bool
	}{
		{
			name: "Valid user",
			user: models.User{
				Username: "testuser",
				FullName: "Test User",
				Email:    "test@example.com",
			},
			expectValid: true,
		},
		{
			name: "Missing username",
			user: models.User{
				Username: "",
				FullName: "Test User",
				Email:    "test@example.com",
			},
			expectValid: false,
		},
		{
			name: "Missing full name",
			user: models.User{
				Username: "testuser",
				FullName: "",
				Email:    "test@example.com",
			},
			expectValid: false,
		},
		{
			name: "Invalid email",
			user: models.User{
				Username: "testuser",
				FullName: "Test User",
				Email:    "invalid-email",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock request with the user data
			body, _ := json.Marshal(tt.user)
			req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			var user models.User
			valid, errors := helpers.ValidateBodyWithDetails(req, &user)

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
func BenchmarkCreateUserHandler(b *testing.B) {
	router := mux.NewRouter()
	router.HandleFunc("/user", CreateUserHandler).Methods("POST")

	user := models.User{
		Username: "testuser",
		FullName: "Test User",
		Email:    "test@example.com",
	}
	body, _ := json.Marshal(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	}
}

func BenchmarkGetUserHandler(b *testing.B) {
	router := mux.NewRouter()
	router.HandleFunc("/user/{userId}", GetUserHandler).Methods("GET")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/user/123", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	}
}
