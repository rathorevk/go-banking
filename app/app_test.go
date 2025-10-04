package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rathorevk/GoBanking/app/api"
	"github.com/stretchr/testify/assert"
)

func TestServerRoutes(t *testing.T) {
	// Create a test router with the same routes as StartServer
	router := mux.NewRouter()

	// Add the same routes as in StartServer
	router.HandleFunc("/user", api.CreateUserHandler).Methods("POST")
	router.HandleFunc("/user/{userId}", api.GetUserHandler).Methods("GET")
	router.HandleFunc("/user/{userId}/transaction", api.CreateTransactionHandler).Methods("POST")
	router.HandleFunc("/user/{userId}/balance", api.GetBalanceHandler).Methods("GET")

	// Test non-existent route returns 404
	req, err := http.NewRequest("GET", "/nonexistent", nil)
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	// Test wrong method returns 405
	req2, err := http.NewRequest("PUT", "/user", nil)
	assert.NoError(t, err)

	recorder2 := httptest.NewRecorder()
	router.ServeHTTP(recorder2, req2)

	assert.Equal(t, http.StatusMethodNotAllowed, recorder2.Code)
}

func TestRouterConfiguration(t *testing.T) {
	router := mux.NewRouter()

	// Test that we can create routes without errors
	assert.NotPanics(t, func() {
		router.HandleFunc("/user", api.CreateUserHandler).Methods("POST")
		router.HandleFunc("/user/{userId}", api.GetUserHandler).Methods("GET")
		router.HandleFunc("/user/{userId}/transaction", api.CreateTransactionHandler).Methods("POST")
		router.HandleFunc("/user/{userId}/balance", api.GetBalanceHandler).Methods("GET")
	})

	// Test that router is properly initialized
	assert.NotNil(t, router)
}
