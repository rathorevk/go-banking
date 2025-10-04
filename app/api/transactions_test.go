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

func TestCreateTransactionHandler(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		headers        map[string]string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "Invalid user ID format",
			userID: "invalid",
			requestBody: models.Transaction{
				Amount:          "100.00",
				Source:          "game",
				TransactionType: "win",
			},
			headers: map[string]string{
				"Source-Type": "game",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			},
		},
		{
			name:   "Zero user ID",
			userID: "0",
			requestBody: models.Transaction{
				Amount:          "100.00",
				Source:          "game",
				TransactionType: "win",
			},
			headers: map[string]string{
				"Source-Type": "game",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			},
		},
		{
			name:   "Negative user ID",
			userID: "-1",
			requestBody: models.Transaction{
				Amount:          "100.00",
				Source:          "game",
				TransactionType: "win",
			},
			headers: map[string]string{
				"Source-Type": "game",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			},
		},
		{
			name:           "Empty request body",
			userID:         "invalid", // Use invalid ID to avoid database calls
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest, // ID validation happens first
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			},
		},
		{
			name:           "Invalid JSON",
			userID:         "invalid", // Use invalid ID to avoid database calls
			requestBody:    "invalid-json",
			expectedStatus: http.StatusBadRequest, // ID validation happens first
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
			router.HandleFunc("/user/{userId}/transaction", CreateTransactionHandler).Methods("POST")

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

			url := "/user/" + tt.userID + "/transaction"
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Set custom headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, recorder)
			}
		})
	}
}

// Test helper functions for transactions
func TestValidateAndParseTransactionAmount(t *testing.T) {
	tests := []struct {
		name           string
		transaction    models.Transaction
		expectError    bool
		expectedAmount float64
	}{
		{
			name: "Valid amount",
			transaction: models.Transaction{
				Amount:          "100.50",
				Source:          "game",
				TransactionType: "win",
			},
			expectError:    false,
			expectedAmount: 100.50,
		},
		{
			name: "Invalid amount format",
			transaction: models.Transaction{
				Amount:          "invalid",
				Source:          "game",
				TransactionType: "win",
			},
			expectError:    true,
			expectedAmount: 0,
		},
		{
			name: "Negative amount",
			transaction: models.Transaction{
				Amount:          "-50.00",
				Source:          "game",
				TransactionType: "win",
			},
			expectError:    true,
			expectedAmount: 0,
		},
		{
			name: "Zero amount",
			transaction: models.Transaction{
				Amount:          "0.00",
				Source:          "game",
				TransactionType: "win",
			},
			expectError:    true,
			expectedAmount: 0,
		},
		{
			name: "Empty amount",
			transaction: models.Transaction{
				Amount:          "",
				Source:          "game",
				TransactionType: "win",
			},
			expectError:    true,
			expectedAmount: 0,
		},
		{
			name: "Very large amount",
			transaction: models.Transaction{
				Amount:          "999999.99",
				Source:          "game",
				TransactionType: "win",
			},
			expectError:    false,
			expectedAmount: 999999.99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateAndParseTransactionAmount(tt.transaction)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, float64(0), result.AmountFloat)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAmount, result.AmountFloat)
				assert.Equal(t, tt.transaction.TransactionType, result.TransactionType)
			}
		})
	}
}

func TestParseAmount(t *testing.T) {
	tests := []struct {
		name           string
		amountStr      string
		expectError    bool
		expectedAmount float64
	}{
		{
			name:           "Valid amount",
			amountStr:      "123.45",
			expectError:    false,
			expectedAmount: 123.45,
		},
		{
			name:           "Integer amount",
			amountStr:      "100",
			expectError:    false,
			expectedAmount: 100.0,
		},
		{
			name:        "Invalid format",
			amountStr:   "abc",
			expectError: true,
		},
		{
			name:        "Negative amount",
			amountStr:   "-50.00",
			expectError: true,
		},
		{
			name:        "Zero amount",
			amountStr:   "0",
			expectError: true,
		},
		{
			name:        "Empty string",
			amountStr:   "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := helpers.ParseAmount(tt.amountStr)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, float64(0), amount)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAmount, amount)
			}
		})
	}
}

func TestTransactionValidationWithDetails(t *testing.T) {
	tests := []struct {
		name        string
		transaction models.Transaction
		expectValid bool
	}{
		{
			name: "Valid transaction",
			transaction: models.Transaction{
				ID:              "123e4567-e89b-12d3-a456-426614174000",
				AccountID:       1,
				Amount:          "100.00",
				Source:          "game",
				TransactionType: "win",
			},
			expectValid: true,
		},
		{
			name: "Missing amount",
			transaction: models.Transaction{
				Amount:          "",
				Source:          "game",
				TransactionType: "win",
			},
			expectValid: false,
		},
		{
			name: "Missing source",
			transaction: models.Transaction{
				Amount:          "100.00",
				Source:          "",
				TransactionType: "win",
			},
			expectValid: false,
		},
		{
			name: "Missing transaction type",
			transaction: models.Transaction{
				Amount:          "100.00",
				Source:          "game",
				TransactionType: "",
			},
			expectValid: false,
		},
		{
			name: "Invalid source",
			transaction: models.Transaction{
				Amount:          "100.00",
				Source:          "invalid",
				TransactionType: "win",
			},
			expectValid: false,
		},
		{
			name: "Invalid transaction type",
			transaction: models.Transaction{
				Amount:          "100.00",
				Source:          "game",
				TransactionType: "invalid",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock request with the transaction data
			body, _ := json.Marshal(tt.transaction)
			req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			var transaction models.Transaction
			valid, errors := helpers.ValidateBodyWithDetails(req, &transaction)

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
func BenchmarkCreateTransactionHandler(b *testing.B) {
	router := mux.NewRouter()
	router.HandleFunc("/user/{userId}/transaction", CreateTransactionHandler).Methods("POST")

	transaction := models.Transaction{
		Amount:          "100.00",
		Source:          "game",
		TransactionType: "win",
	}
	body, _ := json.Marshal(transaction)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/user/invalid/transaction", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Source-Type", "game")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	}
}

func BenchmarkValidateAndParseTransactionAmount(b *testing.B) {
	transaction := models.Transaction{
		Amount:          "100.50",
		Source:          "game",
		TransactionType: "win",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateAndParseTransactionAmount(transaction)
	}
}
