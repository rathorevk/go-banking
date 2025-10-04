package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	ErrBodyCannotBeEmpty      = errors.New("request body cannot be empty")
	ErrInvalidID              = errors.New("invalid ID format")
	ErrInvalidAmount          = errors.New("invalid amount format")
	ErrAmountMustBePositive   = errors.New("amount must be a positive number")
	ErrInsufficientBalance    = errors.New("insufficient balance")
	ErrInvalidTransactionType = errors.New("invalid transaction type")
	ErrUserNotFound           = errors.New("user not found")
	ErrAccountNotFound        = errors.New("user account not found")
	ErrTransactionNotFound    = errors.New("user transaction not found")
	ErrDuplicateUser          = errors.New("user already exists")
	ErrDuplicateAccount       = errors.New("user account already exists")
)

type ValidationErrorResponse struct {
	Errors map[string]string `json:"errors"`
}

type ErrorResponse struct {
	Error string `json:"error,omitempty"`
}

// Common response functions
func RespondSuccess(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func RespondCreated(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}

func RespondError(w http.ResponseWriter, statusCode int, message string) {
	response := ErrorResponse{
		Error: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func RespondValidationError(w http.ResponseWriter, errors map[string]string) {
	response := ValidationErrorResponse{
		Errors: errors,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(response)
}

// Entity-specific validation functions
func ValidateID(userIDStr string) (int64, error) {
	if userIDStr == "" {
		return 0, ErrInvalidID
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, ErrInvalidID
	}

	if userID <= 0 {
		return 0, ErrInvalidID
	}

	return userID, nil
}

func ParseAmount(amountStr string) (float64, error) {
	if amountStr == "" {
		return 0, ErrInvalidAmount
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, ErrInvalidAmount
	}

	if amount <= 0 {
		return 0, ErrAmountMustBePositive
	}

	return float64(math.Round(amount*100) / 100), nil
}

// Error handling and response mapping
func HandleDatabaseError(w http.ResponseWriter, err error, entityType string) {
	log.Printf("Database error for %s: %v", entityType, err)

	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique"):
		RespondError(w, http.StatusConflict, fmt.Sprintf("%s already exists", entityType))
	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "no rows"):
		RespondError(w, http.StatusNotFound, fmt.Sprintf("%s not found", entityType))
	case strings.Contains(errStr, "insufficient balance"):
		RespondError(w, http.StatusBadRequest, "User balance is insufficient for this transaction")
	case strings.Contains(errStr, "foreign key") || strings.Contains(errStr, "constraint"):
		RespondError(w, http.StatusBadRequest, "Invalid reference or constraint violation")
	case strings.Contains(errStr, "connection") || strings.Contains(errStr, "timeout"):
		RespondError(w, http.StatusServiceUnavailable, "Database temporarily unavailable")
	default:
		RespondError(w, http.StatusInternalServerError, "Database operation failed")
	}
}

// Business logic error handling
func HandleAPIError(w http.ResponseWriter, err error) {
	log.Printf("API error: %v", err)
	switch err {
	case ErrUserNotFound:
		RespondError(w, http.StatusNotFound, "User not found")
	case ErrAccountNotFound:
		RespondError(w, http.StatusNotFound, "User Account not found")
	case ErrTransactionNotFound:
		RespondError(w, http.StatusNotFound, "User Transaction not found")
	case ErrInsufficientBalance:
		RespondError(w, http.StatusBadRequest, "Insufficient balance for this transaction")
	case ErrAmountMustBePositive:
		RespondError(w, http.StatusBadRequest, "Amount must be a positive number")
	case ErrInvalidAmount:
		RespondError(w, http.StatusBadRequest, "Invalid amount specified")
	case ErrInvalidTransactionType:
		RespondError(w, http.StatusBadRequest, "Invalid transaction type")
	case ErrInvalidID:
		RespondError(w, http.StatusBadRequest, "Invalid ID format")
	case ErrDuplicateUser:
		RespondError(w, http.StatusConflict, "User already exists")
	case ErrDuplicateAccount:
		RespondError(w, http.StatusConflict, "User Account already exists")
	default:
		log.Printf("Unhandled business error: %v", err)
		RespondError(w, http.StatusInternalServerError, "An unexpected error occurred")
	}
}

// Enhanced body validation with custom error messages
func ValidateBodyWithDetails(r *http.Request, reqData interface{}) (bool, map[string]string) {
	// Decode the JSON request body into the provided struct
	if err := json.NewDecoder(r.Body).Decode(reqData); err != nil {
		return false, map[string]string{"body": "Invalid JSON format: " + err.Error()}
	}

	// Check for empty body by checking if all required fields are empty
	if isEmptyStruct(reqData) {
		return false, map[string]string{"body": "Request body cannot be empty"}
	}

	// Validate the decoded data using the validator
	validate := validator.New()
	err := validate.Struct(reqData)

	if err != nil {
		// Validation syntax is invalid
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return false, map[string]string{"validation": "Invalid validation syntax"}
		}

		// Validation errors occurred
		errors := make(map[string]string)

		// Use reflection to get field information
		reflected := reflect.ValueOf(reqData)
		if reflected.Kind() == reflect.Ptr {
			reflected = reflected.Elem()
		}

		for _, validationErr := range err.(validator.ValidationErrors) {
			// Get the JSON tag name or use lowercase field name
			fieldName := getJSONFieldName(reflected.Type(), validationErr.StructField())

			// Generate user-friendly error message
			errorMessage := generateValidationErrorMessage(fieldName, validationErr)
			errors[fieldName] = errorMessage
		}

		return false, errors
	}

	return true, nil
}

// Helper function to check if a struct is effectively empty
func isEmptyStruct(v interface{}) bool {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return true
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return false
	}

	// Check if all fields are empty
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() == reflect.String && field.String() != "" {
			return false
		}
		if field.Kind() == reflect.Int64 && field.Int() != 0 {
			return false
		}
		if field.Kind() == reflect.Float64 && field.Float() != 0 {
			return false
		}
	}

	return true
}

// Helper function to get JSON field name from struct tag
func getJSONFieldName(structType reflect.Type, fieldName string) string {
	field, found := structType.FieldByName(fieldName)
	if !found {
		return strings.ToLower(fieldName)
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return strings.ToLower(fieldName)
	}

	// Handle comma-separated json tags like "json:name,omitempty"
	if commaIdx := strings.Index(jsonTag, ","); commaIdx != -1 {
		jsonTag = jsonTag[:commaIdx]
	}

	return jsonTag
}

// Helper function to generate user-friendly validation error messages
func generateValidationErrorMessage(fieldName string, err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("The %s field is required", fieldName)
	case "email":
		return fmt.Sprintf("The %s must be a valid email address", fieldName)
	case "eqfield":
		return fmt.Sprintf("The %s must be equal to %s", fieldName, err.Param())
	case "oneof":
		return fmt.Sprintf("The %s must be one of: %s", fieldName, err.Param())
	case "min":
		return fmt.Sprintf("The %s must be at least %s characters long", fieldName, err.Param())
	case "max":
		return fmt.Sprintf("The %s must be at most %s characters long", fieldName, err.Param())
	case "numeric":
		return fmt.Sprintf("The %s must be a valid number", fieldName)
	case "gte":
		return fmt.Sprintf("The %s must be greater than or equal to %s", fieldName, err.Param())
	case "lte":
		return fmt.Sprintf("The %s must be less than or equal to %s", fieldName, err.Param())
	case "len":
		return fmt.Sprintf("The %s must be exactly %s characters long", fieldName, err.Param())
	case "alpha":
		return fmt.Sprintf("The %s must contain only alphabetic characters", fieldName)
	case "alphanum":
		return fmt.Sprintf("The %s must contain only alphanumeric characters", fieldName)
	case "url":
		return fmt.Sprintf("The %s must be a valid URL", fieldName)
	case "uuid":
		return fmt.Sprintf("The %s must be a valid UUID", fieldName)
	default:
		return fmt.Sprintf("The %s field is invalid", fieldName)
	}
}

func IsValidSource(source string) bool {
	validSources := map[string]bool{
		"game":    true,
		"server":  true,
		"payment": true,
	}

	return validSources[source]
}
