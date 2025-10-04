package api

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rathorevk/GoBanking/app/database"
	"github.com/rathorevk/GoBanking/app/database/sqlc"
	"github.com/rathorevk/GoBanking/app/helpers"
	"github.com/rathorevk/GoBanking/app/models"
)

// CreateTransactionHandler handles POST /user/{user_id}/transaction - creates a new transaction
func CreateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]

	// Validate user ID using helper function
	userID, err := helpers.ValidateID(userIDStr)
	if err != nil {
		helpers.HandleAPIError(w, err)
		return
	}

	// Get user account
	account, err := GetAccountByUser(userID)
	if err != nil {
		helpers.HandleDatabaseError(w, err, "Account")
		return
	}

	// Get source from header
	source := r.Header.Get("Source-Type")

	transaction := models.Transaction{
		AccountID: account.ID,
		Source:    source,
	}

	// Validate and decode JSON request body using enhanced validation
	if ok, validationErrors := helpers.ValidateBodyWithDetails(r, &transaction); !ok {
		helpers.RespondValidationError(w, validationErrors)
		return
	}

	transaction, err = validateAndParseTransactionAmount(transaction)
	if err != nil {
		helpers.HandleAPIError(w, err)
		return
	}

	// Execute transaction creation and balance update in a single database transaction
	err = runInTx(database.DBClient, func(queries *sqlc.Queries) error {
		// Create transaction within the transaction
		_, err := createTransactionInTx(queries, transaction)
		if err != nil {
			return err
		}

		// Update balance within the same transaction
		_, err = updateBalanceInTx(queries, account.ID, transaction.AmountFloat, transaction.TransactionType)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		helpers.HandleDatabaseError(w, err, "Transaction")
		return
	}

	// Return success response
	responseData := map[string]interface{}{
		"user_account_id": userID,
		"transaction_id":  transaction.ID,
		"amount":          transaction.AmountFloat,
		"type":            transaction.TransactionType,
		"source":          transaction.Source,
	}
	helpers.RespondSuccess(w, "Transaction created successfully", responseData)
}

func validateAndParseTransactionAmount(transaction models.Transaction) (models.Transaction, error) {
	// Use helper function to validate amount
	amount, err := helpers.ParseAmount(transaction.Amount)
	if err != nil {
		return models.Transaction{}, err
	}

	// Add the parsed float amount to the transaction struct
	transaction.AmountFloat = amount
	return transaction, nil
}

func runInTx(db *database.DB, fn func(queries *sqlc.Queries) error) error {
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return err
	}

	queries := db.Queries.WithTx(tx)
	err = fn(queries)
	if err == nil {
		return tx.Commit(context.Background())
	}

	rollbackErr := tx.Rollback(context.Background())
	if rollbackErr != nil {
		return errors.Join(err, rollbackErr)
	}

	return err
}

func createTransactionInTx(queries *sqlc.Queries, transaction models.Transaction) (sqlc.Transaction, error) {
	log.Println("Creating transaction in TX:", transaction)

	params := sqlc.CreateTransactionParams{
		ID:        transaction.ID,
		AccountID: transaction.AccountID,
		Amount:    transaction.AmountFloat,
		Source:    transaction.Source,
		Type:      transaction.TransactionType,
	}
	return queries.CreateTransaction(context.Background(), params)
}

func updateBalanceInTx(queries *sqlc.Queries, accountID int64, amount float64, transactionType string) (sqlc.Account, error) {
	log.Printf("Updating balance for account ID: %d, amount: %.2f, type: %s", accountID, amount, transactionType)

	// Fetch current balance
	account, err := queries.GetAccount(context.Background(), accountID)
	if err != nil {
		return sqlc.Account{}, err
	}

	currentBalance := account.Balance
	var newBalance float64

	// Calculate new balance based on transaction type
	switch transactionType {
	case "win", "deposit":
		newBalance = currentBalance + amount
	case "lose", "withdrawal":
		newBalance = currentBalance - amount
		if newBalance < 0 {
			return sqlc.Account{}, helpers.ErrInsufficientBalance
		}
	default:
		return sqlc.Account{}, helpers.ErrInvalidTransactionType
	}

	// Update the account balance
	params := sqlc.UpdateAccountParams{
		ID:      accountID,
		Balance: newBalance,
	}

	updatedAccount, err := queries.UpdateAccount(context.Background(), params)
	if err != nil {
		return sqlc.Account{}, err
	}

	log.Printf("Balance updated successfully from %.2f to %.2f", currentBalance, newBalance)
	return updatedAccount, nil
}

// GetTransaction handles GET /transactions/{transactionId} - returns specific transaction
func GetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	transactionID := vars["transactionId"]

	if transactionID == "" {
		helpers.HandleAPIError(w, helpers.ErrInvalidID)
		return
	}

	// TODO: Implement GetTransaction in SQLC
	log.Printf("Fetching transaction with ID: %s", transactionID)

	// Placeholder response
	helpers.RespondError(w, http.StatusNotImplemented, "Get transaction by ID not yet implemented")
}
