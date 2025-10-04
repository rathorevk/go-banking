package api

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rathorevk/GoBanking/app/database"
	"github.com/rathorevk/GoBanking/app/database/sqlc"
	"github.com/rathorevk/GoBanking/app/helpers"
	"github.com/rathorevk/GoBanking/app/models"
)

func CreateAccount(userID int64) (sqlc.Account, error) {
	log.Println("Creating account for user ID:", userID)

	params := sqlc.CreateAccountParams{
		UserID:  userID,
		Balance: 0.0, // Starting balance
	}

	// Create account in the database
	accountCreated, err := database.DBClient.Queries.CreateAccount(context.Background(), params)
	if err != nil {
		return sqlc.Account{}, err
	}

	log.Println("Account created successfully:", accountCreated)
	return accountCreated, nil
}

func GetAccountByUser(user_id int64) (sqlc.Account, error) {
	// Use the generated SQLC method to get user
	account, err := database.DBClient.Queries.GetAccountByUser(context.Background(), user_id)
	return account, err
}

// GetBalanceHandler handles GET /user/{user_id}/balance - retrieves user balance
func GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]

	// Validate user ID using helper function
	userID, err := helpers.ValidateID(userIDStr)
	if err != nil {
		helpers.HandleAPIError(w, err)
		return
	}

	// Use the generated SQLC method to get balance
	account, err := GetAccountByUser(userID)
	if err != nil {
		helpers.HandleDatabaseError(w, err, "Account")
		return
	}

	// Create response data
	balanceStr := strconv.FormatFloat(account.Balance, 'f', 2, 64)
	responseData := models.UserBalance{
		UserID:  userID,
		Balance: balanceStr,
	}

	helpers.RespondSuccess(w, "Balance retrieved successfully", responseData)
}

// CreateAccountHandler handles POST /accounts - creates a new account
func CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var accountData models.Account

	// Validate and decode JSON request body
	if ok, validationErrors := helpers.ValidateBodyWithDetails(r, &accountData); !ok {
		helpers.RespondValidationError(w, validationErrors)
		return
	}

	// Parse user ID from string to int64
	userID, err := strconv.ParseInt(accountData.UserID, 10, 64)
	if err != nil {
		helpers.HandleAPIError(w, helpers.ErrInvalidID)
		return
	}

	// Create account
	account, err := CreateAccount(userID)
	if err != nil {
		helpers.HandleDatabaseError(w, err, "Account")
		return
	}

	helpers.RespondSuccess(w, "Account created successfully", account)
}
