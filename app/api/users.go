package api

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/rathorevk/GoBanking/app/database"
	"github.com/rathorevk/GoBanking/app/database/sqlc"
	"github.com/rathorevk/GoBanking/app/helpers"
	"github.com/rathorevk/GoBanking/app/models"
)

// Database service functions
func getUserByID(userID int64) (sqlc.User, error) {
	user, err := database.DBClient.Queries.GetUser(context.Background(), userID)
	return user, err
}

func createUserInDB(user models.User) (sqlc.User, error) {
	log.Println("Creating user:", user)

	params := sqlc.CreateUserParams{
		FullName: user.FullName,
		Email:    user.Email,
		Username: user.Username,
	}

	userCreated, err := database.DBClient.Queries.CreateUser(context.Background(), params)
	return userCreated, err
}

// HTTP Handlers
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract and validate user ID from URL
	vars := mux.Vars(r)
	userIDStr := vars["userId"]

	userID, err := helpers.ValidateID(userIDStr)
	if err != nil {
		helpers.HandleAPIError(w, err)
		return
	}

	// Fetch user from database
	user, err := getUserByID(userID)
	if err != nil {
		helpers.HandleDatabaseError(w, err, "User")
		return
	}

	// Return successful response
	helpers.RespondSuccess(w, "User retrieved successfully", user)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User

	// Validate and decode JSON request body using enhanced validation
	if ok, validationErrors := helpers.ValidateBodyWithDetails(r, &user); !ok {
		helpers.RespondValidationError(w, validationErrors)
		return
	}

	// Create user in database
	userCreated, err := createUserInDB(user)
	if err != nil {
		helpers.HandleDatabaseError(w, err, "User")
		return
	}

	// Create account for the newly created user
	_, err = CreateAccount(userCreated.ID)
	if err != nil {
		// User was created but account creation failed - this is a partial success
		helpers.RespondError(w, http.StatusInternalServerError, "User created but failed to create account")
		return
	}

	// Return successful response with both user and account data
	responseData := map[string]interface{}{
		"user": userCreated,
	}
	helpers.RespondSuccess(w, "User and account created successfully", responseData)
}
